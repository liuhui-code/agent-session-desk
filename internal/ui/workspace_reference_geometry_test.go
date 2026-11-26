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

// Fixed upstream's real 120x32 interactive frame: terminal top border row 2,
// footer rule row 29, help row 30 and bottom blank row 31 (zero-based).
func TestWorkspaceUserLayoutSupersedesUpstreamFrameBands(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	view := h.m.View()
	lines := strings.Split(ansi.Strip(view.Content), "\n")
	if strings.Contains(view.Content, "╭") || strings.Contains(lines[0], "Agent Session Desk") {
		t.Fatal("removed frame bands returned")
	}
	if view.Cursor == nil || view.Cursor.Y != 1 || len(lines) != 32 {
		t.Fatalf("borderless full-height cursor/frame mismatch: %+v", view.Cursor)
	}
}

func TestWorkspaceDialogOuterWidthsMatchUpstreamActualFrames(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		key        rune
		width      int
	}{{"settings", "Settings", 'S', 66}, {"new-terminal", "New terminal", 'n', 86}} {
		t.Run(tc.name, func(t *testing.T) {
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
			h := newWorkspaceHarness(t, service)
			h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			h.send(tea.KeyPressMsg{Code: tc.key, Text: string(tc.key)})
			h.until(func() bool { return strings.Contains(h.m.View().Content, tc.text) })
			border := "╭" + strings.Repeat("─", tc.width-2) + "╮"
			if !strings.Contains(ansi.Strip(h.m.View().Content), border) {
				t.Fatalf("%s outer width differs from actual upstream %d-column frame", tc.name, tc.width)
			}
		})
	}
}

func TestLaunchDialogKeepsUpstreamRegularAndFirstSessionPadding(t *testing.T) {
	for _, tc := range []struct {
		name      string
		empty     bool
		blankRows int
	}{{"regular", false, 2}, {"first-session", true, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) {
				if tc.empty {
					return nil, nil
				}
				return sampleSessions(), nil
			}}}
			h := newWorkspaceHarness(t, service)
			h.until(func() bool { return !h.m.listBusy })
			if !tc.empty {
				h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
				h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			}
			h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
			h.until(func() bool { return strings.Contains(h.m.View().Content, "New terminal") })
			lines := strings.Split(ansi.Strip(h.m.View().Content), "\n")
			border := "╭" + strings.Repeat("─", 84) + "╮"
			for y, line := range lines {
				if at := strings.Index(line, border); at >= 0 {
					left := ansi.StringWidth(line[:at])
					for gap := 1; gap <= tc.blankRows; gap++ {
						if strings.TrimSpace(ansi.Cut(lines[y+gap], left+1, left+85)) != "" {
							t.Fatalf("%s missing upstream padding row %d", tc.name, gap)
						}
					}
					if !strings.Contains(lines[y+tc.blankRows+1], "New terminal") {
						t.Fatal("title did not follow the expected padding")
					}
					return
				}
			}
			t.Fatal("dialog border not found")
		})
	}
}

func TestWorkspaceNarrowFooterKeepsHelpQuitAndTerminalFocusKeys(t *testing.T) {
	for _, width := range []int{27, 49, 63} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
			h := newWorkspaceHarness(t, service)
			h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			h.send(tea.WindowSizeMsg{Width: width, Height: 25})
			h.send(tea.MouseMotionMsg{X: 0, Y: h.m.height - 1, Button: tea.MouseNone})
			footer := ansi.Strip(strings.Split(h.m.View().Content, "\n")[24])
			for _, key := range []string{"?", "q", "F6"} {
				if !strings.Contains(footer, key) {
					t.Fatalf("%d-column footer lost reachable key %s: %q", width, key, footer)
				}
			}
		})
	}
}
