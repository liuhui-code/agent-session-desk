package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestProgramMouseKeepsOneReleaseSlotUnderMotionPressure(t *testing.T) {
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
	h.until(func() bool { return len(h.m.workspace.inputs) == 0 }) // Writer owns the blocked down.
	for i := range 200 {
		h.send(tea.MouseMotionMsg{X: g.x + 1 + i%10, Y: g.y + 2, Button: tea.MouseLeft})
	}
	if got := len(h.m.workspace.inputs); got != workspaceInputCapacity-1 {
		t.Fatalf("no bounded release reserve: %d", got)
	}
	if !strings.Contains(h.m.status, "release position") {
		t.Fatal("rejected motion was not reported")
	}
	h.send(tea.MouseReleaseMsg{X: g.x + 9, Y: g.y + 2, Button: tea.MouseLeft})
	if len(h.m.workspace.inputs) != workspaceInputCapacity || h.m.workspace.programPointer != nil {
		t.Fatal("release could not use reserved slot")
	}
	close(gate)
	h.until(func() bool { return h.m.workspace.pending == 0 })
	got := directDragWrites(s, "first")
	if strings.Count(got, "pointer:") != workspaceInputCapacity+1 || !strings.HasSuffix(got, "pointer:up,9,2") || strings.Count(got, "pointer:up") != 1 {
		t.Fatalf("bounded queue lost/repeated release: %q", got)
	}
}

func TestProgramMouseCrossesControlsOnlyWithinOriginalPaneCoordinates(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	w, hgt := h.m.workspace.screen.Width, h.m.workspace.screen.Height
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: h.m.width + 20, Y: h.m.height + 20, Button: tea.MouseNone})
	h.send(tea.MouseReleaseMsg{X: h.m.width + 20, Y: h.m.height + 20, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	want := fmt.Sprintf("pointer:down,3,2;pointer:drag,0,0;pointer:up,%d,%d", min(g.width, w)-1, min(g.height, hgt)-1)
	if got := directDragWrites(s, "first"); got != want {
		t.Fatalf("cross-control program gesture: %q", got)
	}
	if h.m.workspace.active != 0 || !h.m.workspace.focused || h.m.workspace.split.dragging {
		t.Fatal("program drag became manager navigation")
	}
}

func TestProgramMouseDeliberateActionsReleaseBeforeTheirOwnInput(t *testing.T) {
	for _, action := range []string{"key", "paste", "wheel", "resize", "focus", "help", "no-button", "modified-motion"} {
		t.Run(action, func(t *testing.T) {
			h, s := newProgramMouseHarness(t)
			g := h.m.workspaceGeometry()
			h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
			h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
			switch action {
			case "key":
				h.send(tea.KeyPressMsg{Code: 'a', Text: "a"})
			case "paste":
				h.send(tea.PasteMsg{Content: "synthetic"})
			case "wheel":
				h.send(tea.MouseWheelMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseWheelUp})
			case "resize":
				h.send(tea.WindowSizeMsg{Width: 94, Height: 25})
			case "focus":
				h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			case "help":
				h.send(tea.KeyPressMsg{Code: tea.KeyF6})
				h.send(tea.KeyPressMsg{Code: '?', Text: "?"})
			case "no-button":
				h.send(tea.MouseMotionMsg{X: g.x + 20, Y: g.y + 2, Button: tea.MouseNone})
			case "modified-motion":
				h.send(tea.MouseMotionMsg{X: g.x + 20, Y: g.y + 2, Button: tea.MouseLeft, Mod: tea.ModSuper})
			}
			h.until(func() bool { return h.m.workspace.pending == 0 })
			if action == "help" {
				h.until(func() bool { return h.m.help })
			}
			h.send(tea.MouseReleaseMsg{X: g.x + 20, Y: g.y + 2, Button: tea.MouseLeft})
			got := directDragWrites(s, "first")
			prefix := "pointer:down,3,2;pointer:drag,8,2;pointer:up,8,2"
			if !strings.HasPrefix(got, prefix) || strings.Count(got, "pointer:up") != 1 {
				t.Fatalf("lost/repeated cleanup: %q", got)
			}
			wantAfter := map[string]string{"key": ";text:a", "paste": ";paste:synthetic", "wheel": ";scroll:", "resize": ";size:", "focus": ";size:", "help": ";size:"}[action]
			if !strings.HasPrefix(strings.TrimPrefix(got, prefix), wantAfter) {
				t.Fatalf("release was not before deliberate input: %q", got)
			}
		})
	}
}

func TestProgramMouseExternalModeChangesDoNotReleaseIntoAnotherMode(t *testing.T) {
	for _, change := range []string{"copy-mode", "mouse-off", "prompt", "height"} {
		t.Run(change, func(t *testing.T) {
			h, s := newProgramMouseHarness(t)
			g := h.m.workspaceGeometry()
			h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
			original := h.m.workspace.screen
			frame := original
			switch change {
			case "copy-mode":
				frame.CopyMode = true
			case "mouse-off":
				frame.WantsMouse = false
			case "prompt":
				frame.CopyPrompt = &terminal.CopyPrompt{Text: "Search:", LabelBytes: 7, CursorBytes: 7}
			case "height":
				frame.Height++
			}
			h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
			h.send(workspaceFrame{generation: h.m.workspace.generation, screen: original})
			h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
			h.until(func() bool { return h.m.workspace.pending == 0 })
			if got := directDragWrites(s, "first"); got != "pointer:down,3,2" {
				t.Fatalf("unknown hold replayed/released in new mode: %q", got)
			}
			if !strings.Contains(h.m.status, "unknown") {
				t.Fatal("release uncertainty missing")
			}
		})
	}
}

func TestProgramMouseStartsOnlyInUnmodifiedFocusedProgramContent(t *testing.T) {
	for _, kind := range []string{"management", "border", "tab", "modified", "prompt", "right"} {
		t.Run(kind, func(t *testing.T) {
			h, s := newProgramMouseHarness(t)
			g := h.m.workspaceGeometry()
			click := tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft}
			switch kind {
			case "management":
				h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			case "border":
				click.X = g.x - 1
			case "tab":
				click.X = 0
				click.Y = 1
			case "modified":
				click.Mod = tea.ModSuper
			case "prompt":
				h.m.workspace.screen.CopyPrompt = &terminal.CopyPrompt{Text: "Search:", LabelBytes: 7, CursorBytes: 7}
			case "right":
				click.Button = tea.MouseBackward
			}
			h.send(click)
			h.send(tea.MouseReleaseMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
			h.until(func() bool { return h.m.workspace.pending == 0 })
			if got := pointerWrites(s, "first"); got != "" {
				t.Fatalf("non-content/non-owner program input: %q", got)
			}
		})
	}
}

func TestProgramMouseCopyModeKeepsSelectionCancellationWithoutSyntheticRelease(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	frame := h.m.workspace.screen
	frame.CopyMode = true
	h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := directDragWrites(s, "first"); got != "pointer:down,3,2;key:Left" {
		t.Fatalf("copy selection was mistaken for program hold: %q", got)
	}
}
