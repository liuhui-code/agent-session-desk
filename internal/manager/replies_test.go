package manager_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestReplyRechecksLiveIdentityWithoutTerminalWrites(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := observed()
	f := &terminal{items: []session.Session{v}}
	m := manager.New(f, db)
	c, err := session.NewAgentCredential(v.ID, v.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.BindAgent(ctx, c); err != nil {
		t.Fatal(err)
	}
	q, err := m.QueueMessage(ctx, session.MessageRequest{RequestID: strings.Repeat("a", 32), TargetSessionID: v.ID, TargetRunID: v.RunID, Body: "request"})
	if err != nil {
		t.Fatal(err)
	}
	f.items[0].RunID = strings.Repeat("b", 32)
	if _, err = m.Inbox(ctx, c); err == nil {
		t.Fatal("old credential followed new Run")
	}
	r := session.ReplyRequest{ID: strings.Repeat("c", 32), ReplyTo: q.ID, Body: "result"}
	if _, err = m.Reply(ctx, c, r); err == nil {
		t.Fatal("stale source replied")
	}
	f.items[0] = v
	rows, err := m.Inbox(ctx, c)
	if err != nil || len(rows) != 1 {
		t.Fatalf("inbox %+v %v", rows, err)
	}
	if _, err = m.Reply(ctx, c, r); err != nil {
		t.Fatal(err)
	}
	if f.starts != 0 || f.stops != 0 {
		t.Fatal("reply changed terminal lifetime")
	}
	q, err = m.Message(ctx, q.ID)
	if err != nil || q.State != "queued" {
		t.Fatal("reply changed delivery evidence")
	}
}
