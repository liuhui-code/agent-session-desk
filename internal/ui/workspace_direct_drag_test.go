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

func (s *workspaceStream) BeginCopyDrag(ctx context.Context, ax, ay, x, y int) error {
	return s.record(ctx, fmt.Sprintf("copy-drag:%d,%d,%d,%d", ax, ay, x, y))
}

func directDragFixture() *workspaceFixture {
	return &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}, frame: func(v session.Session, w, h int) terminal.Screen {
		return terminal.Screen{Width: w, Height: h, Rows: []terminal.Row{{Runs: []terminal.Run{{Text: "shell marker " + v.ID, Width: len("shell marker " + v.ID)}}}}}
	}}
}
func newDirectDragHarness(t *testing.T) (*workspaceHarness, *workspaceFixture) {
	s := directDragFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	return h, s
}

func TestDirectDragStartsNativeSelectionOnFirstMovementAndKeepsAnchor(t *testing.T) {
	h, s := newDirectDragHarness(t)
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := pointerWrites(s, "first"); got != "" {
		t.Fatalf("plain press/jitter sent pointer: %q", got)
	}
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 9, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: g.x + 10, Y: g.y + 2, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	got := strings.Join(stream.writes, ";")
	want := "copy-drag:3,2,8,2;pointer:drag,9,2;pointer:up,10,2"
	if got != want {
		t.Fatalf("direct selection did not hand off original anchor: %q", got)
	}
}

func directDragWrites(s *workspaceFixture, id string) string {
	s.mu.Lock()
	stream := s.streams[id]
	s.mu.Unlock()
	if stream == nil {
		return ""
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	return strings.Join(stream.writes, ";")
}

func TestDirectDragPlainClickOrSameCellJitterNeverStartsCopy(t *testing.T) {
	for _, offset := range []int{0, 7} {
		t.Run(fmt.Sprint(offset), func(t *testing.T) {
			h, s := newDirectDragHarness(t)
			g := h.m.workspaceGeometry()
			h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
			h.send(tea.MouseMotionMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
			h.send(tea.MouseReleaseMsg{X: g.x + 3 + offset, Y: g.y + 2, Button: tea.MouseLeft})
			h.until(func() bool { return h.m.workspace.pending == 0 })
			if got := directDragWrites(s, "first"); got != "" {
				t.Fatalf("click copied or wrote program input: %q", got)
			}
		})
	}
}

func TestDirectDragAcceptsDelayedModeEntryButNeverResumesAfterObservedExit(t *testing.T) {
	h, s := newDirectDragHarness(t)
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	frame := h.m.workspace.screen
	h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame}) // Earlier live frame may arrive after write acceptance.
	frame.CopyMode = true
	h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
	h.send(tea.MouseMotionMsg{X: g.x + 9, Y: g.y + 2, Button: tea.MouseLeft})
	frame.CopyMode = false
	h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
	frame.CopyMode = true
	h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
	h.send(tea.MouseReleaseMsg{X: g.x + 10, Y: g.y + 2, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := directDragWrites(s, "first"); got != "copy-drag:3,2,8,2;pointer:drag,9,2" {
		t.Fatalf("delayed/ended mode ownership: %q", got)
	}
}

func TestDirectDragChangedModeOrMouseRequestCancelsUnstartedAnchor(t *testing.T) {
	for _, change := range []string{"copy-mode", "mouse-request", "prompt", "screen-size"} {
		t.Run(change, func(t *testing.T) {
			h, s := newDirectDragHarness(t)
			g := h.m.workspaceGeometry()
			h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
			original := h.m.workspace.screen
			frame := original
			switch change {
			case "copy-mode":
				frame.CopyMode = true
			case "mouse-request":
				frame.WantsMouse = true
			case "prompt":
				frame.CopyPrompt = &terminal.CopyPrompt{Text: "Search:", LabelBytes: 7, CursorBytes: 7}
			case "screen-size":
				frame.Width--
			}
			h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
			h.send(workspaceFrame{generation: h.m.workspace.generation, screen: original})
			h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
			h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
			h.until(func() bool { return h.m.workspace.pending == 0 })
			if got := directDragWrites(s, "first"); got != "" {
				t.Fatalf("changed then restored state resumed anchor: %q", got)
			}
		})
	}
}

func TestDirectDragDrainsOriginalBeforeSwitchWithoutReleaseToNewRun(t *testing.T) {
	h, s := newDirectDragHarness(t)
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
	h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.KeyPressMsg{Code: 'z', Text: "z"})
	close(gate)
	h.until(func() bool {
		return h.m.workspace.active == 1 && h.m.workspace.conn != nil && h.m.workspace.pending == 0
	})
	h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	if got := directDragWrites(s, "first"); got != "copy-drag:3,2,8,2" {
		t.Fatalf("original queue: %q", got)
	}
	if got := directDragWrites(s, "second"); got != "text:z" {
		t.Fatalf("new Run received old pointer or lost its own input: %q", got)
	}
}

type pointerOnlyConnection struct{ terminal.Connection }

func (c pointerOnlyConnection) Pointer(ctx context.Context, action string, x, y int) error {
	return c.Connection.(terminal.PointerConnection).Pointer(ctx, action, x, y)
}

type withoutDirectDragService struct{ *workspaceFixture }

func (s withoutDirectDragService) OpenTerminal(ctx context.Context, v session.Session, w, h int) (terminal.Connection, error) {
	c, e := s.workspaceFixture.OpenTerminal(ctx, v, w, h)
	if e != nil {
		return nil, e
	}
	return pointerOnlyConnection{c}, nil
}

func TestDirectDragUnsupportedConnectionKeepsExistingNativePointerUsable(t *testing.T) {
	base := directDragFixture()
	h := newWorkspaceHarness(t, withoutDirectDragService{base})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := directDragWrites(base, "first"); got != "" {
		t.Fatalf("unsupported capability fell back to input: %q", got)
	}
	frame := h.m.workspace.screen
	frame.CopyMode = true
	h.send(workspaceFrame{generation: h.m.workspace.generation, screen: frame})
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := directDragWrites(base, "first"); got != "pointer:down,3,2;pointer:drag,8,2;pointer:up,8,2" {
		t.Fatalf("existing native pointer lost: %q", got)
	}
}
