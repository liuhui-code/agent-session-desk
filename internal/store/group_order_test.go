package store_test

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"path/filepath"
	"slices"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func groupOrderSnapshot(t *testing.T, db *store.Store, id string) session.GroupOrder {
	t.Helper()
	groups, err := db.Groups(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var parent string
	found := false
	for _, g := range groups {
		if g.ID == id {
			parent = g.ParentID
			found = true
		}
	}
	if !found {
		t.Fatal("test group missing")
	}
	order := session.GroupOrder{TargetID: id}
	for _, g := range groups {
		if g.ParentID == parent {
			order.Siblings = append(order.Siblings, g)
		}
	}
	return order
}

func TestGroupOrderChildSwapKeepsOtherParentsAndRootOrder(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	root, err := db.CreateGroup(ctx, "", "root", "")
	if err != nil {
		t.Fatal(err)
	}
	first, err := db.CreateGroup(ctx, root.ID, "first", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.CreateGroup(ctx, root.ID, "second", "")
	if err != nil {
		t.Fatal(err)
	}
	other, err := db.CreateGroup(ctx, "", "other", "")
	if err != nil {
		t.Fatal(err)
	}
	foreignChild, err := db.CreateGroup(ctx, other.ID, "foreign", "")
	if err != nil {
		t.Fatal(err)
	}
	groups, err := db.ReorderGroup(ctx, groupOrderSnapshot(t, db, second.ID), -1)
	if err != nil || len(groups) != 6 || groups[1] != root || groups[2].ID != second.ID || groups[3].ID != first.ID || groups[4] != other || groups[5] != foreignChild {
		t.Fatalf("child swap crossed parent/subtree boundary: %+v %v", groups, err)
	}
	if groups[2].ParentID != root.ID || groups[3].ParentID != root.ID {
		t.Fatal("child sort reparents groups")
	}
}

func TestGroupOrderRejectsStaleSiblingEditsInsertionsAndABA(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	first, err := db.CreateGroup(ctx, "", "first", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.CreateGroup(ctx, "", "second", "")
	if err != nil {
		t.Fatal(err)
	}
	original := groupOrderSnapshot(t, db, second.ID)
	if _, err = db.ReorderGroup(ctx, original, -1); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ReorderGroup(ctx, original, -1); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("stale order accepted: %v", err)
	}
	if _, err = db.ReorderGroup(ctx, groupOrderSnapshot(t, db, second.ID), 1); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ReorderGroup(ctx, original, -1); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("ABA order accepted: %v", err)
	}
	current := groupOrderSnapshot(t, db, second.ID)
	var currentFirst session.Group
	for _, g := range current.Siblings {
		if g.ID == first.ID {
			currentFirst = g
		}
	}
	if _, err = db.RenameGroup(ctx, currentFirst, "first renamed"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ReorderGroup(ctx, current, -1); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("changed neighbor accepted: %v", err)
	}
	current = groupOrderSnapshot(t, db, second.ID)
	if _, err = db.CreateGroup(ctx, "", "inserted", ""); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ReorderGroup(ctx, current, -1); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("changed sibling set accepted: %v", err)
	}
	current = groupOrderSnapshot(t, db, second.ID)
	child, err := db.CreateGroup(ctx, first.ID, "independent child", "")
	if err != nil {
		t.Fatal(err)
	}
	ordered, err := db.ReorderGroup(ctx, current, -1)
	if err != nil {
		t.Fatalf("unrelated child invalidated root siblings: %v", err)
	}
	if !slices.Contains(ordered, child) {
		t.Fatal("root reorder lost newly observed child")
	}
}

func TestGroupOrderConcurrentConnectionsOnlyOneCommits(t *testing.T) {
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
	for _, name := range []string{"first", "second", "third"} {
		if _, err = db.CreateGroup(ctx, "", name, ""); err != nil {
			t.Fatal(err)
		}
	}
	groups, err := db.Groups(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expected := session.GroupOrder{TargetID: groups[2].ID, Siblings: groups}
	start := make(chan struct{})
	results := make(chan error, 2)
	for i, connection := range []*store.Store{db, other} {
		go func(i int, connection *store.Store) {
			<-start
			_, e := connection.ReorderGroup(ctx, expected, []int{-1, 1}[i])
			results <- e
		}(i, connection)
	}
	close(start)
	success, conflicts := 0, 0
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
		t.Fatalf("concurrent edits: %d successes %d conflicts", success, conflicts)
	}
}

func TestGroupOrderPinnedBoundariesAndMalformedSnapshotsDoNotWrite(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	first, err := db.CreateGroup(ctx, "", "first", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.CreateGroup(ctx, "", "second", "")
	if err != nil {
		t.Fatal(err)
	}
	before, err := db.Groups(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		id    string
		delta int
	}{{session.DefaultGroupID, -1}, {session.DefaultGroupID, 1}, {first.ID, -1}, {second.ID, 1}} {
		got, e := db.ReorderGroup(ctx, groupOrderSnapshot(t, db, test.id), test.delta)
		if e != nil || !slices.Equal(got, before) {
			t.Fatalf("boundary/pin changed data: %+v %v", got, e)
		}
	}
	expected := groupOrderSnapshot(t, db, second.ID)
	for _, delta := range []int{-2, 0, 2} {
		if _, err = db.ReorderGroup(ctx, expected, delta); err == nil {
			t.Fatalf("invalid direction %d accepted", delta)
		}
	}
	missing := expected
	missing.Siblings = expected.Siblings[1:]
	if _, err = db.ReorderGroup(ctx, missing, -1); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("partial sibling snapshot accepted: %v", err)
	}
	foreign, err := store.Open(ctx, path, "other-scope")
	if err != nil {
		t.Fatal(err)
	}
	defer foreign.Close()
	if _, err = foreign.ReorderGroup(ctx, expected, -1); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("foreign group accepted: %v", err)
	}
	after, err := db.Groups(ctx)
	if err != nil || !slices.Equal(before, after) {
		t.Fatalf("rejected/no-op order changed stored versions: %+v %v", after, err)
	}
}

func TestGroupOrderRollsBackAllWritesOnFailureAndRevisionExhaustion(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	first, err := db.CreateGroup(ctx, "", "first", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.CreateGroup(ctx, "", "second", "")
	if err != nil {
		t.Fatal(err)
	}
	before, err := db.Groups(ctx)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	// Reject the second update after the first sibling was already changed.
	if _, err = raw.Exec(`CREATE TRIGGER reject_order BEFORE UPDATE OF ordinal ON groups WHEN OLD.id='` + second.ID + `' BEGIN SELECT RAISE(ABORT,'synthetic second order update failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ReorderGroup(ctx, session.GroupOrder{TargetID: second.ID, Siblings: before}, -1); err == nil {
		t.Fatal("synthetic second write failure ignored")
	}
	after, err := db.Groups(ctx)
	if err != nil || !slices.Equal(before, after) {
		t.Fatalf("failure left a half-committed swap: %+v %v", after, err)
	}
	if _, err = raw.Exec(`DROP TRIGGER reject_order`); err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`UPDATE groups SET revision=? WHERE scope=? AND id=?`, int64(math.MaxInt64), "scope", first.ID); err != nil {
		t.Fatal(err)
	}
	exhausted := groupOrderSnapshot(t, db, second.ID)
	if _, err = db.ReorderGroup(ctx, exhausted, -1); err == nil {
		t.Fatal("revision overflow accepted")
	}
	after, err = db.Groups(ctx)
	if err != nil || !slices.Equal(exhausted.Siblings, after) {
		t.Fatalf("overflow changed order/version: %+v %v", after, err)
	}
}

func TestGroupOrderKeepsDefaultFirstWhenLegacySiblingsHaveEqualOrdinals(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	for _, id := range []string{"11111111111111111111111111111111", "22222222222222222222222222222222"} {
		if _, err = raw.Exec(`INSERT INTO groups VALUES(?,?,NULL,?,'',1,0)`, "scope", id, id); err != nil {
			t.Fatal(err)
		}
	}
	groups, err := db.Groups(ctx)
	if err != nil || len(groups) != 3 || groups[0].ID != session.DefaultGroupID {
		t.Fatalf("equal orders displaced reserved default: %+v %v", groups, err)
	}
	ordered, err := db.ReorderGroup(ctx, session.GroupOrder{TargetID: groups[2].ID, Siblings: groups}, -1)
	if err != nil || len(ordered) != 3 || ordered[0] != groups[0] || ordered[1].ID != groups[2].ID || ordered[2].ID != groups[1].ID || ordered[1].Order != 1 || ordered[2].Order != 2 {
		t.Fatalf("equal ordinals prevented an observable swap: %+v %v", ordered, err)
	}
}

func TestGroupOrderSkipsExpandedDescendantsAndPersistsTheOriginalTree(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	first, err := db.CreateGroup(ctx, "", "first", `C:\work`)
	if err != nil {
		t.Fatal(err)
	}
	child, err := db.CreateGroup(ctx, first.ID, "child", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.CreateGroup(ctx, "", "second", "")
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	moved, err := db.MoveToGroup(ctx, sample(), child.ID)
	if err != nil {
		t.Fatal(err)
	}
	groups, err := db.Groups(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expected := session.GroupOrder{TargetID: second.ID, Siblings: []session.Group{groups[0], first, second}}
	ordered, err := db.ReorderGroup(ctx, expected, -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(ordered) != 4 || ordered[0] != groups[0] || ordered[1].ID != second.ID || ordered[2].ID != first.ID || ordered[3] != child {
		t.Fatalf("did not move among root siblings: %+v", ordered)
	}
	for _, pair := range [][2]session.Group{{first, ordered[2]}, {second, ordered[1]}} {
		old, now := pair[0], pair[1]
		if old.ID != now.ID || old.Name != now.Name || old.ParentID != now.ParentID || old.DefaultDir != now.DefaultDir || now.Revision != old.Revision+1 {
			t.Fatalf("reorder changed group identity/config: %+v", pair)
		}
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	view, err := db.Workspace(ctx, []session.Session{sample()})
	if err != nil || len(view.Groups) != 4 || view.Groups[1] != ordered[1] || view.Groups[2] != ordered[2] || view.Groups[3] != child || len(view.Sessions) != 1 || view.Sessions[0] != moved {
		t.Fatalf("reopen lost order, descendants or original Run: %+v %v", view, err)
	}
}
