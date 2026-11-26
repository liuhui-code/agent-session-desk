package store_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func createV2(t *testing.T, path string) {
	t.Helper()
	createV1(t, path)
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	_, err = raw.Exec(`CREATE TABLE pane_exits(scope TEXT NOT NULL,run_id TEXT NOT NULL,pane_id TEXT NOT NULL,source TEXT NOT NULL,observed_ns INTEGER NOT NULL,PRIMARY KEY(scope,run_id),FOREIGN KEY(scope,run_id) REFERENCES runs(scope,id)); PRAGMA user_version=2;`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestV2UpgradeAndNamedSnapshotPreserveOriginalRunEvidence(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path, backup := filepath.Join(dir, "source.sqlite3"), filepath.Join(dir, "backup.sqlite3")
	createV2(t, path)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	check, err := store.Check(ctx, path)
	if err != nil || check.SchemaVersion != 2 {
		t.Fatalf("read-only v2: %+v %v", check, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("read-only check modified v2")
	}
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Rename(ctx, sample(), "已迁移中文"); err != nil {
		db.Close()
		t.Fatal(err)
	}
	db.Close()
	if err = store.Snapshot(ctx, path, backup); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, backup, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.History(ctx, 100)
	if err != nil || len(rows) != 1 || rows[0].Session.Title != "已迁移中文" || rows[0].Session.TitleRevision != 1 || rows[0].Session.RunID != sample().RunID || rows[0].RecordedAt.UnixNano() != 1234 || rows[0].StopConfirmedAt == nil || rows[0].StopConfirmedAt.UnixNano() != 5678 {
		t.Fatalf("snapshot lost evidence: %+v %v", rows, err)
	}
	check, err = store.Check(ctx, backup)
	if err != nil || check.SchemaVersion != 8 {
		t.Fatalf("current schema check: %+v %v", check, err)
	}
}

func TestFailedNameMigrationRollsBackWholeV1OrV2Upgrade(t *testing.T) {
	for _, version := range []string{"v1", "v2"} {
		t.Run(version, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "metadata.sqlite3")
			if version == "v1" {
				createV1(t, path)
			} else {
				createV2(t, path)
			}
			raw, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = raw.Exec("CREATE TABLE session_names(unrelated TEXT)")
			raw.Close()
			if err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			db, err := store.Open(context.Background(), path, "scope-a")
			if err == nil {
				db.Close()
				t.Fatal("conflicting names table accepted")
			}
			after, err := os.ReadFile(path)
			if err != nil || sha256.Sum256(before) != sha256.Sum256(after) {
				t.Fatal("failed migration changed source")
			}
		})
	}
}
