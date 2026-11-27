package store_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestClaimV7MigrationPreservesLegacyReplyWithoutMintingOwnership(t *testing.T) {
	db, path, c, msg := claimFixture(t)
	ctx := context.Background()
	db.Close()
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	legacy := session.Reply{ID: strings.Repeat("c", 32), ReplyTo: msg.ID, Source: msg.Target, Body: "legacy result", RecordedAt: time.Now().UTC()}
	data, _ := json.Marshal(legacy)
	if _, err = raw.Exec(`INSERT INTO replies VALUES(?,?,?,?,?,?)`, "a", legacy.ID, msg.ID, msg.Target.RunID, string(data), legacy.RecordedAt.UnixNano()); err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`DROP TABLE attention; DROP TABLE task_claims; PRAGMA user_version=7`); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	before, _ := os.ReadFile(path)
	check, err := store.Check(ctx, path)
	if err != nil || check.SchemaVersion != 7 {
		t.Fatal("v7 source rejected")
	}
	after, _ := os.ReadFile(path)
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("v7 check migrated source")
	}
	db, err = store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Replies(ctx, msg.ID)
	if err != nil || len(rows) != 1 || rows[0].ClaimID != "" || rows[0].ID != legacy.ID {
		t.Fatal("legacy reply lost or received invented ownership")
	}
	if status, err := db.TaskClaimStatus(ctx, msg.ID); err != nil || status != nil {
		t.Fatal("migration minted an old processing grant")
	}
	if _, err = db.ClaimTask(ctx, c, newClaimKey(t, msg.ID), 300); err == nil {
		t.Fatal("answered legacy work became runnable")
	}
	check, err = store.Check(ctx, path)
	if err != nil || check.SchemaVersion != 10 {
		t.Fatal("v8 validation failed")
	}
}

func TestClaimSnapshotPreservesUnknownAndHidesPrivateKey(t *testing.T) {
	db, path, c, msg := claimFixture(t)
	ctx := context.Background()
	key := claimForReply(t, db, c, msg.ID)
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`UPDATE task_claims SET expires_ns=?`, time.Now().Add(-time.Hour).UnixNano()); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	if _, err = db.TaskClaimStatus(ctx, msg.ID); err != nil {
		t.Fatal(err)
	}
	db.Close()
	backup := filepath.Join(t.TempDir(), "backup.sqlite3")
	if err = store.Snapshot(ctx, path, backup); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(backup)
	if bytes.Contains(data, []byte(key.Secret)) || !bytes.Contains(data, []byte(key.Digest())) {
		t.Fatal("snapshot leaked private claim or lost digest")
	}
	restored, err := store.Open(ctx, backup, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	status, err := restored.TaskClaimStatus(ctx, msg.ID)
	if err != nil || status.State != "unknown" || status.NewlyClaimed {
		t.Fatal("snapshot recovery revived uncertainty")
	}
	if _, err = restored.ClaimTask(ctx, c, newClaimKey(t, msg.ID), 300); err == nil {
		t.Fatal("restored work was reassigned")
	}
}

func TestClaimFailedMigrationPreservesV7Source(t *testing.T) {
	db, path, _, _ := claimFixture(t)
	db.Close()
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`DROP TABLE attention; DROP TABLE task_claims; CREATE TABLE task_claims(unrelated TEXT); PRAGMA user_version=7`); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	before, _ := os.ReadFile(path)
	opened, err := store.Open(context.Background(), path, "a")
	if err == nil {
		opened.Close()
		t.Fatal("conflicting task table accepted")
	}
	after, _ := os.ReadFile(path)
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("failed migration changed v7")
	}
}
