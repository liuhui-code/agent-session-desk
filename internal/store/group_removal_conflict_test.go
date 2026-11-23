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

type removalFixture struct {
	ctx                context.Context
	db                 *store.Store
	path               string
	root, child, other session.Group
	first, second      session.Session
}

func newRemovalFixture(t *testing.T) *removalFixture {
	t.Helper()
	f := &removalFixture{ctx: context.Background(), path: filepath.Join(t.TempDir(), "groups.sqlite3"), first: sample(), second: sample()}
	var err error
	f.db, err = store.Open(f.ctx, f.path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.db.Close() })
	f.root, err = f.db.CreateGroup(f.ctx, "", "project", "")
	if err != nil {
		t.Fatal(err)
	}
	f.child, err = f.db.CreateGroup(f.ctx, f.root.ID, "review", "")
	if err != nil {
		t.Fatal(err)
	}
	f.other, err = f.db.CreateGroup(f.ctx, "", "other", "")
	if err != nil {
		t.Fatal(err)
	}
	f.second.ID = "33333333333333333333333333333333"
	f.second.RunID = "44444444444444444444444444444444"
	f.second.BackendID = "asd_second"
	for _, v := range []session.Session{f.first, f.second} {
		if err = f.db.Save(f.ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	f.first, err = f.db.MoveToGroup(f.ctx, f.first, f.child.ID)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestGroupRemovalRejectsChangedTreeSameCountMemberReplacementAndABA(t *testing.T) {
	for _, change := range []string{"new-child", "destination-name", "unrelated-group-name", "same-count-replacement", "membership-ABA"} {
		t.Run(change, func(t *testing.T) {
			f := newRemovalFixture(t)
			plan, err := f.db.PreviewGroupRemoval(f.ctx, f.root.ID)
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "new-child":
				_, err = f.db.CreateGroup(f.ctx, f.root.ID, "new child", "")
			case "destination-name":
				_, err = f.db.RenameGroup(f.ctx, plan.Destination, "new default name")
			case "unrelated-group-name":
				_, err = f.db.RenameGroup(f.ctx, f.other, "new other name")
			case "same-count-replacement", "membership-ABA":
				f.first, err = f.db.MoveToGroup(f.ctx, f.first, session.DefaultGroupID)
				if err == nil {
					if change == "membership-ABA" {
						f.first, err = f.db.MoveToGroup(f.ctx, f.first, f.child.ID)
					} else {
						f.second, err = f.db.MoveToGroup(f.ctx, f.second, f.child.ID)
					}
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = f.db.RemoveGroup(f.ctx, plan, nil); !errors.Is(err, store.ErrGroupConflict) {
				t.Fatalf("changed preview accepted: %v", err)
			}
			after, err := f.db.PreviewGroupRemoval(f.ctx, f.root.ID)
			if err != nil || after.Root.ID != f.root.ID || after.Stamp == plan.Stamp {
				t.Fatalf("stale removal changed/lost root: %+v %v", after, err)
			}
			if change == "same-count-replacement" || change == "membership-ABA" {
				if after.MemberCount != plan.MemberCount {
					t.Fatal("test did not retain same count")
				}
			}
		})
	}
}

func TestGroupRemovalPreservesIndependentMembershipAndNameEditsInItsProjection(t *testing.T) {
	f := newRemovalFixture(t)
	plan, err := f.db.PreviewGroupRemoval(f.ctx, f.root.ID)
	if err != nil {
		t.Fatal(err)
	}
	named, err := f.db.Rename(f.ctx, f.second, "独立名称")
	if err != nil {
		t.Fatal(err)
	}
	moved, err := f.db.MoveToGroup(f.ctx, named, f.other.ID)
	if err != nil {
		t.Fatal(err)
	}
	view, err := f.db.RemoveGroup(f.ctx, plan, []session.Session{f.first, f.second})
	if err != nil || len(view.Sessions) != 2 || view.Sessions[1] != moved || view.Sessions[0].RunID != f.first.RunID || session.GroupOf(view.Sessions[0]) != session.DefaultGroupID {
		t.Fatalf("removal overwrote independent metadata/runtime: %+v %v", view, err)
	}
}

func TestGroupRemovalPreviewUsesAReadOnlySnapshotWhileAWriterIsReserved(t *testing.T) {
	f := newRemovalFixture(t)
	raw, err := sql.Open("sqlite", f.path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if _, err = raw.Exec("BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	defer raw.Exec("ROLLBACK")
	plan, err := f.db.PreviewGroupRemoval(f.ctx, f.root.ID)
	if err != nil || plan.MemberCount != 1 || len(plan.Groups) != 2 {
		t.Fatalf("read-only preview required writer lock or missed records: %+v %v", plan, err)
	}
}

func TestGroupRemovalConcurrentConnectionsOnlyOneCommits(t *testing.T) {
	f := newRemovalFixture(t)
	other, err := store.Open(f.ctx, f.path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	plan, err := f.db.PreviewGroupRemoval(f.ctx, f.root.ID)
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, db := range []*store.Store{f.db, other} {
		go func(db *store.Store) { <-start; _, e := db.RemoveGroup(f.ctx, plan, nil); results <- e }(db)
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
		t.Fatalf("concurrent removals: %d success %d conflicts", success, conflicts)
	}
	view, err := f.db.Workspace(f.ctx, []session.Session{f.first})
	if err != nil || len(view.Groups) != 2 || view.Sessions[0].GroupRevision != 2 {
		t.Fatalf("membership moved more than once: %+v %v", view, err)
	}
}

func TestGroupRemovalRollsBackOnProjectionSQLFailureAndVersionExhaustion(t *testing.T) {
	for _, failure := range []string{"projection", "parent-delete", "membership-exhaustion"} {
		t.Run(failure, func(t *testing.T) {
			f := newRemovalFixture(t)
			raw, err := sql.Open("sqlite", f.path)
			if err != nil {
				t.Fatal(err)
			}
			defer raw.Close()
			if failure == "parent-delete" {
				_, err = raw.Exec(`CREATE TRIGGER reject_parent BEFORE DELETE ON groups WHEN OLD.id='` + f.root.ID + `' BEGIN SELECT RAISE(ABORT,'synthetic parent delete failure'); END`)
			}
			if failure == "membership-exhaustion" {
				_, err = raw.Exec(`UPDATE session_groups SET revision=? WHERE scope=? AND session_id=?`, int64(math.MaxInt64), "scope", f.first.ID)
			}
			if err != nil {
				t.Fatal(err)
			}
			plan, err := f.db.PreviewGroupRemoval(f.ctx, f.root.ID)
			if err != nil {
				t.Fatal(err)
			}
			var observed []session.Session
			if failure == "projection" {
				observed = []session.Session{{ID: "ffffffffffffffffffffffffffffffff"}}
			}
			if _, err = f.db.RemoveGroup(f.ctx, plan, observed); err == nil {
				t.Fatal("synthetic failure was ignored")
			}
			after, err := f.db.PreviewGroupRemoval(f.ctx, f.root.ID)
			if err != nil || after.Stamp != plan.Stamp || after.MemberCount != plan.MemberCount || len(after.Groups) != 2 {
				t.Fatalf("failure left partial member/tree changes: %+v %v", after, err)
			}
		})
	}
}

func TestGroupRemovalProtectsDefaultScopeAndTheReviewedPlanFields(t *testing.T) {
	f := newRemovalFixture(t)
	if _, err := f.db.PreviewGroupRemoval(f.ctx, session.DefaultGroupID); err == nil {
		t.Fatal("default removal preview accepted")
	}
	groups, err := f.db.Groups(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.RemoveGroup(f.ctx, session.GroupRemoval{Root: groups[0]}, nil); err == nil {
		t.Fatal("forged default removal accepted")
	}
	plan, err := f.db.PreviewGroupRemoval(f.ctx, f.root.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"count", "destination", "groups", "stamp"} {
		forged := plan
		switch field {
		case "count":
			forged.MemberCount = 0
		case "destination":
			forged.Destination.Name = "misleading destination"
		case "groups":
			forged.Groups = forged.Groups[:1]
		case "stamp":
			forged.Stamp = ""
		}
		if _, err = f.db.RemoveGroup(f.ctx, forged, nil); !errors.Is(err, store.ErrGroupConflict) {
			t.Fatalf("forged %s accepted: %v", field, err)
		}
	}
	other, err := store.Open(f.ctx, f.path, "other-scope")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if _, err = other.RemoveGroup(f.ctx, plan, nil); !errors.Is(err, store.ErrGroupConflict) {
		t.Fatalf("foreign removal accepted: %v", err)
	}
	after, err := f.db.PreviewGroupRemoval(f.ctx, f.root.ID)
	if err != nil || after.Stamp != plan.Stamp {
		t.Fatal("rejected operation changed reviewed metadata")
	}
}
