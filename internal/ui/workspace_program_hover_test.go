package ui

import (
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestProgramHoverCoalescesBlockedMotionAndSendsLatestCoordinate(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	s.mu.Lock()
	first := s.streams["first"]
	s.mu.Unlock()
	gate := make(chan struct{})
	var closeOnce sync.Once
	release := func() { closeOnce.Do(func() { close(gate) }) }
	t.Cleanup(release)
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(tea.MouseMotionMsg{X: g.x + 1, Y: g.y + 2, Button: tea.MouseNone})
	if h.m.workspace.pending != 1 {
		t.Fatal("bare program motion was not accepted")
	}
	for i := 0; i < 200; i++ {
		h.send(tea.MouseMotionMsg{X: g.x + 2 + i%10, Y: g.y + 2, Button: tea.MouseNone})
	}
	if h.m.workspace.pending != 1 || len(h.m.workspace.inputs) > 1 {
		t.Fatal("hover created an input backlog")
	}
	release()
	h.until(func() bool {
		return h.m.workspace.pending == 0 && strings.HasSuffix(directDragWrites(s, "first"), "mouse:move,none,0,11,2")
	})
	if got := directDragWrites(s, "first"); got != "mouse:move,none,0,1,2;mouse:move,none,0,11,2" {
		t.Fatalf("hover did not coalesce to first+latest: %q", got)
	}
}

func TestUnsupportedHoverCancelsTheQueuedUnmodifiedPosition(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	s.mu.Lock()
	first := s.streams["first"]
	s.mu.Unlock()
	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	t.Cleanup(release)
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(tea.MouseMotionMsg{X: g.x + 1, Y: g.y + 2, Button: tea.MouseNone})
	h.send(tea.MouseMotionMsg{X: g.x + 5, Y: g.y + 2, Button: tea.MouseNone})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseNone, Mod: tea.ModShift})
	release()
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := directDragWrites(s, "first"); got != "mouse:move,none,0,1,2" {
		t.Fatalf("stale pending hover survived unsupported event: %q", got)
	}
}

func TestProgramHoverDiscardsPendingPositionAcrossModeGeometryAndNavigation(t *testing.T) {
	for _, change := range []string{"copy-mode", "mouse-off", "width", "outside", "key", "switch"} {
		t.Run(change, func(t *testing.T) {
			h, s := newProgramMouseHarness(t)
			g := h.m.workspaceGeometry()
			s.mu.Lock()
			first := s.streams["first"]
			s.mu.Unlock()
			gate := make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(gate) }) }
			t.Cleanup(release)
			first.mu.Lock()
			first.gate = gate
			first.mu.Unlock()
			h.send(tea.MouseMotionMsg{X: g.x + 1, Y: g.y + 2, Button: tea.MouseNone})
			h.send(tea.MouseMotionMsg{X: g.x + 5, Y: g.y + 2, Button: tea.MouseNone})
			switch change {
			case "copy-mode", "mouse-off", "width":
				original := h.m.workspace.screen
				frame := original
				if change == "copy-mode" {
					frame.CopyMode = true
				} else if change == "mouse-off" {
					frame.WantsMouse = false
				} else {
					frame.Width++
				}
				h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
				h.send(workspaceFrame{generation: h.m.workspace.generation, screen: original})
			case "outside":
				h.send(tea.MouseMotionMsg{X: 0, Y: 1, Button: tea.MouseNone})
			case "key":
				h.send(tea.KeyPressMsg{Code: 'x', Text: "x"})
			case "switch":
				h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
				h.send(tea.KeyPressMsg{Code: 'z', Text: "z"})
			}
			release()
			h.until(func() bool {
				return h.m.workspace.pending == 0 && (change != "switch" || (h.m.workspace.active == 1 && h.m.workspace.conn != nil))
			})
			want := "mouse:move,none,0,1,2"
			if change == "key" {
				want += ";text:x"
			}
			if got := directDragWrites(s, "first"); got != want {
				t.Fatalf("invalid pending hover sent: %q", got)
			}
			if change == "switch" && directDragWrites(s, "second") != "text:z" {
				t.Fatal("new Run received hover or lost its own input")
			}
		})
	}
}
func TestProgramHoverDuplicateOldReceiptCannotAcknowledgeNewConnection(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	h.send(tea.MouseMotionMsg{X: g.x + 1, Y: g.y + 2, Button: tea.MouseNone})
	old := h.m.workspace.programHover.running
	generation := h.m.workspace.generation
	h.until(func() bool { return h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.until(func() bool {
		return h.m.workspace.active == 1 && h.m.workspace.conn != nil && h.m.workspace.screen.WantsMouse
	})
	s.mu.Lock()
	second := s.streams["second"]
	s.mu.Unlock()
	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	t.Cleanup(release)
	second.mu.Lock()
	second.gate = gate
	second.mu.Unlock()
	g = h.m.workspaceGeometry()
	h.send(tea.MouseMotionMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseNone})
	current := h.m.workspace.programHover.running
	h.send(workspaceProgramHoverDone{done: workspaceInputDone{generation: generation}, target: old})
	if h.m.workspace.pending != 1 || h.m.workspace.programHover.running != current {
		t.Fatal("stale receipt released a new connection's hover")
	}
	release()
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := directDragWrites(s, "second"); got != "mouse:move,none,0,3,2" {
		t.Fatal(got)
	}
}
