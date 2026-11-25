package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func programMouseFixture() *workspaceFixture {
	s := directDragFixture()
	frame := s.frame
	s.frame = func(v session.Session, w, h int) terminal.Screen { f := frame(v, w, h); f.WantsMouse = true; return f }
	return s
}
func newProgramMouseHarness(t *testing.T) (*workspaceHarness, *workspaceFixture) {
	s := programMouseFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.screen.WantsMouse })
	return h, s
}

func TestProgramMouseLeftGestureReachesOriginalProgramWithoutSelectionOrKeys(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: g.x + 9, Y: g.y + 2, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	got := directDragWrites(s, "first")
	if got != "pointer:down,3,2;pointer:drag,8,2;pointer:up,9,2" {
		t.Fatalf("program did not receive its native mouse gesture: %q", got)
	}
	if strings.Contains(got, "copy-drag") || strings.Contains(got, "key:") || strings.Contains(got, "text:") {
		t.Fatal("program mouse became selection or keyboard input")
	}
}

func TestProgramMouseSwitchReleasesOriginalBeforeNewRunInput(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	s.mu.Lock()
	first := s.streams["first"]
	s.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: 'z', Text: "z"})
	close(gate)
	h.until(func() bool {
		return h.m.workspace.active == 1 && h.m.workspace.conn != nil && h.m.workspace.pending == 0
	})
	h.send(tea.MouseReleaseMsg{X: g.x + 9, Y: g.y + 2, Button: tea.MouseLeft})
	if got := directDragWrites(s, "first"); got != "pointer:down,3,2;pointer:drag,8,2;pointer:up,8,2" {
		t.Fatalf("old Run not released before switch: %q", got)
	}
	if got := directDragWrites(s, "second"); got != "text:z" {
		t.Fatalf("new Run inherited program pointer: %q", got)
	}
}

func TestProgramMouseExternalScreenGrowthThenRestoreCannotResumeGesture(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	original := h.m.workspace.screen
	grown := original
	grown.Width++
	h.send(workspaceFrame{generation: h.m.workspace.generation, screen: grown})
	h.send(workspaceFrame{generation: h.m.workspace.generation, screen: original})
	h.send(tea.MouseReleaseMsg{X: g.x + 9, Y: g.y + 2, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := directDragWrites(s, "first"); got != "pointer:down,3,2" {
		t.Fatalf("unknown external geometry resumed/released old gesture: %q", got)
	}
	if !strings.Contains(h.m.status, "Unknown") {
		t.Fatal("external change lost uncertain release notice")
	}
}
