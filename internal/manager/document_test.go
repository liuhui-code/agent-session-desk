package manager_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type documentTerminal struct {
	terminal
	directory string
	lookups   int
}

func (t *documentTerminal) CurrentDirectory(context.Context, session.Session) (string, error) {
	t.lookups++
	return t.directory, nil
}
func TestMarkdownResolutionKeepsOriginalRunDirectory(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	path := filepath.Join(root, "plan.md")
	if err := os.WriteFile(path, []byte("# test"), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(ctx, filepath.Join(root, "db.sqlite3"), "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	original := observed()
	backend := &documentTerminal{terminal: terminal{items: []session.Session{original}}, directory: root}
	m := manager.New(backend, db)
	expected, _ := filepath.EvalSymlinks(path)
	got, err := m.ResolveMarkdown(ctx, original, "plan.md")
	if err != nil || got != expected || backend.lookups != 1 {
		t.Fatalf("%q %v", got, err)
	}
	if _, err := m.ResolveMarkdown(ctx, original, path); err != nil || backend.lookups != 1 {
		t.Fatal("absolute path consulted cwd", err)
	}
	changed := original
	changed.BackendID = "another-pane"
	if _, err := m.ResolveMarkdown(ctx, changed, "plan.md"); err == nil {
		t.Fatal("changed pane accepted")
	}
	backend.items = nil
	if _, err := m.ResolveMarkdown(ctx, original, path); err == nil {
		t.Fatal("ended Run accepted")
	}
}

func TestCurrentDirectoryRejectsDifferentOriginalBeforeLookup(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "db.sqlite3"), "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	original := observed()
	backend := &documentTerminal{terminal: terminal{items: []session.Session{original}}, directory: t.TempDir()}
	m := manager.New(backend, db)
	dir, err := m.CurrentDirectory(ctx, original)
	if err != nil || dir != backend.directory || backend.lookups != 1 {
		t.Fatalf("directory: %q %v", dir, err)
	}
	changed := original
	changed.BackendID = "wrong-pane"
	if _, err = m.CurrentDirectory(ctx, changed); err == nil || backend.lookups != 1 {
		t.Fatal("changed identity reached directory lookup")
	}
	backend.items = nil
	if _, err = m.CurrentDirectory(ctx, original); err == nil {
		t.Fatal("ended original returned directory")
	}
}
