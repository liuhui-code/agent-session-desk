package store_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/store"
)

// Older fixtures built from the current schema must remove all newer objects,
// not merely change user_version and accidentally test a hybrid database.
func removeRecoveryV11(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`DROP TABLE recovery_attempts; DROP TABLE recovery_profiles; DROP TABLE run_launches; ALTER TABLE runs DROP COLUMN launch_title; ALTER TABLE runs DROP COLUMN launch_command;`); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryV10MigrationBackupAndRunSnapshot(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "db")
	db, e := store.Open(ctx, path, "a")
	if e != nil {
		t.Fatal(e)
	}
	v := sample()
	if e = db.Save(ctx, v); e != nil {
		t.Fatal(e)
	}
	db.Close()
	raw, e := sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	removeRecoveryV11(t, raw)
	if _, e = raw.Exec(`PRAGMA user_version=10`); e != nil {
		t.Fatal(e)
	}
	raw.Close()
	if check, e := store.Check(ctx, path); e != nil || check.SchemaVersion != 10 {
		t.Fatal("v10 check", check, e)
	}
	backup := filepath.Join(t.TempDir(), "backup")
	if e = store.Snapshot(ctx, path, backup); e != nil {
		t.Fatal(e)
	}
	db, e = store.Open(ctx, backup, "a")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = db.Save(ctx, v); e != nil {
		t.Fatal("snapshot migrated differently", e)
	}
	if check, e := store.Check(ctx, backup); e != nil || check.SchemaVersion != 11 {
		t.Fatal("v11 check", check, e)
	}
}
