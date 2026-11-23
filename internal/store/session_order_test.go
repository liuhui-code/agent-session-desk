package store_test

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
	"path/filepath"
	"testing"
)

type sessionOrderStore interface {
	ReorderSession(context.Context, session.SessionOrder, int, []session.Session) (session.Workspace, error)
}

func TestSessionOrderPersistsOriginalRowsAcrossDiscoveryAndReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	items := []session.Session{sample(), sample(), sample()}
	for i := range items {
		items[i].ID = string("abc"[i]) + items[i].ID[1:]
		items[i].RunID = string("abc"[i]) + items[i].RunID[1:]
		items[i].BackendID = items[i].ID
		items[i].Title = string("ABC"[i])
		if err = db.Save(ctx, items[i]); err != nil {
			t.Fatal(err)
		}
	}
	view, err := db.Workspace(ctx, items)
	if err != nil {
		t.Fatal(err)
	}
	service, ok := any(db).(sessionOrderStore)
	if !ok {
		t.Fatal("Session rows have no persistent reorder operation")
	}
	view, err = service.ReorderSession(ctx, session.SessionOrder{TargetID: items[1].ID, Siblings: view.Sessions, Stamp: view.OrderStamps[session.DefaultGroupID]}, -1, items)
	if err != nil {
		t.Fatal(err)
	}
	if view.Sessions[0].ID != items[1].ID || view.Sessions[1].ID != items[0].ID || view.Sessions[2].ID != items[2].ID {
		t.Fatalf("swap not visible: %+v", view.Sessions)
	}
	for _, v := range view.Sessions {
		if v.OrderRevision < 1 {
			t.Fatal("manual peer positions not persisted")
		}
		for _, original := range items {
			if v.ID == original.ID && (v.RunID != original.RunID || v.BackendID != original.BackendID || v.Command != original.Command || v.GroupRevision != original.GroupRevision) {
				t.Fatal("ordering changed terminal/membership identity")
			}
		}
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	refreshed, err := db.Workspace(ctx, []session.Session{items[2], items[0], items[1]})
	if err != nil || !session.SameOrderSnapshot(view.Sessions, refreshed.Sessions) {
		t.Fatalf("reopen/discovery reordered original rows: %+v %v", refreshed, err)
	}
}
