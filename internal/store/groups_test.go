package store_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestGroupsSurviveReopenAndMoveKeepsOriginalRunAndHistory(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	root, err := db.CreateGroup(ctx, "", "项目 甲", "")
	if err != nil {
		t.Fatal(err)
	}
	child, err := db.CreateGroup(ctx, root.ID, "Review", `C:\work\project`)
	if err != nil {
		t.Fatal(err)
	}
	moved, err := db.MoveToGroup(ctx, sample(), child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if moved.ID != sample().ID || moved.RunID != sample().RunID || moved.BackendID != sample().BackendID || moved.Command != sample().Command || moved.GroupID != child.ID || moved.GroupRevision != 1 {
		t.Fatalf("move changed the original target: %+v", moved)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	view, err := db.Workspace(ctx, []session.Session{sample()})
	if err != nil || len(view.Groups) != 3 || len(view.Sessions) != 1 || view.Sessions[0].GroupID != child.ID || view.Sessions[0].GroupRevision != 1 {
		t.Fatalf("lost tree/membership: %+v %v", view, err)
	}
	if view.Groups[0].ID != session.DefaultGroupID || view.Groups[1].ID != root.ID || view.Groups[2].ParentID != root.ID || view.Groups[2].DefaultDir != `C:\work\project` {
		t.Fatalf("lost parent/default directory: %+v", view.Groups)
	}
	history, err := db.History(ctx, 10)
	if err != nil || len(history) != 1 || history[0].Session.RunID != sample().RunID || history[0].Session.GroupID != child.ID || history[0].Session.TerminalState != session.TerminalUnknown {
		t.Fatalf("history changed Run or lost current logical group: %+v %v", history, err)
	}
}

func TestWorkspaceReadDoesNotTakeTheGroupWriterLock(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if _, err = raw.Exec("BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	defer raw.Exec("ROLLBACK")
	view, err := db.Workspace(ctx, []session.Session{sample()})
	if err != nil || len(view.Groups) != 1 || len(view.Sessions) != 1 {
		t.Fatalf("read-only snapshot required a writer lock: %+v %v", view, err)
	}
}
