package manager_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestGroupRemovalPreviewsAndCollectsMembersWithoutATerminal(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "groups.sqlite3"), "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	root, err := db.CreateGroup(ctx, "", "project", "")
	if err != nil {
		t.Fatal(err)
	}
	v := observed()
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	v, err = db.MoveToGroup(ctx, v, root.ID)
	if err != nil {
		t.Fatal(err)
	}
	m := manager.New(nil, db)
	plan, err := m.PreviewGroupRemoval(ctx, root.ID)
	if err != nil || plan.MemberCount != 1 {
		t.Fatalf("preview needed terminal: %+v %v", plan, err)
	}
	view, err := m.RemoveGroup(ctx, plan, []session.Session{v})
	if err != nil || len(view.Groups) != 1 || len(view.Sessions) != 1 || view.Sessions[0].RunID != v.RunID || session.GroupOf(view.Sessions[0]) != session.DefaultGroupID {
		t.Fatalf("removal needed/changed terminal: %+v %v", view, err)
	}
}
