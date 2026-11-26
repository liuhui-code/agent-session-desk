package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type attentionFixture struct {
	workspaceFixture
	item     session.AttentionItem
	openErr  error
	openGate <-chan struct{}
	visits   int
}

func (s *attentionFixture) OpenAttention(ctx context.Context, id string, version int64) (session.AttentionItem, error) {
	s.mu.Lock()
	s.visits++
	item, err, gate := s.item, s.openErr, s.openGate
	s.mu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return item, ctx.Err()
		}
	}
	if id != item.ID || version != item.Version {
		return item, fmt.Errorf("changed")
	}
	now := time.Now()
	item.ViewedAt = &now
	return item, err
}
func attentionUIFixture(t *testing.T) (*workspaceHarness, *attentionFixture) {
	s := &attentionFixture{}
	s.fakeService.list = func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }
	s.item = session.AttentionItem{AttentionEvent: session.AttentionEvent{ID: strings.Repeat("a", 32), Version: 1, Kind: "approval", State: "pending", Title: "Allow operation?"}, Source: sampleSessions()[1], TerminalAvailable: true}
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	return h, s
}
func TestAttentionReminderPreservesGeometryFocusAndInput(t *testing.T) {
	h, s := attentionUIFixture(t)
	before := h.m.workspaceGeometry()
	active := h.m.workspace.active
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	if h.m.workspaceGeometry() != before || h.m.workspace.active != active || !h.m.workspace.focused {
		t.Fatal("reminder changed terminal geometry/focus")
	}
	if !strings.Contains(ansi.Strip(h.m.View().Content), "1 need attention") {
		t.Fatal("reminder missing")
	}
	if !h.m.sessionNeedsAttention(sampleSessions()[1]) {
		t.Fatal("session marker missing")
	}
	s.mu.Lock()
	visits := s.visits
	stream := s.streams["first"]
	s.mu.Unlock()
	stream.mu.Lock()
	writes := append([]string(nil), stream.writes...)
	stream.mu.Unlock()
	if visits != 0 || len(writes) != 0 {
		t.Fatalf("passive reminder acted on terminal: %d %v", visits, writes)
	}
}
func TestAttentionShortcutNavigatesWithoutAnsweringAndViewedStaysPending(t *testing.T) {
	h, s := attentionUIFixture(t)
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	h.send(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	h.until(func() bool { return h.m.workspace.active == 1 && h.m.workspace.conn != nil && !h.m.attention.busy })
	if len(h.m.attention.items) != 1 || h.m.attention.items[0].ViewedAt == nil {
		t.Fatal("view receipt lost or resolved item")
	}
	s.mu.Lock()
	stream := s.streams[s.item.Source.ID]
	s.mu.Unlock()
	stream.mu.Lock()
	writes := append([]string(nil), stream.writes...)
	stream.mu.Unlock()
	if len(writes) != 0 {
		t.Fatalf("navigation answered form: %v", writes)
	}
}

func TestAttentionDrainsAcceptedInputAndConsumesLaterPaste(t *testing.T) {
	h, s := attentionUIFixture(t)
	s.mu.Lock()
	first := s.streams["first"]
	s.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	h.send(tea.KeyPressMsg{Code: 'a', Text: "accepted draft"})
	h.send(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	h.send(tea.PasteMsg{Content: "y\n"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	close(gate)
	h.until(func() bool { return h.m.workspace.active == 1 && h.m.workspace.conn != nil && !h.m.attention.busy })
	first.mu.Lock()
	writes := strings.Join(first.writes, ",")
	first.mu.Unlock()
	if writes != "text:accepted draft" {
		t.Fatalf("navigation accepted later input: %q", writes)
	}
}

func TestAttentionLateResolvedResultAndUnavailableObservationDoNotNavigate(t *testing.T) {
	h, s := attentionUIFixture(t)
	gate := make(chan struct{})
	s.openGate = gate
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	h.send(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	if !h.m.attention.busy {
		t.Fatal("lookup did not start")
	}
	h.send(sessionsMsg{sessions: sampleSessions()})
	close(gate)
	h.until(func() bool { return !h.m.attention.busy && h.m.workspace.conn != nil })
	if h.m.workspace.active != 0 {
		t.Fatal("late result opened resolved request")
	}
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	h.m.listErr = fmt.Errorf("disconnected")
	h.send(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	if h.m.attention.busy || !strings.Contains(ansi.Strip(h.m.View().Content), "last observed") {
		t.Fatal("unknown observation claimed freshness")
	}
}

func TestAttentionHeaderClickAndRenderingBoundaries(t *testing.T) {
	for _, size := range [][2]int{{160, 40}, {120, 32}, {63, 20}, {49, 12}, {12, 4}, {2, 1}} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			h, s := attentionUIFixture(t)
			h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
			h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
			frame := ansi.Strip(h.m.View().Content)
			if lines := strings.Split(frame, "\n"); len(lines) > size[1] {
				t.Fatal("frame overflow")
			}
			for _, line := range strings.Split(frame, "\n") {
				if ansi.StringWidth(line) > size[0] {
					t.Fatalf("row overflow %q", line)
				}
			}
			text, x, y := h.m.attentionHeader()
			if text != "" {
				h.send(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
				if !h.m.attention.busy {
					t.Fatal("reminder not clickable")
				}
			}
		})
	}
}

func TestAttentionHeaderClickWorksWithSidebarHidden(t *testing.T) {
	h, s := attentionUIFixture(t)
	h.m.workspace.sidebarHidden = true
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	text, x, y := h.m.attentionHeader()
	if text == "" {
		t.Fatal("hidden sidebar lost reminder")
	}
	h.send(tea.MouseClickMsg{Button: tea.MouseLeft, X: x, Y: y})
	h.until(func() bool { return h.m.workspace.active == 1 && h.m.workspace.conn != nil && !h.m.attention.busy })
	if !h.m.workspace.sidebarHidden || !h.m.workspace.focused {
		t.Fatal("reminder changed sidebar preference or failed to focus original form")
	}
}

func TestAttentionHeaderClickReleasesOriginalProgramPointer(t *testing.T) {
	h, s := attentionUIFixture(t)
	h.until(func() bool { return h.m.workspace.screen.Width > 0 })
	frame := h.m.workspace.screen
	frame.WantsMouse = true
	h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 1, Y: g.y, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if h.m.workspace.programPointer == nil {
		t.Fatal("program gesture did not start")
	}
	_, x, y := h.m.attentionHeader()
	h.send(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.active == 1 && h.m.workspace.conn != nil })
	s.mu.Lock()
	first := s.streams["first"]
	s.mu.Unlock()
	first.mu.Lock()
	writes := strings.Join(first.writes, ";")
	first.mu.Unlock()
	if writes != "pointer:down,1,0;pointer:up,1,0" {
		t.Fatalf("reminder lost original program release: %s", writes)
	}
}
