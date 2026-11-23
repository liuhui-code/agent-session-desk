package store_test

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestGroupRenameRejectsStaleABAAndSiblingCollision(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	other, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	root, _ := db.CreateGroup(ctx, "", "first", "")
	sibling, _ := db.CreateGroup(ctx, "", "second", "")
	child, _ := db.CreateGroup(ctx, root.ID, "second", "")
	if _, err = other.RenameGroup(ctx, root, "second"); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("same-parent collision accepted: %v", err)
	}
	named, err := other.RenameGroup(ctx, root, "third")
	if err != nil {
		t.Fatal(err)
	}
	reverted, err := other.RenameGroup(ctx, named, root.Name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.RenameGroup(ctx, root, "stale"); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("ABA editor accepted: %v", err)
	}
	// A child's same name under a different parent is independent.
	updated, err := db.RenameGroup(ctx, reverted, "second-child")
	if err != nil {
		t.Fatal(err)
	}
	groups, err := db.Groups(ctx)
	if err != nil || len(groups) != 4 || groups[1] != updated || groups[2] != child || groups[3] != sibling {
		t.Fatalf("collision/stale edit changed tree: %+v %v", groups, err)
	}
	unchanged, err := db.RenameGroup(ctx, updated, "second/child")
	if err != nil || unchanged != updated {
		t.Fatalf("cleaned no-op changed version: %+v %v", unchanged, err)
	}
}

func TestGroupRenameScopesDefaultAndTwoConcurrentEditors(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	other, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	foreign, err := store.Open(ctx, path, "other-scope")
	if err != nil {
		t.Fatal(err)
	}
	defer foreign.Close()
	groups, err := db.Groups(ctx)
	if err != nil {
		t.Fatal(err)
	}
	def := groups[0]
	start := make(chan struct{})
	results := make(chan error, 2)
	for i, connection := range []*store.Store{db, other} {
		go func(i int, connection *store.Store) {
			<-start
			_, e := connection.RenameGroup(ctx, def, []string{"甲", "乙"}[i])
			results <- e
		}(i, connection)
	}
	close(start)
	var success, conflicts int
	for range 2 {
		e := <-results
		if e == nil {
			success++
		} else if errors.Is(e, store.ErrGroupConflict) {
			conflicts++
		} else {
			t.Fatal(e)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("concurrent edits: %d success, %d conflicts", success, conflicts)
	}
	groups, err = foreign.Groups(ctx)
	if err != nil || len(groups) != 1 || groups[0] != def {
		t.Fatalf("edited another scope: %+v %v", groups, err)
	}
	custom, err := db.CreateGroup(ctx, "", "custom", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = foreign.RenameGroup(ctx, custom, "foreign edit"); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("foreign identity accepted: %v", err)
	}
}

func TestGroupRenameRejectsInvalidInputAndRevisionExhaustionWithoutChangingMetadata(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	g, err := db.CreateGroup(ctx, "", "original", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", " ", "line\nnext", "nul\x00", "\xff", "row\u2028next"} {
		if _, err = db.RenameGroup(ctx, g, name); err == nil {
			t.Fatalf("invalid name accepted: %q", name)
		}
	}
	forged := g
	forged.DefaultDir = "changed independently"
	if _, err = db.RenameGroup(ctx, forged, "new"); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("changed snapshot accepted: %v", err)
	}
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if _, err = raw.Exec(`UPDATE groups SET revision=? WHERE scope=? AND id=?`, int64(math.MaxInt64), "scope", g.ID); err != nil {
		t.Fatal(err)
	}
	g.Revision = math.MaxInt64
	if _, err = db.RenameGroup(ctx, g, "new"); err == nil {
		t.Fatal("revision overflow accepted")
	}
	unchanged, err := db.RenameGroup(ctx, g, g.Name)
	if err != nil || unchanged != g {
		t.Fatalf("no-op exhausted version: %+v %v", unchanged, err)
	}
	groups, err := db.Groups(ctx)
	if err != nil || groups[1] != g {
		t.Fatalf("failed edit changed metadata: %+v %v", groups, err)
	}
}

func TestGroupRenamePersistsWithoutChangingDescendantsMembershipOrRun(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	if err = db.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	root, err := db.CreateGroup(ctx, "", "项目", `C:\work`)
	if err != nil {
		t.Fatal(err)
	}
	child, err := db.CreateGroup(ctx, root.ID, "评审", "")
	if err != nil {
		t.Fatal(err)
	}
	moved, err := db.MoveToGroup(ctx, sample(), child.ID)
	if err != nil {
		t.Fatal(err)
	}
	named, err := db.RenameGroup(ctx, root, "项目/新版")
	if err != nil {
		t.Fatal(err)
	}
	want := root
	want.Name, want.Revision = "项目-新版", root.Revision+1
	if named != want {
		t.Fatalf("rename changed group identity/configuration: %+v", named)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	view, err := db.Workspace(ctx, []session.Session{sample()})
	if err != nil || len(view.Groups) != 3 || view.Groups[1] != named || view.Groups[2] != child || len(view.Sessions) != 1 || view.Sessions[0] != moved {
		t.Fatalf("rename lost original tree/member/Run after reopen: %+v %v", view, err)
	}
	history, err := db.History(ctx, 10)
	if err != nil || len(history) != 1 || history[0].Session.GroupID != child.ID || history[0].Session.RunID != moved.RunID {
		t.Fatalf("rename rewrote historical Run: %+v %v", history, err)
	}
}
