package store_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestMessageSchemaV5UpgradeBackupAndCapacity(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, e := store.Open(ctx, path, "a")
	if e != nil {
		t.Fatal(e)
	}
	target := sample()
	if e = db.Save(ctx, target); e != nil {
		t.Fatal(e)
	}
	db.Close()
	raw, e := sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = raw.Exec(`DROP TABLE attention; DROP TABLE task_claims; DROP TABLE replies; DROP TABLE agent_credentials; DROP TABLE messages; PRAGMA user_version=5`); e != nil {
		t.Fatal(e)
	}
	raw.Close()
	before, e := store.Check(ctx, path)
	if e != nil || before.SchemaVersion != 5 {
		t.Fatalf("v5 source: %+v %v", before, e)
	}
	db, e = store.Open(ctx, path, "a")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	for i := 0; i < store.MaxMessages; i++ {
		_, e = db.QueueMessage(ctx, session.Message{ID: fmt.Sprintf("%032x", i+1), Target: target, Body: "bounded", Fingerprint: strings.Repeat("a", 64)})
		if e != nil {
			t.Fatal(e)
		}
	}
	if _, e = db.QueueMessage(ctx, session.Message{ID: strings.Repeat("f", 32), Target: target, Body: "overflow", Fingerprint: strings.Repeat("b", 64)}); e == nil {
		t.Fatal("capacity silently exceeded or receipts pruned")
	}
	// Capacity must not stop idempotent lookup of an existing request.
	if _, e = db.QueueMessage(ctx, session.Message{ID: fmt.Sprintf("%032x", 1), Target: target, Body: "bounded", Fingerprint: strings.Repeat("a", 64)}); e != nil {
		t.Fatal(e)
	}
	check, e := store.Check(ctx, path)
	if e != nil || check.SchemaVersion != 9 {
		t.Fatalf("v6 integrity: %+v %v", check, e)
	}
	backup := filepath.Join(t.TempDir(), "backup.sqlite3")
	if e = store.Snapshot(ctx, path, backup); e != nil {
		t.Fatal(e)
	}
	copy, e := store.Open(ctx, backup, "a")
	if e != nil {
		t.Fatal(e)
	}
	defer copy.Close()
	got, e := copy.Message(ctx, fmt.Sprintf("%032x", 1))
	if e != nil || got.Body != "bounded" || got.State != "queued" {
		t.Fatalf("message backup: %+v %v", got, e)
	}
}
