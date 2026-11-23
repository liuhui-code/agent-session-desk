package store_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestRenamedRunStillRecordsOutputWithoutChangingLaunchSnapshot(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := sample()
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	renamed, err := db.Rename(ctx, v, "新名称")
	if err != nil {
		t.Fatal(err)
	}
	record, err := db.RecordOutput(ctx, renamed, session.Preview{Text: "same pane"})
	if err != nil || record.RunID != v.RunID {
		t.Fatalf("rename broke capture: %+v %v", record, err)
	}
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	mutated := v
	mutated.Title = renamed.Title
	if err = db.Save(ctx, mutated); err == nil {
		t.Fatal("display name weakened immutable launch snapshot")
	}
}

func TestConcurrentNameWritersCommitOnlyOneRevision(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	a, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	v := sample()
	if err = a.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	start, result := make(chan struct{}), make(chan error, 2)
	for i, db := range []*store.Store{a, b} {
		go func(db *store.Store, name string) { <-start; _, e := db.Rename(ctx, v, name); result <- e }(db, []string{"A", "B"}[i])
	}
	close(start)
	passed, conflicts := 0, 0
	for range 2 {
		err := <-result
		if err == nil {
			passed++
		} else if errors.Is(err, store.ErrNameConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if passed != 1 || conflicts != 1 {
		t.Fatalf("commits=%d conflicts=%d", passed, conflicts)
	}
	items, err := a.WithNames(ctx, []session.Session{v})
	if err != nil || items[0].TitleRevision != 1 {
		t.Fatalf("revision: %+v %v", items, err)
	}
}

func TestNamesStayScopedAndRejectUnrelatedRunOrExpectedTitle(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	a, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := store.Open(ctx, path, "b")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	v := sample()
	for _, db := range []*store.Store{a, b} {
		if err = db.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	for _, mutate := range []func(*session.Session){func(v *session.Session) { v.ID = "unrelated" }, func(v *session.Session) { v.RunID = "unrelated" }, func(v *session.Session) { v.BackendID = "unrelated" }, func(v *session.Session) { v.Command = "unrelated" }, func(v *session.Session) { v.Title = "unrelated" }, func(v *session.Session) { v.TitleRevision = 1 }} {
		wrong := v
		mutate(&wrong)
		if _, err = a.Rename(ctx, wrong, "new"); err == nil {
			t.Fatal("unrelated/stale target accepted")
		}
	}
	renamed, err := a.Rename(ctx, v, "A")
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := a.Rename(ctx, renamed, " A ")
	if err != nil || unchanged.TitleRevision != renamed.TitleRevision {
		t.Fatalf("no-op manufactured a revision: %+v %v", unchanged, err)
	}
	rows, err := b.WithNames(ctx, []session.Session{v})
	if err != nil || rows[0].Title != v.Title || rows[0].TitleRevision != 0 {
		t.Fatalf("scope leaked: %+v %v", rows, err)
	}
}
