package manager_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestGroupOrderNeedsNoTerminalAndKeepsTheOriginalSiblingIdentities(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "groups.sqlite3"), "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m := manager.New(nil, db)
	first, err := m.CreateGroup(ctx, "", "first", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := m.CreateGroup(ctx, "", "second", "")
	if err != nil {
		t.Fatal(err)
	}
	groups, err := db.Groups(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ordered, err := m.ReorderGroup(ctx, session.GroupOrder{TargetID: second.ID, Siblings: groups}, -1)
	if err != nil || len(ordered) != 3 || ordered[1].ID != second.ID || ordered[2].ID != first.ID {
		t.Fatalf("metadata order required/changed terminal: %+v %v", ordered, err)
	}
}
