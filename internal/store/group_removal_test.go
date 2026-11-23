package store_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestGroupRemovalPreservesLiveAndHistoricalMembersOriginalRunsAfterReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "groups.sqlite3")
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	root, err := db.CreateGroup(ctx, "", "project", "")
	if err != nil {
		t.Fatal(err)
	}
	child, err := db.CreateGroup(ctx, root.ID, "review", "")
	if err != nil {
		t.Fatal(err)
	}
	other, err := db.CreateGroup(ctx, "", "unrelated", "")
	if err != nil {
		t.Fatal(err)
	}
	live := sample()
	ended := sample()
	ended.ID = "33333333333333333333333333333333"
	ended.RunID = "44444444444444444444444444444444"
	ended.BackendID = "asd_ended"
	for _, v := range []session.Session{live, ended} {
		if err = db.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	live, err = db.MoveToGroup(ctx, live, child.ID)
	if err != nil {
		t.Fatal(err)
	}
	savedOutput, err := db.RecordOutput(ctx, live, session.Preview{Text: "retained output"})
	if err != nil {
		t.Fatal(err)
	}
	ended, err = db.MoveToGroup(ctx, ended, root.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.MarkStopped(ctx, ended); err != nil {
		t.Fatal(err)
	}
	plan, err := db.PreviewGroupRemoval(ctx, root.ID)
	if err != nil || plan.Root != root || len(plan.Groups) != 2 || plan.Groups[1] != child || plan.MemberCount != 2 || plan.Destination.ID != session.DefaultGroupID || len(plan.Stamp) != 64 {
		t.Fatalf("preview omitted historical member/subtree: %+v %v", plan, err)
	}
	view, err := db.RemoveGroup(ctx, plan, []session.Session{live})
	if err != nil || len(view.Groups) != 2 || view.Groups[1] != other || len(view.Sessions) != 1 {
		t.Fatalf("removal lost unrelated group/live projection: %+v %v", view, err)
	}
	want := live
	want.GroupID = ""
	want.GroupRevision++
	if view.Sessions[0] != want {
		t.Fatalf("removal changed original live Run: %+v", view.Sessions[0])
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	history, err := db.History(ctx, 10)
	if err != nil || len(history) != 2 {
		t.Fatalf("removal deleted Run history: %+v %v", history, err)
	}
	for _, record := range history {
		v := record.Session
		if session.GroupOf(v) != session.DefaultGroupID || v.GroupRevision != 2 {
			t.Fatalf("history not collected into default: %+v", v)
		}
		if v.ID == ended.ID && (v.RunID != ended.RunID || record.StopConfirmedAt == nil) {
			t.Fatalf("historical exit/Run rewritten: %+v", record)
		}
	}
	output, err := db.Output(ctx, live.RunID)
	if err != nil || len(output) != 1 || output[0].ID != savedOutput.ID || output[0].Text != savedOutput.Text || output[0].RunID != live.RunID {
		t.Fatalf("group removal changed bounded output: %+v %v", output, err)
	}
}
