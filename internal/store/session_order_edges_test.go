package store_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
	"math"
	"path/filepath"
	"sync"
	"testing"
)

func orderFixture(t *testing.T) (*store.Store, []session.Session, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "order.sqlite3")
	db, err := store.Open(context.Background(), path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	var rows []session.Session
	for i := 0; i < 3; i++ {
		v := sample()
		v.ID = fmt.Sprintf("%032x", i+1)
		v.RunID = fmt.Sprintf("%032x", i+11)
		v.BackendID = v.ID
		v.Title = fmt.Sprintf("row-%d", i)
		if err = db.Save(context.Background(), v); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, v)
	}
	return db, rows, path
}
func orderSnapshot(t *testing.T, db *store.Store, rows []session.Session, id string) session.SessionOrder {
	t.Helper()
	view, err := db.Workspace(context.Background(), rows)
	if err != nil {
		t.Fatal(err)
	}
	return session.SessionOrder{TargetID: id, Siblings: session.OrderPeers(view.Sessions, id), Stamp: view.OrderStamps[session.GroupOf(session.OrderPeers(view.Sessions, id)[0])]}
}
func TestSessionOrderRejectsStaleSnapshotsABAAndObservedSetChanges(t *testing.T) {
	ctx := context.Background()
	db, rows, _ := orderFixture(t)
	old := orderSnapshot(t, db, rows, rows[1].ID)
	if _, e := db.ReorderSession(ctx, old, -1, rows); e != nil {
		t.Fatal(e)
	}
	if _, e := db.ReorderSession(ctx, old, -1, rows); !errors.Is(e, store.ErrSessionOrderConflict) {
		t.Fatalf("stale write: %v", e)
	}
	current := orderSnapshot(t, db, rows, rows[1].ID)
	if _, e := db.ReorderSession(ctx, current, 1, rows); e != nil {
		t.Fatal(e)
	}
	if _, e := db.ReorderSession(ctx, old, -1, rows); !errors.Is(e, store.ErrSessionOrderConflict) {
		t.Fatalf("ABA write: %v", e)
	}
	for _, kind := range []string{"missing", "inserted", "replacement", "rename", "membership", "stopped", "duplicate", "foreign"} {
		t.Run(kind, func(t *testing.T) {
			db, items, path := orderFixture(t)
			expected := orderSnapshot(t, db, items, items[1].ID)
			observed := append([]session.Session(nil), items...)
			target := db
			switch kind {
			case "missing":
				observed = observed[:2]
			case "inserted":
				v := sample()
				v.ID = "ffffffffffffffffffffffffffffffff"
				v.RunID = v.ID
				v.BackendID = v.ID
				v.Title = "new"
				if e := db.Save(ctx, v); e != nil {
					t.Fatal(e)
				}
				observed = append(observed, v)
			case "replacement":
				observed[0].RunID = "ffffffffffffffffffffffffffffffff"
				observed[0].BackendID = "replacement"
				if e := db.Save(ctx, observed[0]); e != nil {
					t.Fatal(e)
				}
			case "rename":
				if _, e := db.Rename(ctx, items[0], "new name"); e != nil {
					t.Fatal(e)
				}
			case "membership":
				g, e := db.CreateGroup(ctx, "", "other", "")
				if e != nil {
					t.Fatal(e)
				}
				if _, e = db.MoveToGroup(ctx, items[0], g.ID); e != nil {
					t.Fatal(e)
				}
			case "stopped":
				if e := db.MarkStopped(ctx, items[0]); e != nil {
					t.Fatal(e)
				}
			case "duplicate":
				observed = append(observed, items[0])
			case "foreign":
				var e error
				target, e = store.Open(ctx, path, "foreign")
				if e != nil {
					t.Fatal(e)
				}
				defer target.Close()
			}
			if _, e := target.ReorderSession(ctx, expected, -1, observed); e == nil {
				t.Fatal("changed ordered review accepted")
			}
		})
	}
}
func TestSessionOrderBoundariesNoWriteAndTransactionalFailure(t *testing.T) {
	ctx := context.Background()
	db, rows, path := orderFixture(t)
	for _, c := range []struct{ i, d int }{{0, -1}, {2, 1}} {
		before := orderSnapshot(t, db, rows, rows[c.i].ID)
		if _, e := db.ReorderSession(ctx, before, c.d, rows); e != nil {
			t.Fatal(e)
		}
		after := orderSnapshot(t, db, rows, rows[c.i].ID)
		if !session.SameOrderSnapshot(before.Siblings, after.Siblings) {
			t.Fatal("boundary wrote metadata")
		}
	}
	expected := orderSnapshot(t, db, rows, rows[1].ID)
	for _, d := range []int{-2, 0, 2} {
		if _, e := db.ReorderSession(ctx, expected, d, rows); e == nil {
			t.Fatal("invalid direction accepted")
		}
	}
	raw, e := sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	defer raw.Close()
	if _, e = raw.Exec(`CREATE TRIGGER fail_position BEFORE INSERT ON session_positions WHEN NEW.session_id='` + rows[1].ID + `' BEGIN SELECT RAISE(ABORT,'synthetic position failure'); END`); e != nil {
		t.Fatal(e)
	}
	if _, e = db.ReorderSession(ctx, expected, -1, rows); e == nil {
		t.Fatal("failed write accepted")
	}
	after := orderSnapshot(t, db, rows, rows[1].ID)
	if !session.SameOrderSnapshot(expected.Siblings, after.Siblings) {
		t.Fatal("first position write did not roll back")
	}
	if _, e = raw.Exec(`DROP TRIGGER fail_position`); e != nil {
		t.Fatal(e)
	}
	if _, e = db.ReorderSession(ctx, expected, -1, rows); e != nil {
		t.Fatal(e)
	}
	if _, e = raw.Exec(`UPDATE session_positions SET revision=? WHERE scope='scope' AND session_id=?`, int64(math.MaxInt64), rows[1].ID); e != nil {
		t.Fatal(e)
	}
	expected = orderSnapshot(t, db, rows, rows[1].ID)
	if _, e = db.ReorderSession(ctx, expected, 1, rows); e == nil {
		t.Fatal("revision overflow accepted")
	}
	after = orderSnapshot(t, db, rows, rows[1].ID)
	if !session.SameOrderSnapshot(expected.Siblings, after.Siblings) {
		t.Fatal("overflow partially wrote positions")
	}
}
func TestSessionOrderConcurrentConnectionsOnlyOneCommits(t *testing.T) {
	ctx := context.Background()
	db, rows, path := orderFixture(t)
	other, e := store.Open(ctx, path, "scope")
	if e != nil {
		t.Fatal(e)
	}
	defer other.Close()
	expected := orderSnapshot(t, db, rows, rows[1].ID)
	gate := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, connection := range []*store.Store{db, other} {
		wg.Add(1)
		go func(s *store.Store) {
			defer wg.Done()
			<-gate
			_, e := s.ReorderSession(ctx, expected, -1, rows)
			results <- e
		}(connection)
	}
	close(gate)
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("expected one commit, got %d", success)
	}
}
func TestSessionOrderNewAndMovedMembersFollowManualRowsAndRemovalClearsOldPositions(t *testing.T) {
	ctx := context.Background()
	db, rows, _ := orderFixture(t)
	g, e := db.CreateGroup(ctx, "", "other", "")
	if e != nil {
		t.Fatal(e)
	}
	moved, e := db.MoveToGroup(ctx, rows[2], g.ID)
	if e != nil {
		t.Fatal(e)
	}
	view, e := db.Workspace(ctx, rows)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.ReorderSession(ctx, session.SessionOrder{TargetID: rows[1].ID, Siblings: session.OrderPeers(view.Sessions, rows[1].ID), Stamp: view.OrderStamps[session.DefaultGroupID]}, -1, rows); e != nil {
		t.Fatal(e)
	}
	moved, e = db.MoveToGroup(ctx, moved, session.DefaultGroupID)
	if e != nil {
		t.Fatal(e)
	}
	view, e = db.Workspace(ctx, rows)
	if e != nil || view.Sessions[0].ID != rows[1].ID || view.Sessions[2].ID != moved.ID || view.Sessions[2].OrderRevision != 0 {
		t.Fatalf("move/reset wrong: %+v %v", view, e)
	}
	for _, v := range view.Sessions {
		if _, e = db.MoveToGroup(ctx, v, g.ID); e != nil {
			t.Fatal(e)
		}
	}
	expected := orderSnapshot(t, db, rows, rows[1].ID)
	if _, e = db.ReorderSession(ctx, expected, -1, rows); e != nil {
		t.Fatal(e)
	}
	plan, e := db.PreviewGroupRemoval(ctx, g.ID)
	if e != nil {
		t.Fatal(e)
	}
	view, e = db.RemoveGroup(ctx, plan, rows)
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range view.Sessions {
		if session.GroupOf(v) != session.DefaultGroupID || v.OrderRevision != 0 {
			t.Fatalf("group removal kept stale position: %+v", v)
		}
	}
}

func TestSessionOrderRejectsMetadataInsertionAfterDiscoveryWithoutReplaying(t *testing.T) {
	ctx := context.Background()
	db, rows, _ := orderFixture(t)
	expected := orderSnapshot(t, db, rows, rows[1].ID)
	added := sample()
	added.ID = "ffffffffffffffffffffffffffffffff"
	added.RunID = added.ID
	added.BackendID = "new-between-discovery-and-commit"
	added.Title = "new member"
	if e := db.Save(ctx, added); e != nil {
		t.Fatal(e)
	}
	if _, e := db.ReorderSession(ctx, expected, -1, rows); !errors.Is(e, store.ErrSessionOrderConflict) {
		t.Fatalf("metadata-only new member did not invalidate reviewed group: %v", e)
	}
	current := orderSnapshot(t, db, rows, rows[1].ID)
	for _, v := range current.Siblings {
		if v.OrderRevision != 0 {
			t.Fatal("rejected operation partially wrote positions")
		}
	}
}
