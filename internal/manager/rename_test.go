package manager_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type sessionRenamer interface {
	Rename(context.Context, session.Session, string) (session.Session, error)
}

func TestRenameRejectsStaleRevisionEvenAfterNameReturnsToOriginal(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := &terminal{items: []session.Session{observed()}}
	m := manager.New(f, db)
	items, err := m.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	original := items[0]
	changed, err := m.Rename(ctx, original, "临时名")
	if err != nil {
		t.Fatal(err)
	}
	back, err := m.Rename(ctx, changed, original.Title)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Rename(ctx, original, "过期覆盖"); !errors.Is(err, store.ErrNameConflict) {
		t.Fatalf("stale editor accepted after ABA: %v", err)
	}
	items, err = m.List(ctx)
	if err != nil || items[0].Title != original.Title || items[0].TitleRevision != back.TitleRevision {
		t.Fatalf("conflict changed persisted label: %+v %v", items, err)
	}
}

func TestRenamePersistsDisplayNameWithoutChangingBackendRun(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	f := &terminal{items: []session.Session{observed()}}
	m := manager.New(f, db)
	items, err := m.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rename, ok := any(m).(sessionRenamer)
	if !ok {
		t.Fatal("manager has no stable-identity rename operation")
	}
	got, err := rename.Rename(ctx, items[0], "  中文评审 ; 'literal'  ")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != items[0].ID || got.RunID != items[0].RunID || got.BackendID != items[0].BackendID || got.Title != "中文评审 ; 'literal'" {
		t.Fatalf("rename changed identity or text: %+v", got)
	}
	if f.starts != 0 || f.stops != 0 || f.items[0].Title != observed().Title {
		t.Fatal("display rename changed the running backend")
	}
	db.Close()
	db, err = store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	items, err = manager.New(f, db).List(ctx)
	if err != nil || len(items) != 1 || items[0].Title != got.Title || items[0].RunID != got.RunID {
		t.Fatalf("old discovery overwrote persisted name: %+v %v", items, err)
	}
	history, err := db.History(ctx, 100)
	if err != nil || len(history) != 1 || history[0].Session.Title != got.Title {
		t.Fatalf("offline name not persisted: %+v %v", history, err)
	}
}

func TestRenameRefusesTargetNoLongerDiscovered(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := &terminal{items: []session.Session{observed()}}
	m := manager.New(f, db)
	items, err := m.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	f.items = nil
	if _, err = m.Rename(ctx, items[0], "should not save"); err == nil {
		t.Fatal("missing target renamed")
	}
	history, err := db.History(ctx, 100)
	if err != nil || history[0].Session.Title != observed().Title {
		t.Fatal("missing target changed offline name")
	}
	if f.starts != 0 || f.stops != 0 {
		t.Fatal("rename changed terminal lifetime")
	}
}
