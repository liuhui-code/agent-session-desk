package store_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func createV3(t *testing.T, path string) {
	t.Helper()
	createV2(t, path)
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	_, err = raw.Exec(`CREATE TABLE session_names(scope TEXT NOT NULL,session_id TEXT NOT NULL,title TEXT NOT NULL CHECK(length(title) BETWEEN 1 AND 128 AND length(CAST(title AS BLOB))<=512),revision INTEGER NOT NULL CHECK(revision>0),PRIMARY KEY(scope,session_id),FOREIGN KEY(scope,session_id) REFERENCES sessions(scope,id)); PRAGMA user_version=3;`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`INSERT INTO session_names VALUES(?,?,?,?)`, "scope-a", sample().ID, "迁移前名称", 2); err != nil {
		t.Fatal(err)
	}
}

func TestV3GroupMigrationAndSnapshotPreserveOriginalEvidence(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	source, backup, restored := filepath.Join(dir, "source.sqlite3"), filepath.Join(dir, "backup.sqlite3"), filepath.Join(dir, "restored.sqlite3")
	createV3(t, source)
	before, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	check, err := store.Check(ctx, source)
	if err != nil || check.SchemaVersion != 3 {
		t.Fatalf("v3 check: %+v %v", check, err)
	}
	after, err := os.ReadFile(source)
	if err != nil || sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("read-only check changed v3")
	}
	db, err := store.Open(ctx, source, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	g, err := db.CreateGroup(ctx, session.DefaultGroupID, "评审", `C:\work`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.MoveToGroup(ctx, sample(), g.ID); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	if err = store.Snapshot(ctx, source, backup); err != nil {
		t.Fatal(err)
	}
	if err = store.Snapshot(ctx, backup, restored); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, restored, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	view, err := db.Workspace(ctx, []session.Session{sample()})
	if err != nil || len(view.Groups) != 2 || view.Groups[1].ID != g.ID || view.Groups[1].DefaultDir != g.DefaultDir || view.Sessions[0].GroupID != g.ID || view.Sessions[0].Title != "迁移前名称" || view.Sessions[0].TitleRevision != 2 {
		t.Fatalf("lost restored workspace: %+v %v", view, err)
	}
	rows, err := db.History(ctx, 10)
	if err != nil || len(rows) != 1 || rows[0].Session.RunID != sample().RunID || rows[0].Session.GroupID != g.ID || rows[0].RecordedAt.UnixNano() != 1234 || rows[0].StopConfirmedAt == nil || rows[0].StopConfirmedAt.UnixNano() != 5678 {
		t.Fatalf("lost original Run evidence: %+v %v", rows, err)
	}
	check, err = store.Check(ctx, restored)
	if err != nil || check.SchemaVersion != 8 {
		t.Fatalf("current schema check: %+v %v", check, err)
	}
}
