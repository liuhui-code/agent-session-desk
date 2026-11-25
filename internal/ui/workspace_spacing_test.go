package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestWorkspaceChromeNeverCoversContentOnHover(t *testing.T) {
	h := newWorkspaceHarness(t, &workspaceFixture{fakeService: directDragFixture().fakeService})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker") })
	h.send(tea.MouseMotionMsg{X: h.m.width - 1, Y: h.m.height - 1, Button: tea.MouseNone})
	if h.m.workspaceFooterVisible() || strings.Contains(h.m.View().Content, "F6") {
		t.Fatal("ordinary hover still reveals mode hints over content")
	}
	h.m.status = "Save failed"
	if !strings.Contains(h.m.View().Content, "Save failed") {
		t.Fatal("critical feedback was hidden")
	}
}

func TestGroupHeaderHasStableGutterWithoutNumericBadge(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	defer m.Close()
	m.sessions = sampleSessions()
	m.sidebarMode, m.cursor = "grouped", 0
	var previous string
	for _, focused := range []bool{false, true} {
		m.workspace.focused = focused
		line := ansi.Strip(m.renderSidebarGroupAt("default", 0, 40))
		if strings.Contains(line, "1·") {
			t.Fatal("group header still shows a root number")
		}
		if previous != "" && line != previous {
			t.Fatal("group header shifts when focus changes")
		}
		previous = line
	}
}

func TestSidebarSpacingUsesSharedVisibleRows(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	defer m.Close()
	m.sessions = sampleSessions()
	m.sidebarMode, m.cursor = "flat", 0
	lines := strings.Split(m.renderWorkspaceSidebar(40, 20), "\n")
	window := m.sidebarNavigationWindow(40, 20)
	if len(window.rows) != 2 || window.rows[1].top <= m.workspaceSidebarRows(40, 20) {
		t.Fatal("session cards still touch without breathing room")
	}
	if strings.TrimSpace(ansi.Strip(lines[window.rows[1].top-1])) != "" {
		t.Fatal("space between cards contains decoration")
	}
}
