package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type renameOrderFixture struct{ *sessionOrderFixture }

func (s *renameOrderFixture) Rename(_ context.Context, target session.Session, name string) (session.Session, error) {
	s.orderMu.Lock()
	defer s.orderMu.Unlock()
	for i := range s.rows {
		if sameRun(s.rows[i], target) {
			s.rows[i].Title, s.rows[i].TitleRevision = name, target.TitleRevision+1
			renamed := s.rows[i]
			session.SortSessions(s.rows)
			return renamed, nil
		}
	}
	return session.Session{}, context.Canceled
}

func TestRenameRefreshesNameOrderBeforeMovingSelectedOriginalSession(t *testing.T) {
	s := &renameOrderFixture{newSessionOrderFixture()}
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.PasteMsg{Content: "original draft"})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	s.mu.Lock()
	original := s.streams["first"]
	s.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return h.m.rename != nil })
	h.send(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	h.send(tea.PasteMsg{Content: "龘"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return h.m.rename == nil && !h.m.listBusy })
	if h.m.sessions[0].ID != "second" {
		t.Fatal("successful rename left old name order instead of reconciling discovery")
	}
	selected, ok := h.m.selected()
	if !ok || selected.ID != "first" || h.m.workspace.tabs[h.m.workspace.active].session.ID != "first" || h.m.workspace.focused {
		t.Fatal("rename refresh lost original selection, active Tab, or management focus")
	}
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.until(func() bool { return h.m.sessionOrder == nil && h.m.sessions[0].ID == "first" })
	s.orderMu.Lock()
	defer s.orderMu.Unlock()
	if s.calls != 1 {
		t.Fatal("rename/order was not a single reviewed move")
	}
	original.mu.Lock()
	defer original.mu.Unlock()
	if got := strings.Join(original.writes, ";"); strings.Count(got, "paste:original draft") != 1 || strings.Contains(got, "key:enter") {
		t.Fatalf("management changed or submitted original draft: %q", got)
	}
}
