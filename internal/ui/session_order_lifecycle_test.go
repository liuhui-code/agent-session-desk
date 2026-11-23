package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"strings"
	"testing"
)

func TestSessionOrderQuitWaitsForMetadataInputAndPreferenceAndNeverReplays(t *testing.T) {
	s := newSessionOrderFixture()
	metadata, input, prefs := make(chan struct{}), make(chan struct{}), make(chan struct{})
	for _, g := range []chan struct{}{metadata, input, prefs} {
		t.Cleanup(func() { releaseGroupOrderGate(g) })
	}
	s.gate = metadata
	h := newWorkspaceHarness(t, s, Options{SaveUISettings: func(ctx context.Context, _ config.UISettings) error {
		select {
		case <-prefs:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return h.m.workspace.conn != nil })
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	stream.mu.Lock()
	stream.gate = input
	stream.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "accepted-draft"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.send(tea.KeyPressMsg{Code: 'v', Mod: tea.ModAlt})
	if h.m.sessionOrder == nil || !h.m.preferences.busy || h.m.workspace.pending == 0 {
		t.Fatal("three accepted effects were not held")
	}
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.PasteMsg{Content: "must-not-leak"})
	if h.m.closed || h.m.workspace.focused {
		t.Fatal("quit discarded save or granted input focus")
	}
	close(metadata)
	h.until(func() bool { return h.m.sessionOrder == nil })
	if h.m.closed {
		t.Fatal("save abandoned input/preferences")
	}
	close(input)
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if h.m.closed {
		t.Fatal("input completion abandoned preferences")
	}
	close(prefs)
	h.until(func() bool { return h.m.closed })
	stream.mu.Lock()
	defer stream.mu.Unlock()
	writes := strings.Join(stream.writes, ",")
	if strings.Count(writes, "text:accepted-draft") != 1 || strings.Contains(writes, "must-not-leak") {
		t.Fatalf("lost/replayed input: %s", writes)
	}
}
func TestSessionOrderFailureAndOtherFailuresCancelPendingQuit(t *testing.T) {
	for _, kind := range []string{"metadata", "terminal", "preferences"} {
		t.Run(kind, func(t *testing.T) {
			s := newSessionOrderFixture()
			gate := make(chan struct{})
			t.Cleanup(func() { releaseGroupOrderGate(gate) })
			s.gate = gate
			if kind == "metadata" {
				s.orderErr = fmt.Errorf("synthetic uncertain save")
			}
			h := newWorkspaceHarness(t, s, Options{})
			h.until(func() bool { return h.m.workspace.conn != nil })
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			h.send(tea.KeyPressMsg{Code: tea.KeyDown})
			h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
			h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
			if kind == "terminal" {
				h.send(workspaceFrame{generation: h.m.workspace.generation, err: fmt.Errorf("synthetic disconnect")})
			}
			if kind == "preferences" {
				h.m.preferences.busy = true
				h.send(uiPreferencesSaved{generation: h.m.preferences.generation, err: fmt.Errorf("synthetic preferences failure")})
			}
			close(gate)
			h.until(func() bool { return h.m.sessionOrder == nil })
			if h.m.closed {
				t.Fatal("failure silently quit after order acknowledgement")
			}
			s.orderMu.Lock()
			defer s.orderMu.Unlock()
			if s.calls != 1 {
				t.Fatal("failed/uncertain operation retried")
			}
		})
	}
}
func TestSessionOrderLateResultCannotApplyToChangedRowsOrRestoreMissingRows(t *testing.T) {
	for _, kind := range []string{"missing", "run", "backend", "command", "name", "membership", "newer-position"} {
		t.Run(kind, func(t *testing.T) {
			m := New(context.Background(), newSessionOrderFixture())
			defer m.Close()
			m.sessions = sampleSessions()
			incoming := append([]session.Session(nil), m.sessions...)
			incoming[0].OrderRevision = 1
			incoming[0].Order = 1
			switch kind {
			case "missing":
				m.sessions = m.sessions[1:]
			case "run":
				m.sessions[0].RunID = "replacement"
			case "backend":
				m.sessions[0].BackendID = "replacement"
			case "command":
				m.sessions[0].Command = "replacement"
			case "name":
				m.sessions[0].TitleRevision++
			case "membership":
				m.sessions[0].GroupRevision++
			case "newer-position":
				m.sessions[0].OrderRevision = 2
			}
			before := append([]session.Session(nil), m.sessions...)
			m.acceptSessionOrder(incoming)
			if !session.SameOrderSnapshot(before, m.sessions) {
				t.Fatal("late order overwrote or recreated a changed row")
			}
		})
	}
}
func TestSessionOrderKeepsTerminalKeysAndBlocksMouseNewWhileSaving(t *testing.T) {
	s := newSessionOrderFixture()
	gate := make(chan struct{})
	t.Cleanup(func() { releaseGroupOrderGate(gate) })
	s.gate = gate
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	button := h.m.workspaceNewTabCell()
	h.send(tea.MouseClickMsg{Button: tea.MouseLeft, X: button.x, Y: 1})
	if h.m.launch != nil || h.m.workspace.transition != workspaceIdle {
		t.Fatal("mouse new bypassed metadata barrier")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'J', Text: "J"})
	close(gate)
	h.until(func() bool { return h.m.sessionOrder == nil && h.m.workspace.pending == 0 })
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	writes := strings.Join(stream.writes, ",")
	if strings.Count(writes, "text:K") != 1 || strings.Count(writes, "text:J") != 1 {
		t.Fatalf("terminal management keys lost/intercepted: %s", writes)
	}
}

func TestMembershipAcknowledgementResetsOldGroupPositionWithoutUndoingNewerSort(t *testing.T) {
	m := New(context.Background(), newSessionOrderFixture())
	defer m.Close()
	m.sessions = sampleSessions()
	m.sessions[0].Order = 3
	m.sessions[0].OrderRevision = 4
	m.workspace.tabs = []workspaceTab{{session: m.sessions[0]}}
	moved := m.sessions[0]
	moved.GroupID = "other"
	moved.GroupRevision = 1
	moved.Order, moved.OrderRevision = 0, 0
	m.acceptGroupMembership(moved)
	if m.sessions[0].OrderRevision != 0 || m.sessions[0].Order != 0 || m.workspace.tabs[0].session.OrderRevision != 0 {
		t.Fatal("move kept previous group's manual position")
	}
	m.sessions[0].Order, m.sessions[0].OrderRevision = 1, 2
	m.workspace.tabs[0].session = m.sessions[0]
	m.acceptGroupMembership(moved)
	if m.sessions[0].OrderRevision != 2 || m.workspace.tabs[0].session.OrderRevision != 2 {
		t.Fatal("late same-membership receipt undid a newer sort")
	}
}
