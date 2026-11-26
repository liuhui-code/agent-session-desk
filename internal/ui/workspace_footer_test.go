package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestWorkspaceManagementFooterKeepsWideActionChipsAndCompactNavigation(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	// Exact pinned source output; authored accent remains #7aa2f7.
	want := "187,187,187"
	for _, width := range []int{120, 80, 63} {
		h.send(tea.WindowSizeMsg{Width: width, Height: 32})
		line := h.m.renderManagementFooter()
		plain := ansi.Strip(line)
		if width == 120 {
			i := strings.Index(plain, "Enter")
			if i < 0 {
				t.Fatalf("wide footer lost selected-session Enter action: %q", plain)
			}
			if got := testBackgroundAt(line, ansi.StringWidth(plain[:i])); got != want {
				t.Fatalf("wide action lost accent chip: %s want %s", got, want)
			}
		} else if !strings.Contains(plain, "Alt+J/K") {
			t.Fatalf("compact footer lost session navigation: %q", plain)
		}
		for _, key := range []string{"F6", "?", "q"} {
			if !strings.Contains(plain, key) {
				t.Fatalf("%d footer lost %s", width, key)
			}
		}
	}
}

func TestWorkspaceFooterContextMatchesSelectedGroupAndUnavailableSession(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	t.Cleanup(m.Close)
	m.width, m.height = 160, 32
	m.sidebarMode = "grouped"
	m.cursor = -1
	footer := ansi.Strip(m.renderManagementFooter())
	if !strings.Contains(footer, "Collapse") || strings.Contains(footer, "Open") {
		t.Fatalf("group footer advertises terminal open: %s", footer)
	}
	m.sidebarCollapsed = true
	if footer = ansi.Strip(m.renderManagementFooter()); !strings.Contains(footer, "Expand") {
		t.Fatalf("collapsed group action not reflected: %s", footer)
	}
	m.sessions = sampleSessions()
	m.cursor = 0
	m.sessions[0].TerminalState = session.TerminalExited
	if footer = ansi.Strip(m.renderManagementFooter()); strings.Contains(footer, "Enter") {
		t.Fatalf("unavailable session advertises Enter: %s", footer)
	}
}

func TestWorkspaceFooterBreakpointsAndLongActiveTitlesStayInsideViewport(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	h.m.workspace.tabs[0].session.Title = strings.Repeat("中文👩🏽‍💻é", 40)
	for _, width := range []int{1, 10, 22, 27, 49, 50, 63, 69, 70, 80, 100, 101, 120, 160} {
		for _, focused := range []bool{false, true} {
			h.m.width, h.m.height = width, 32
			h.m.workspace.focused = focused
			view := h.m.View()
			lines := strings.Split(view.Content, "\n")
			if len(lines) != 32 || !utf8.ValidString(view.Content) {
				t.Fatalf("%d focused=%t invalid frame", width, focused)
			}
			for _, line := range lines {
				if lipgloss.Width(line) != width {
					t.Fatalf("%d focused=%t row overflow", width, focused)
				}
			}
		}
	}
}

func TestWorkspaceFullFooterKeepsQuitAtSourceRightGutter(t *testing.T) {
	m := New(context.Background(), newGroupedFixture())
	t.Cleanup(m.Close)
	m.sessions = sampleSessions()
	m.cursor = 0
	m.width, m.height = 160, 32
	m.sidebarMode = "flat"
	line := ansi.Strip(m.renderManagementFooter())
	if !strings.HasSuffix(line, "q Quit") || lipgloss.Width(line) != m.width {
		t.Fatalf("source full-tier quit/right-gutter anchor missing: width=%d %q", lipgloss.Width(line), line)
	}
}

func TestWorkspaceNarrowTerminalFooterOmitsUnrenderableTargetInsteadOfAnIDSuffix(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	t.Cleanup(m.Close)
	m.width, m.height = 49, 32
	m.sessions = sampleSessions()
	m.sessions[0].Title = "Copy A"
	m.sessions[1].Title = "Copy B"
	m.workspace.tabs = []workspaceTab{{viewID: 1, session: m.sessions[0]}, {viewID: 2, session: m.sessions[1]}}
	m.workspace.focused = true
	line := ansi.Strip(m.renderTerminalFooter())
	if strings.Contains(line, m.tokens.footer.targetPrefix) || strings.Contains(line, "]") || !strings.Contains(line, "F6") || !strings.Contains(line, "Alt+J/K") {
		t.Fatalf("narrow footer displays a target fragment or loses navigation: %q", line)
	}
}

func TestWorkspaceTerminalFooterUsesSourceYellowBadgeAndNamesTheActiveRun(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	r, g, b, _ := h.m.tokens.colors.yellow.RGBA()
	want := fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8)
	check := func(title string) {
		line := h.m.renderTerminalFooter()
		plain := ansi.Strip(line)
		i := strings.Index(plain, "Terminal")
		if i < 0 || !strings.Contains(plain, title) || !strings.Contains(plain, "F6") {
			t.Fatalf("terminal badge/target missing: %q", plain)
		}
		if got := testBackgroundAt(line, ansi.StringWidth(plain[:i])); got != want {
			t.Fatalf("terminal mode lost source yellow badge: %s want %s", got, want)
		}
	}
	check("开发")
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	check("评审")
}
