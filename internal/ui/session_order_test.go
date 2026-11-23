package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"strings"
	"sync"
	"testing"
)

type sessionOrderFixture struct {
	workspaceFixture
	orderMu  sync.Mutex
	rows     []session.Session
	calls    int
	gate     <-chan struct{}
	orderErr error
}

func newSessionOrderFixture() *sessionOrderFixture {
	s := &sessionOrderFixture{rows: sampleSessions()}
	s.fakeService.list = func(context.Context) ([]session.Session, error) {
		s.orderMu.Lock()
		defer s.orderMu.Unlock()
		return append([]session.Session(nil), s.rows...), nil
	}
	return s
}
func (s *sessionOrderFixture) ReorderSession(ctx context.Context, e session.SessionOrder, delta int) (session.Workspace, error) {
	s.orderMu.Lock()
	s.calls++
	gate, err := s.gate, s.orderErr
	s.orderMu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return session.Workspace{}, ctx.Err()
		}
	}
	if err != nil {
		return session.Workspace{}, err
	}
	s.orderMu.Lock()
	defer s.orderMu.Unlock()
	peers := session.OrderPeers(s.rows, e.TargetID)
	if !session.SameOrderSnapshot(peers, e.Siblings) {
		return session.Workspace{}, context.Canceled
	}
	for i := range peers {
		if peers[i].ID == e.TargetID {
			peers[i], peers[i+delta] = peers[i+delta], peers[i]
			break
		}
	}
	for i, v := range peers {
		for j := range s.rows {
			if s.rows[j].ID == v.ID {
				s.rows[j].Order = int64(i)
				s.rows[j].OrderRevision++
			}
		}
	}
	session.SortSessions(s.rows)
	return session.Workspace{Sessions: append([]session.Session(nil), s.rows...)}, nil
}
func TestSessionRowOrderFollowsSelectionWithoutMovingTabsOrSendingInput(t *testing.T) {
	s := newSessionOrderFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.until(func() bool { s.orderMu.Lock(); defer s.orderMu.Unlock(); return s.calls == 1 })
	h.until(func() bool { return h.m.sessions[0].ID == "second" })
	if h.m.cursor != 0 || h.m.workspace.tabs[h.m.workspace.active].session.ID != "first" || h.m.workspace.focused {
		t.Fatal("order changed selection identity, Tab or focus")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stream := range s.streams {
		stream.mu.Lock()
		for _, v := range stream.writes {
			if strings.HasPrefix(v, "text:") || strings.HasPrefix(v, "key:") {
				t.Fatal("management ordering leaked terminal input")
			}
		}
		stream.mu.Unlock()
	}
}

func TestSessionOrderAllSourceKeysKeepSelectionAndNoOpBoundaries(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{{Code: 'K', Text: "K"}, {Code: '+', Text: "+"}, {Code: tea.KeyUp, Mod: tea.ModShift}, {Code: tea.KeyUp, Mod: tea.ModCtrl}, {Code: 'J', Text: "J"}, {Code: '-', Text: "-"}, {Code: tea.KeyDown, Mod: tea.ModShift}, {Code: tea.KeyDown, Mod: tea.ModCtrl}} {
		t.Run(key.String(), func(t *testing.T) {
			s := newSessionOrderFixture()
			h := newWorkspaceHarness(t, s)
			h.until(func() bool { return h.m.workspace.conn != nil })
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			up := key.String() == "K" || key.String() == "+" || key.String() == "shift+up" || key.String() == "ctrl+up"
			if up {
				h.send(tea.KeyPressMsg{Code: tea.KeyDown})
			}
			h.send(key)
			h.until(func() bool { return h.m.sessionOrder == nil && h.m.sessions[0].ID == "second" })
			selected, _ := h.m.selected()
			if selected.ID != map[bool]string{true: "second", false: "first"}[up] {
				t.Fatal("sort selected by row index")
			}
			h.send(key)
			s.orderMu.Lock()
			defer s.orderMu.Unlock()
			if s.calls != 1 {
				t.Fatal("boundary queued another save")
			}
		})
	}
}
