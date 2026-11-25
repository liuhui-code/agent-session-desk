package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestNestedSessionCardHighlightIncludesIndentationAtEveryDensity(t *testing.T) {
	for _, density := range []string{"minimal", "compact", "full", "auto"} {
		for _, focused := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/focused=%t", density, focused), func(t *testing.T) {
				m := New(context.Background(), newGroupedFixture())
				t.Cleanup(m.Close)
				m.sidebarMode, m.sidebarDensity = "grouped", density
				m.groups = []session.Group{{ID: session.DefaultGroupID, Name: "Default"}, {ID: "root", Name: "Project"}, {ID: "nested", ParentID: "root", Name: "Review"}}
				m.sessions = sampleSessions()
				m.sessions[0].GroupID = "nested"
				m.sessions[0].Title = strings.Repeat("Review 中文 ", 8)
				m.cursor = 0
				m.workspace.focused = focused
				if focused {
					m.workspace.tabs = []workspaceTab{{viewID: 1, session: m.sessions[0]}}
				}
				const width = 42
				window := m.sidebarNavigationWindow(width, 20)
				lines := strings.Split(m.renderWorkspaceSidebar(width, 20), "\n")
				r, g, b, _ := m.tokens.colors.surface.RGBA()
				want := fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8)
				found := false
				for _, hit := range window.rows {
					if hit.index != 0 {
						continue
					}
					found = true
					for y := hit.top; y < hit.top+hit.height; y++ {
						if ansi.StringWidth(lines[y]) != width {
							t.Fatalf("card row width=%d", ansi.StringWidth(lines[y]))
						}
						for x := 0; x < width; x++ {
							if got := testBackgroundAt(lines[y], x); got != want {
								t.Fatalf("highlight hole at nested card (%d,%d): got %q want %s", x, y, got, want)
							}
						}
					}
				}
				if !found {
					t.Fatal("nested selection missing")
				}
			})
		}
	}
}

func TestWorkspaceDividerStaysQuietAcrossFocusHoverAndDrag(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	t.Cleanup(m.Close)
	m.width, m.height = 120, 32
	m.workspace.focused = false
	if m.View().MouseMode != tea.MouseModeAllMotion {
		t.Fatal("divider hover needs no-button motion reporting in the actual View")
	}
	check := func(glyph string) {
		t.Helper()
		g := m.workspaceGeometry()
		line := strings.Split(m.View().Content, "\n")[g.y+2]
		cell := ansi.Cut(line, g.left, g.left+3)
		if ansi.Strip(cell) != " "+glyph+" " {
			t.Fatalf("divider glyph: %q", ansi.Strip(cell))
		}
		c := m.tokens.colors.border
		r, gc, b, _ := c.RGBA()
		want := fmt.Sprintf("%d,%d,%d", r>>8, gc>>8, b>>8)
		if got := testForegroundAt(line, g.left+1); got != want {
			t.Fatalf("divider role differs: got %s want %s", got, want)
		}
	}
	check("│")
	m.workspace.focused = true
	check("│")
	m.workspace.focused = false
	g := m.workspaceGeometry()
	m.workspaceSplitMotion(tea.MouseMotionMsg{X: g.left + 1, Y: g.y, Button: tea.MouseNone})
	check("┃")
	if m.cursor != 0 || m.workspace.focused || m.workspace.pending != 0 || m.workspace.split.override {
		t.Fatal("hover changed selection, input ownership or split")
	}
	m.workspaceSplitMotion(tea.MouseMotionMsg{X: 0, Y: g.y, Button: tea.MouseNone})
	check("│")
	m.workspaceClick(tea.MouseClickMsg{X: g.left + 1, Y: g.y, Button: tea.MouseLeft})
	check("┃")
	m.workspaceSplitRelease(tea.MouseReleaseMsg{X: 0, Y: g.y, Button: tea.MouseLeft})
	check("│")
	m.workspaceSplitMotion(tea.MouseMotionMsg{X: g.left + 1, Y: g.y, Button: tea.MouseNone})
	m.Update(tea.WindowSizeMsg{Width: 63, Height: 25})
	if m.workspace.split.hovered {
		t.Fatal("resize retained a hover for an obsolete divider")
	}
}
