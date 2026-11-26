package store_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
	"os"
	"path/filepath"
	"testing"
)

// Frozen v4 tables; not generated from the v5 initializer under test.
const frozenV4Groups = `CREATE TABLE groups(scope TEXT NOT NULL,id TEXT NOT NULL,parent_id TEXT,name TEXT NOT NULL CHECK(length(name) BETWEEN 1 AND 50 AND length(CAST(name AS BLOB))<=200),default_dir TEXT NOT NULL CHECK(length(CAST(default_dir AS BLOB))<=4096),revision INTEGER NOT NULL CHECK(revision>0),ordinal INTEGER NOT NULL CHECK(ordinal>=0),PRIMARY KEY(scope,id),FOREIGN KEY(scope,parent_id) REFERENCES groups(scope,id),CHECK(id='default' OR (length(id)=32 AND id NOT GLOB '*[^0-9a-f]*')),CHECK(id<>'default' OR (parent_id IS NULL AND ordinal=0)));CREATE UNIQUE INDEX groups_sibling_name ON groups(scope,coalesce(parent_id,''),name);CREATE TABLE session_groups(scope TEXT NOT NULL,session_id TEXT NOT NULL,group_id TEXT NOT NULL,revision INTEGER NOT NULL CHECK(revision>0),PRIMARY KEY(scope,session_id),FOREIGN KEY(scope,session_id) REFERENCES sessions(scope,id),FOREIGN KEY(scope,group_id) REFERENCES groups(scope,id));INSERT INTO groups SELECT DISTINCT scope,'default',NULL,'我的会话','',1,0 FROM sessions;PRAGMA user_version=4;`

func createV4(t *testing.T, path string) {
	t.Helper()
	createV3(t, path)
	db, e := sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if _, e = db.Exec(frozenV4Groups); e != nil {
		t.Fatal(e)
	}
}
func TestSessionOrderV4ReadOnlyMigrationAndV5Snapshot(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "v4.sqlite3")
	createV4(t, path)
	before, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	check, e := store.Check(ctx, path)
	if e != nil || check.SchemaVersion != 4 {
		t.Fatalf("v4 read-only check: %+v %v", check, e)
	}
	after, _ := os.ReadFile(path)
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("read-only v4 check mutated source")
	}
	db, e := store.Open(ctx, path, "scope-a")
	if e != nil {
		t.Fatal(e)
	}
	check, e = store.Check(ctx, path)
	if e != nil || check.SchemaVersion != 6 {
		t.Fatalf("v5 migration: %+v %v", check, e)
	}
	original, e := db.History(ctx, 10)
	if e != nil {
		t.Fatal(e)
	}
	rows := []session.Session{sample(), sample()}
	rows[1].ID = "cccccccccccccccccccccccccccccccc"
	rows[1].RunID = "dddddddddddddddddddddddddddddddd"
	rows[1].BackendID = "new"
	rows[1].Title = "new"
	if e = db.Save(ctx, rows[1]); e != nil {
		t.Fatal(e)
	}
	// Original v4 Run is already stopped; use its preserved history to prove migration,
	// and a new pair of live rows for manual positions.
	rows[0] = rows[1]
	rows[0].ID = "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	rows[0].RunID = "ffffffffffffffffffffffffffffffff"
	rows[0].BackendID = "other"
	rows[0].Title = "aaa"
	if e = db.Save(ctx, rows[0]); e != nil {
		t.Fatal(e)
	}
	view, e := db.Workspace(ctx, rows)
	if e != nil {
		t.Fatal(e)
	}
	view, e = db.ReorderSession(ctx, session.SessionOrder{TargetID: rows[1].ID, Siblings: view.Sessions, Stamp: view.OrderStamps[session.DefaultGroupID]}, -1, rows)
	if e != nil {
		t.Fatal(e)
	}
	db.Close()
	backup, restored := filepath.Join(dir, "backup.sqlite3"), filepath.Join(dir, "restored.sqlite3")
	if e = store.Snapshot(ctx, path, backup); e != nil {
		t.Fatal(e)
	}
	if e = store.Snapshot(ctx, backup, restored); e != nil {
		t.Fatal(e)
	}
	db, e = store.Open(ctx, restored, "scope-a")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	got, e := db.Workspace(ctx, rows)
	if e != nil || !session.SameOrderSnapshot(view.Sessions, got.Sessions) {
		t.Fatalf("restore lost positions: %+v %v", got, e)
	}
	history, e := db.History(ctx, 10)
	if e != nil || len(history) != 3 {
		t.Fatalf("history lost: %+v %v", history, e)
	}
	var found bool
	for _, r := range history {
		if r.Session.ID == sample().ID {
			found = r.Session.Title == original[0].Session.Title && r.RecordedAt == original[0].RecordedAt && r.StopConfirmedAt.Equal(*original[0].StopConfirmedAt)
		}
	}
	if !found {
		t.Fatal("migration/restore rewrote original historical Run")
	}
}
func TestSessionOrderMigrationConflictRollsBackAndCorruptPositionIsRejected(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "v4.sqlite3")
	createV4(t, path)
	raw, e := sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = raw.Exec(`CREATE TABLE session_positions(sentinel TEXT);INSERT INTO session_positions VALUES('preserve')`); e != nil {
		t.Fatal(e)
	}
	raw.Close()
	before, _ := os.ReadFile(path)
	if db, e := store.Open(ctx, path, "scope-a"); e == nil {
		db.Close()
		t.Fatal("v4 extension collision accepted")
	}
	after, _ := os.ReadFile(path)
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("failed migration mutated v4 source")
	}
	db, rows, path := orderFixture(t)
	if _, e = db.ReorderSession(ctx, orderSnapshot(t, db, rows, rows[1].ID), -1, rows); e != nil {
		t.Fatal(e)
	}
	g, e := db.CreateGroup(ctx, "", "other", "")
	if e != nil {
		t.Fatal(e)
	}
	raw, e = sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	defer raw.Close()
	if _, e = raw.Exec(`UPDATE session_positions SET group_id=? WHERE scope='scope'`, g.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = store.Check(ctx, path); e == nil {
		t.Fatal("read-only check accepted inconsistent position group")
	}
	if _, e = db.Workspace(ctx, rows); e == nil {
		t.Fatal("workspace accepted inconsistent position group")
	}
}
