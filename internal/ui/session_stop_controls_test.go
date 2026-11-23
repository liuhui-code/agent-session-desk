package ui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func newStopFixture() *sessionStopFixture {
	return &sessionStopFixture{workspaceFixture: &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}}
}
func openStop(t *testing.T, h *workspaceHarness) {
	t.Helper()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	h.until(func() bool { return h.m.sessionStop != nil })
}

func TestSessionStopDrainsDraftAndNeverLeaksModalKeysOrPaste(t *testing.T) {
	s := newStopFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil && h.m.workspace.pending == 0 })
	s.mu.Lock()
	original := s.streams["first"]
	s.mu.Unlock()
	gate := make(chan struct{})
	original.mu.Lock()
	original.gate = gate
	original.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "draft-dyn"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	if h.m.sessionStop != nil || h.m.workspace.transition != workspaceStop || s.count() != 0 {
		t.Fatal("stop bypassed pending original input")
	}
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	close(gate)
	h.until(func() bool { return h.m.sessionStop != nil })
	if s.count() != 0 {
		t.Fatal("confirmation was replayed after input drain")
	}
	h.send(tea.PasteMsg{Content: "y\n"})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.send(tea.MouseClickMsg{})
	if s.count() != 0 || h.m.workspace.active != 0 || h.m.View().Cursor != nil {
		t.Fatal("modal lost input authority")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return h.m.sessionStop == nil && h.m.workspace.conn != nil })
	original.mu.Lock()
	defer original.mu.Unlock()
	if strings.Count(strings.Join(original.writes, ","), "text:draft-dyn") != 1 {
		t.Fatal("original draft lost/replayed")
	}
	for _, v := range original.writes {
		if strings.Contains(v, "text:y") || strings.HasPrefix(v, "paste:") {
			t.Fatal("modal input leaked")
		}
	}
}

func TestSessionStopBusyAndUncertainResultCannotReplayOrQuit(t *testing.T) {
	gate := make(chan struct{})
	s := newStopFixture()
	s.stopGate, s.stopErr = gate, errors.New("backend may have stopped; metadata unavailable")
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	openStop(t, h)
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.until(func() bool { return s.count() == 1 })
	for _, k := range []tea.KeyPressMsg{{Code: 'y', Text: "y"}, {Code: tea.KeyEnter}, {Code: tea.KeyEsc}, {Code: 'q', Text: "q"}, {Code: 'c', Mod: tea.ModCtrl}, {Code: tea.KeyF6}} {
		h.send(k)
	}
	if h.m.closed || h.m.sessionStop == nil || !h.m.sessionStop.busy {
		t.Fatal("busy stop discarded")
	}
	close(gate)
	h.until(func() bool { return h.m.sessionStop != nil && h.m.sessionStop.err != nil && !h.m.sessionStop.busy })
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if s.count() != 1 || h.m.sessionStop == nil || h.m.workspace.tabs[0].session.TerminalState == session.TerminalExited {
		t.Fatal("uncertain stop retried or claimed success")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEsc})
	h.until(func() bool { return h.m.sessionStop == nil && h.m.workspace.conn != nil })
}

func TestSessionStopUnreadableReviewNeverConfirms(t *testing.T) {
	s := newStopFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	openStop(t, h)
	for _, size := range [][2]int{{120, 10}, {63, 10}, {49, 8}, {27, 8}, {9, 3}, {1, 1}, {0, 0}} {
		h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		view := h.m.View()
		if size[0] == 0 {
			if view.Content != "" {
				t.Fatal("zero view not empty")
			}
		} else {
			lines := strings.Split(ansi.Strip(view.Content), "\n")
			if len(lines) != size[1] {
				t.Fatal("wrong frame height")
			}
			for _, line := range lines {
				if ansi.StringWidth(line) != size[0] {
					t.Fatal("wrong frame width")
				}
			}
		}
		h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
		if h.m.sessionStop.submitted || s.count() != 0 {
			t.Fatal("hidden review confirmed")
		}
	}
	h.send(tea.WindowSizeMsg{Width: 120, Height: 32})
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return h.m.sessionStop == nil && s.count() == 1 })
	if len(h.m.workspace.tabs) != 2 || h.m.workspace.tabs[0].live || h.m.workspace.tabs[0].session.TerminalState != session.TerminalExited || h.m.workspace.conn != nil {
		t.Fatal("stopped active view reopened or disappeared")
	}
}

func TestSessionStopLateReceiptCannotTouchReplacementRun(t *testing.T) {
	s := newStopFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	openStop(t, h)
	old := h.m.sessionStop
	replacement := sampleSessions()[0]
	replacement.RunID = "replacement"
	replacement.BackendID = "new-backend"
	h.m.sessions[0], h.m.workspace.tabs[0].session = replacement, replacement
	h.m.sessionStop.busy, h.m.sessionStop.submitted = true, true
	h.send(sessionStopResult{old, nil})
	if !sameStopTarget(h.m.workspace.tabs[0].session, replacement) || !h.m.workspace.tabs[0].live || h.m.sessions[0].RunID != replacement.RunID {
		t.Fatal("old success retired replacement")
	}
	next := &sessionStopDialog{target: replacement, focus: 1}
	h.m.sessionStop = next
	h.send(sessionStopResult{old, errors.New("late")})
	if h.m.sessionStop != next || next.err != nil {
		t.Fatal("late receipt changed new review")
	}
}

func TestSessionStopUnknownAndUnsupportedServiceDoNotOpenDialog(t *testing.T) {
	s := newStopFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.m.sessions[0].TerminalState = session.TerminalUnknown
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	if h.m.sessionStop != nil || h.m.sessionStopIntent != nil || s.count() != 0 {
		t.Fatal("unknown target accepted")
	}
}

func TestSessionStopInputFailureCancelsUnopenedReview(t *testing.T) {
	s := newStopFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil && h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.m.workspace.pending = 1
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	if h.m.sessionStopIntent == nil {
		t.Fatal("no identity-bound pending review")
	}
	h.send(workspaceInputDone{generation: h.m.workspace.generation, err: errors.New("uncertain input")})
	if h.m.sessionStopIntent != nil || h.m.sessionStop != nil || h.m.workspace.transition != workspaceIdle || s.count() != 0 {
		t.Fatal("input failure retained stop authority")
	}
}

func TestSessionStopServiceCapabilityRequired(t *testing.T) {
	s := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	if h.m.sessionStopIntent != nil || h.m.sessionStop != nil {
		t.Fatal("unsupported stop service exposed confirmation")
	}
}

func TestSessionStopCannotConfirmSilentlyTruncatedTargetFields(t *testing.T) {
	for _, field := range []string{"title", "program"} {
		t.Run(field, func(t *testing.T) {
			m := New(context.Background(), newStopFixture())
			defer m.Close()
			m.width, m.height = 120, 500
			target := sampleSessions()[0]
			if field == "title" {
				target.Title = strings.Repeat("a", maxLabelBytes+1)
			} else {
				target.Command = strings.Repeat("a", maxLabelBytes+1)
			}
			m.sessionStop = &sessionStopDialog{target: target, focus: 0}
			if cmd := m.submitSessionStop(); cmd != nil || m.sessionStop.submitted {
				t.Fatal("bounded display silently hid part of reviewed target")
			}
		})
	}
}
