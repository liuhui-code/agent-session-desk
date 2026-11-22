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

// This is the actual previously released development schema, not a mock store.
func createV1(t *testing.T, path string) {
	t.Helper()
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	_, err = raw.Exec(`CREATE TABLE sessions(scope TEXT NOT NULL,id TEXT NOT NULL,title TEXT NOT NULL,command TEXT NOT NULL,PRIMARY KEY(scope,id));
 CREATE TABLE runs(scope TEXT NOT NULL,id TEXT NOT NULL,session_id TEXT NOT NULL,backend_id TEXT NOT NULL,recorded_ns INTEGER NOT NULL,stopped_ns INTEGER,PRIMARY KEY(scope,id),FOREIGN KEY(scope,session_id) REFERENCES sessions(scope,id));
 CREATE INDEX runs_recent ON runs(scope,recorded_ns DESC,id DESC);
 PRAGMA application_id=1095975985; PRAGMA user_version=1;`)
	if err != nil {
		t.Fatal(err)
	}
	v := sample()
	if _, err = raw.Exec(`INSERT INTO sessions VALUES(?,?,?,?)`, "scope-a", v.ID, v.Title, v.Command); err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`INSERT INTO runs VALUES(?,?,?,?,?,?)`, "scope-a", v.RunID, v.ID, v.BackendID, 1234, 5678); err != nil {
		t.Fatal(err)
	}
}

func TestV1UpgradePreservesIdentityAndStopEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	createV1(t, path)
	ctx := context.Background()
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.History(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Session.RunID != sample().RunID || rows[0].RecordedAt.UnixNano() != 1234 || rows[0].StopConfirmedAt == nil || rows[0].StopConfirmedAt.UnixNano() != 5678 || rows[0].PaneExit != nil {
		t.Fatalf("v1 data changed: %#v", rows)
	}
	if err = db.Save(ctx, exited()); err != nil {
		t.Fatal(err)
	}
	rows, err = db.History(ctx, 100)
	if err != nil || rows[0].PaneExit == nil {
		t.Fatalf("upgrade cannot store observations: %#v %v", rows, err)
	}
}

func TestFailedV1UpgradeDoesNotModifyExistingDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	createV1(t, path)
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec("CREATE TABLE pane_exits(unrelated TEXT)"); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(context.Background(), path, "scope-a")
	if err == nil {
		db.Close()
		t.Fatal("conflicting migration accepted")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("failed upgrade modified v1")
	}
}
