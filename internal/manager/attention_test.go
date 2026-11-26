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

func TestAttentionOpenRevalidatesOriginalRunAndNeverControlsTerminal(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "db.sqlite3"), "a")
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
	e := session.AttentionEvent{ID: strings.Repeat("a", 32), Version: 1, Kind: "approval", State: "pending", Title: "Allow?"}
	if _, err = m.ReportAttention(ctx, c, e); err != nil {
		t.Fatal(err)
	}
	view, err := m.Workspace(ctx)
	if err != nil || len(view.Attention) != 1 || !view.Attention[0].TerminalAvailable {
		t.Fatal("pending request missing", err)
	}
	f.items[0].RunID = strings.Repeat("b", 32)
	if _, err = m.OpenAttention(ctx, e.ID, 1); err == nil {
		t.Fatal("activation followed replacement Run")
	}
	item, err := db.Attention(ctx, e.ID)
	if err != nil || item.ViewedAt != nil || item.State != "pending" {
		t.Fatal("failed navigation hid request")
	}
	if _, err = m.ReportAttention(ctx, c, e); err == nil {
		t.Fatal("stale producer accepted")
	}
	f.items[0] = v
	if _, err = m.OpenAttention(ctx, e.ID, 2); err == nil {
		t.Fatal("stale version accepted")
	}
	item, err = m.OpenAttention(ctx, e.ID, 1)
	if err != nil || item.ViewedAt == nil || item.State != "pending" {
		t.Fatal("view resolved request", err)
	}
	if f.starts != 0 || f.stops != 0 {
		t.Fatal("notification changed process lifetime")
	}
	e.Version = 2
	e.State = "resolved"
	if _, err = m.ReportAttention(ctx, c, e); err != nil {
		t.Fatal(err)
	}
	if _, err = m.OpenAttention(ctx, e.ID, 1); err == nil {
		t.Fatal("resolved notification still navigated")
	}
}
