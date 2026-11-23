package manager_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestWorkspaceGroupMovePersistsWithoutTerminalMutation(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	f := &terminal{items: []session.Session{observed()}}
	m := manager.New(f, db)
	view, err := m.Workspace(ctx)
	if err != nil || len(view.Groups) != 1 || len(view.Sessions) != 1 {
		t.Fatalf("workspace: %+v %v", view, err)
	}
	g, err := m.CreateGroup(ctx, "", "项目", `C:\work`)
	if err != nil {
		t.Fatal(err)
	}
	named, err := m.Rename(ctx, view.Sessions[0], "评审")
	if err != nil {
		t.Fatal(err)
	}
	moved, err := m.MoveToGroup(ctx, view.Sessions[0], g.ID)
	if err != nil || moved.Title != named.Title || moved.TitleRevision != named.TitleRevision || moved.GroupID != g.ID {
		t.Fatalf("move lost independent name update: %+v %v", moved, err)
	}
	if _, err = m.MoveToGroup(ctx, view.Sessions[0], session.DefaultGroupID); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("stale membership accepted: %v", err)
	}
	if f.starts != 0 || f.stops != 0 || f.items[0] != observed() {
		t.Fatal("group changed original terminal")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m = manager.New(f, db)
	items, err := m.List(ctx)
	if err != nil || len(items) != 1 || items[0].GroupID != g.ID || items[0].Title != named.Title {
		t.Fatalf("reopen lost group: %+v %v", items, err)
	}
	f.items = nil
	if _, err = m.MoveToGroup(ctx, items[0], session.DefaultGroupID); err == nil {
		t.Fatal("absent Run moved")
	}
	history, err := db.History(ctx, 10)
	if err != nil || history[0].Session.GroupID != g.ID {
		t.Fatal("absent Run changed membership")
	}
}

func TestGroupedLaunchValidatesBeforeStartAndCommitsRunWithMembership(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := &terminal{items: []session.Session{observed()}}
	m := manager.New(f, db)
	if _, err = m.Start(ctx, session.Launch{GroupID: "missing"}); err == nil || f.starts != 0 {
		t.Fatal("created terminal for missing group")
	}
	g, err := m.CreateGroup(ctx, "", "project", `C:\work`)
	if err != nil {
		t.Fatal(err)
	}
	v, err := m.Start(ctx, session.Launch{GroupID: g.ID})
	if err != nil || v.GroupID != g.ID || v.GroupRevision != 1 {
		t.Fatalf("grouped start: %+v %v", v, err)
	}
	view, err := m.Workspace(ctx)
	if err != nil || view.Sessions[0].GroupID != g.ID || view.Sessions[0].RunID != v.RunID || f.starts != 1 {
		t.Fatalf("grouped start not durable: %+v %v", view, err)
	}
}

func TestGroupedLaunchMetadataFailureRollsBackRunAndNeverRetriesOrStopsTheTerminal(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := &terminal{items: []session.Session{observed()}}
	m := manager.New(f, db)
	g, err := m.CreateGroup(ctx, "", "project", "")
	if err != nil {
		t.Fatal(err)
	}
	f.afterStart = func() {
		raw, e := sql.Open("sqlite", path)
		if e != nil {
			t.Fatal(e)
		}
		defer raw.Close()
		if _, e = raw.Exec(`CREATE TRIGGER reject_membership BEFORE INSERT ON session_groups BEGIN SELECT RAISE(ABORT,'synthetic membership failure'); END`); e != nil {
			t.Fatal(e)
		}
	}
	v, err := m.Start(ctx, session.Launch{GroupID: g.ID})
	if err == nil || !strings.Contains(err.Error(), "do not retry start") || v.ID != observed().ID || v.RunID != observed().RunID || f.starts != 1 || f.stops != 0 {
		t.Fatalf("ambiguous external result hidden or retried: %+v %v", v, err)
	}
	history, e := db.History(ctx, 10)
	if e != nil || len(history) != 0 {
		t.Fatalf("failed membership left a half-committed Run: %+v %v", history, e)
	}
}
