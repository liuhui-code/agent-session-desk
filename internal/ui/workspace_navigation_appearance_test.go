package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestWorkspaceNavigationStartsAtTopAndClickUsesTheSameRow(t *testing.T) {
	for _, width := range []int{120, 63, 49} {
		m := New(context.Background(), &workspaceFixture{})
		m.width, m.height = width, 24
		m.sessions, m.cursor = sampleSessions(), 1
		m.sidebarMode = "flat"
		m.workspace.focused = false
		m.workspace.screen = terminal.Screen{Width: 10, Height: 1, Rows: []terminal.Row{{Runs: []terminal.Run{{Text: "CONTENT", Width: 7}}}}}
		m.workspace.tabs = []workspaceTab{{session: m.sessions[0]}}
		view := ansi.Strip(m.View().Content)
		first := strings.Split(view, "\n")[0]
		if !strings.Contains(first, m.sessions[0].Title) || strings.Contains(view, "会话 平铺") {
			t.Errorf("width %d: navigation has a heading or top offset", width)
		}
		if width == 120 && !strings.Contains(first, "CONTENT") {
			t.Error("content and navigation do not start on the same row")
		}
		m.workspaceClick(tea.MouseClickMsg{X: 3, Y: 0, Button: tea.MouseLeft})
		if m.cursor != 0 {
			t.Errorf("width %d: first visible row did not select first session", width)
		}
		m.Close()
	}
}

func TestWorkspaceSelectionUsesColorInsteadOfTriangleAndDiffersFromActive(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	defer m.Close()
	m.sessions, m.cursor = sampleSessions(), 1
	m.workspace.tabs = []workspaceTab{{session: m.sessions[0]}}
	active := m.workspaceSidebarCard(0, 40, 2, 0)
	selected := m.workspaceSidebarCard(1, 40, 2, 0)
	if strings.Contains(ansi.Strip(strings.Join(selected, "")), "▶") {
		t.Error("selected session still has a triangle")
	}
	if testBackgroundAt(active[0], 0) == testBackgroundAt(selected[0], 0) {
		t.Error("active content target and list selection have the same surface")
	}
	for _, card := range [][]string{active, selected} {
		for _, line := range card {
			if testBackgroundAt(line, 0) != testBackgroundAt(line, 39) {
				t.Error("card color does not cover the full row")
			}
		}
	}
}
