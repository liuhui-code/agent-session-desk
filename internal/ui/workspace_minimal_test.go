package ui

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestMinimalWorkspaceUsesAllRowsWithoutDuplicateChrome(t *testing.T) {
	h := newWorkspaceHarness(t, &workspaceFixture{fakeService: directDragFixture().fakeService})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker") })
	view := ansi.Strip(h.m.View().Content)
	for _, unwanted := range []string{"Agent Session Desk", "╭", "╰", "Terminal", "Ctrl+PgUp"} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("unrequested chrome remains: %q", unwanted)
		}
	}
	g := h.m.workspaceGeometry()
	if strings.Count(view, h.m.workspace.tabs[0].session.Title) != 1 {
		t.Fatal("session title is duplicated outside the sidebar")
	}
	if g.y != 1 || g.height != h.m.height-1 || g.width != g.outerWidth {
		t.Fatalf("content still reserves header/footer/border: %+v", g)
	}
	if h.m.View().Cursor == nil || h.m.View().Cursor.Y != 1 {
		t.Fatal("native cursor did not follow borderless content")
	}
}

func TestQuietFooterRestoresLastRowAndCursorAndKeepsCriticalFeedback(t *testing.T) {
	h := newWorkspaceHarness(t, &workspaceFixture{fakeService: directDragFixture().fakeService})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker") })
	w := h.m.workspace
	w.screen.Rows = make([]terminal.Row, h.m.workspaceGeometry().height)
	w.screen.Rows[h.m.workspaceGeometry().height-1] = terminal.Row{Runs: []terminal.Run{{Text: "LAST_ROW", Width: 8}}}
	w.screen.Cursor.Y = h.m.workspaceGeometry().height - 1
	h.m.status = "Save failed"
	h.send(tea.MouseMotionMsg{X: h.m.width - 1, Y: h.m.height - 1, Button: tea.MouseNone})
	if strings.Contains(h.m.View().Content, "LAST_ROW") || h.m.View().Cursor != nil {
		t.Fatal("overlay did not own the last row and cursor")
	}
	h.m.status = ""
	h.send(tea.MouseMotionMsg{X: h.m.width - 1, Y: h.m.height - 2, Button: tea.MouseNone})
	if !strings.Contains(h.m.View().Content, "LAST_ROW") || h.m.View().Cursor == nil || h.m.View().Cursor.Y != h.m.height-1 {
		t.Fatal("leave did not restore original output/cursor")
	}
	h.m.status = "保存失败；未覆盖原配置"
	if !strings.Contains(h.m.View().Content, h.m.status) {
		t.Fatal("critical feedback was hidden with ordinary hints")
	}
}

func TestQuietFooterConsumesItsMouseSurfaceButPreservesNativeDragRelease(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	h.m.status = "Save failed"
	h.send(tea.MouseMotionMsg{X: g.x + 3, Y: h.m.height - 1, Button: tea.MouseNone})
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: h.m.height - 1, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: g.x + 3, Y: h.m.height - 1, Button: tea.MouseLeft})
	h.send(tea.MouseWheelMsg{X: g.x + 3, Y: h.m.height - 1, Button: tea.MouseWheelDown})
	if got := directDragWrites(s, "first"); got != "" {
		t.Fatalf("overlay leaked mouse events: %q", got)
	}
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: h.m.height - 2, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: g.x + 4, Y: h.m.height - 1, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: g.x + 4, Y: h.m.height - 1, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := directDragWrites(s, "first"); got != fmt.Sprintf("pointer:down,3,%d;pointer:drag,4,%d;pointer:up,4,%d", h.m.height-2-g.y, h.m.height-1-g.y, h.m.height-1-g.y) {
		t.Fatalf("bottom edge stole a native drag/release: %q", got)
	}
}

func TestQuietFooterHoverDoesNotResizeOrWriteAndTypingHidesIt(t *testing.T) {
	h := newWorkspaceHarness(t, &workspaceFixture{fakeService: directDragFixture().fakeService})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker") && h.m.workspace.pending == 0 })
	g := h.m.workspaceGeometry()
	stream := h.m.workspace.conn.(*workspaceStream)
	readWrites := func() []string {
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return append([]string(nil), stream.writes...)
	}
	before := readWrites()
	h.send(tea.MouseMotionMsg{X: g.x + 1, Y: h.m.height - 1, Button: tea.MouseNone})
	if strings.Contains(ansi.Strip(h.m.View().Content), "F6 Navigate") {
		t.Fatal("bottom-edge hover revealed a removed mode hint")
	}
	if h.m.workspaceGeometry() != g || !reflect.DeepEqual(readWrites(), before) {
		t.Fatal("revealing the footer resized or wrote to the terminal")
	}
	h.send(tea.MouseMotionMsg{X: g.x + 1, Y: h.m.height - 2, Button: tea.MouseNone})
	if strings.Contains(ansi.Strip(h.m.View().Content), "F6 Navigate") {
		t.Fatal("footer did not hide on leave")
	}
	h.send(tea.MouseMotionMsg{X: g.x + 1, Y: h.m.height - 1, Button: tea.MouseNone})
	h.send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if strings.Contains(ansi.Strip(h.m.View().Content), "F6 Navigate") {
		t.Fatal("typing left the last terminal row obscured")
	}
	if got := readWrites(); !reflect.DeepEqual(got, append(before, "text:x")) {
		t.Fatalf("hover or typing changed input ownership: %q", got)
	}
}
