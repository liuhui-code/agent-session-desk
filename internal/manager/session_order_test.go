package manager_test

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
	"path/filepath"
	"testing"
)

func TestSessionOrderRevalidatesDiscoveryAndNeverStartsOrStopsTerminals(t *testing.T) {
	ctx := context.Background()
	db, e := store.Open(ctx, filepath.Join(t.TempDir(), "order.sqlite3"), "scope")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	a, b := observed(), observed()
	a.Title = "A"
	b.ID = "33333333333333333333333333333333"
	b.RunID = "44444444444444444444444444444444"
	b.BackendID = "other"
	b.Title = "B"
	f := &terminal{items: []session.Session{b, a}}
	m := manager.New(f, db)
	view, e := m.Workspace(ctx)
	if e != nil {
		t.Fatal(e)
	}
	expected := session.SessionOrder{TargetID: b.ID, Siblings: view.Sessions, Stamp: view.OrderStamps[session.DefaultGroupID]}
	view, e = m.ReorderSession(ctx, expected, -1)
	if e != nil || view.Sessions[0].ID != b.ID {
		t.Fatalf("manager swap: %+v %v", view, e)
	}
	expected = session.SessionOrder{TargetID: b.ID, Siblings: view.Sessions, Stamp: view.OrderStamps[session.DefaultGroupID]}
	f.items = []session.Session{b}
	if _, e = m.ReorderSession(ctx, expected, 1); e == nil {
		t.Fatal("disappeared peer accepted")
	}
	if f.starts != 0 || f.stops != 0 || f.items[0].OrderRevision != 0 {
		t.Fatal("position changed raw backend or terminal lifetime")
	}
}
