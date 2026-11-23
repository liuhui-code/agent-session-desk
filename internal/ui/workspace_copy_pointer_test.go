package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func (s *workspaceStream) Pointer(ctx context.Context, action string, x, y int) error {
	return s.record(ctx, fmt.Sprintf("pointer:%s,%d,%d", action, x, y))
}

func newCopyPointerHarness(t *testing.T) (*workspaceHarness, *workspaceFixture) {
	s := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}, frame: func(v session.Session, w, h int) terminal.Screen {
		return terminal.Screen{Width: w, Height: h, CopyMode: true, Rows: []terminal.Row{{Runs: []terminal.Run{{Text: "shell marker " + v.ID, Width: len("shell marker " + v.ID)}}}}}
	}}
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.screen.CopyMode })
	return h, s
}

func pointerWrites(s *workspaceFixture, id string) string {
	s.mu.Lock()
	stream := s.streams[id]
	s.mu.Unlock()
	if stream == nil {
		return ""
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	var result []string
	for _, v := range stream.writes {
		if strings.HasPrefix(v, "pointer:") {
			result = append(result, v)
		}
	}
	return strings.Join(result, ";")
}

func TestCopyPointerSendsContentCoordinatesInOrderWithoutKeyboardInput(t *testing.T) {
	h, s := newCopyPointerHarness(t)
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := pointerWrites(s, "first"); got != "pointer:down,3,2;pointer:drag,8,2;pointer:up,8,2" {
		t.Fatalf("native selection gesture missing or wrong: %q", got)
	}
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	for _, v := range stream.writes {
		if strings.HasPrefix(v, "text:") || strings.HasPrefix(v, "key:") {
			t.Fatalf("selection became keyboard input: %s", v)
		}
	}
}

func TestCopyPointerOwnsEdgeCrossingAndConsumesReleaseOnce(t *testing.T) {
	h, s := newCopyPointerHarness(t)
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: 1, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: 1, Button: tea.MouseNone})
	h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: 1, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	want := fmt.Sprintf("pointer:down,3,2;pointer:drag,8,%d;pointer:up,8,%d", 1-g.y, 1-g.y)
	if got := pointerWrites(s, "first"); got != want {
		t.Fatalf("edge/duplicate gesture: %q", got)
	}
	if !h.m.workspace.focused || h.m.workspace.active != 0 || h.m.workspace.split.dragging || h.m.workspace.tabDrag.viewID != 0 {
		t.Fatal("terminal drag became Tab/divider management")
	}
}

func TestCopyPointerDrainsOriginalBeforeSwitchWithoutContinuingOnNewRun(t *testing.T) {
	h, s := newCopyPointerHarness(t)
	h.until(func() bool { return h.m.workspace.pending == 0 })
	s.mu.Lock()
	first := s.streams["first"]
	s.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.KeyPressMsg{Code: 'z', Text: "z"})
	close(gate)
	h.until(func() bool {
		return h.m.workspace.active == 1 && h.m.workspace.conn != nil && h.m.workspace.pending == 0
	})
	if got := pointerWrites(s, "first"); got != "pointer:down,3,2;pointer:drag,8,2" {
		t.Fatalf("original FIFO: %q", got)
	}
	if got := pointerWrites(s, "second"); got != "" {
		t.Fatalf("new Run inherited gesture: %q", got)
	}
	s.mu.Lock()
	second := s.streams["second"]
	s.mu.Unlock()
	second.mu.Lock()
	defer second.mu.Unlock()
	if !strings.Contains(strings.Join(second.writes, ","), "text:z") {
		t.Fatal("new Run lost its independently typed text")
	}
}

func TestCopyPointerCancellationNeverSynthesizesRelease(t *testing.T) {
	for _, action := range []string{"resize", "key", "paste", "wheel", "no-button-motion", "modified-motion", "mode-exit", "prompt", "screen-size", "dialog"} {
		t.Run(action, func(t *testing.T) {
			h, s := newCopyPointerHarness(t)
			g := h.m.workspaceGeometry()
			h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
			switch action {
			case "resize":
				h.send(tea.WindowSizeMsg{Width: 94, Height: 25})
			case "key":
				h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
			case "paste":
				h.send(tea.PasteMsg{Content: "synthetic"})
			case "wheel":
				h.send(tea.MouseWheelMsg{X: g.x + 5, Y: g.y + 3, Button: tea.MouseWheelUp})
			case "no-button-motion":
				h.send(tea.MouseMotionMsg{X: g.x + 5, Y: g.y + 3, Button: tea.MouseNone})
			case "modified-motion":
				h.send(tea.MouseMotionMsg{X: g.x + 5, Y: g.y + 3, Button: tea.MouseLeft, Mod: tea.ModAlt})
			case "mode-exit", "prompt", "screen-size":
				frame := h.m.workspace.screen
				if action == "mode-exit" {
					frame.CopyMode = false
				} else if action == "screen-size" {
					frame.Width--
				} else {
					frame.CopyPrompt = &terminal.CopyPrompt{Text: "Search:", LabelBytes: 7, CursorBytes: 7}
				}
				h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
				if action == "screen-size" {
					frame.Width++
					h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
				}
			case "dialog":
				h.send(tea.KeyPressMsg{Code: tea.KeyF6})
				h.send(tea.KeyPressMsg{Code: '?', Text: "?"})
			}
			h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
			h.until(func() bool { return h.m.workspace.pending == 0 })
			if got := pointerWrites(s, "first"); got != "pointer:down,3,2" {
				t.Fatalf("cancelled gesture continued/copied: %q", got)
			}
		})
	}
}

func TestCopyPointerStartsOnlyInFocusedNativeContent(t *testing.T) {
	for _, mode := range []string{"management", "prompt", "border", "tab", "modified", "right-button"} {
		t.Run(mode, func(t *testing.T) {
			h, s := newCopyPointerHarness(t)
			g := h.m.workspaceGeometry()
			msg := tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft}
			switch mode {
			case "management":
				h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			case "prompt":
				h.m.workspace.screen.CopyPrompt = &terminal.CopyPrompt{Text: "Search:", LabelBytes: 7, CursorBytes: 7}
			case "border":
				msg.X = g.x - 1
			case "tab":
				msg.Y = 1
				msg.X = 0
			case "modified":
				msg.Mod = tea.ModShift
			case "right-button":
				msg.Button = tea.MouseRight
			}
			h.send(msg)
			h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
			h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
			h.until(func() bool { return h.m.workspace.pending == 0 })
			if got := pointerWrites(s, "first"); got != "" {
				t.Fatalf("non-terminal/non-mode gesture reached pane: %q", got)
			}
		})
	}
}
