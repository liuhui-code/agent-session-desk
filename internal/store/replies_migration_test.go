package store_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestReplySchemaV6MigrationRecoveryAndNoPlaintextCredential(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	path := filepath.Join(root, "metadata.sqlite3")
	db, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	v := sample()
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	msg := session.Message{ID: strings.Repeat("a", 32), Target: v, Body: "request", Fingerprint: strings.Repeat("b", 64)}
	if _, err = db.QueueMessage(ctx, msg); err != nil {
		t.Fatal(err)
	}
	db.Close()
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`DROP TABLE replies; DROP TABLE agent_credentials; PRAGMA user_version=6`); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	check, err := store.Check(ctx, path)
	if err != nil || check.SchemaVersion != 6 {
		t.Fatalf("v6 source: %+v %v", check, err)
	}
	after, _ := os.ReadFile(path)
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("read-only v6 check migrated source")
	}
	db, err = store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	c, err := session.NewAgentCredential(v.ID, v.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.BindAgent(ctx, c, v); err != nil {
		t.Fatal(err)
	}
	r := session.ReplyRequest{ID: strings.Repeat("c", 32), ReplyTo: msg.ID, Body: "result 中文"}
	first, err := db.RecordReply(ctx, c, r)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	backup := filepath.Join(root, "backup.sqlite3")
	if err = store.Snapshot(ctx, path, backup); err != nil {
		t.Fatal(err)
	}
	check, err = store.Check(ctx, backup)
	if err != nil || check.SchemaVersion != 7 {
		t.Fatalf("v7 snapshot %+v %v", check, err)
	}
	data, _ := os.ReadFile(backup)
	if bytes.Contains(data, []byte(c.Secret)) || !bytes.Contains(data, []byte(c.Digest())) {
		t.Fatal("metadata stored a plaintext token or lost its digest")
	}
	db, err = store.Open(ctx, backup, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	again, err := db.RecordReply(ctx, c, r)
	if err != nil || !again.RecordedAt.Equal(first.RecordedAt) {
		t.Fatal("restored reply retry duplicated or lost original")
	}
	rows, err := db.Replies(ctx, msg.ID)
	if err != nil || len(rows) != 1 {
		t.Fatal("result lost after metadata restore")
	}
}

func TestReplyFailedV6MigrationRollsBackBindingTable(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`DROP TABLE replies; DROP TABLE agent_credentials; CREATE TABLE replies(unrelated TEXT); PRAGMA user_version=6`); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	before, _ := os.ReadFile(path)
	db, err = store.Open(ctx, path, "a")
	if err == nil {
		db.Close()
		t.Fatal("conflicting reply table accepted")
	}
	after, _ := os.ReadFile(path)
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("failed migration changed v6 database")
	}
}
