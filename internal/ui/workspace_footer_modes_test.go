package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestWorkspaceCuratedFooterUsesThreeRelevantActionsThenSettingsAndHelp(t *testing.T) {
	m := NewWithOptions(context.Background(), newGroupedFixture(), Options{Footer: "curated", SidebarMode: "flat"})
	defer m.Close()
	m.width, m.height = 160, 32
	m.sessions = sampleSessions()
	m.cursor = 0
	m.workspace.focused = false
	plain := strings.TrimSpace(ansi.Strip(strings.Split(m.View().Content, "\n")[30]))
	if plain != "Enter 打开  n 新建  r 改名  S 设置  ? 帮助" {
		t.Fatalf("curated mode is not rendered: %q", plain)
	}
}

func TestWorkspaceCuratedFooterUsesDimRolesAndActualSelectionContext(t *testing.T) {
	for _, theme := range []string{"dark", "light"} {
		m := NewWithOptions(context.Background(), newGroupedFixture(), Options{Theme: theme, Footer: "curated", SidebarMode: "flat"})
		t.Cleanup(m.Close)
		m.width, m.height = 160, 32
		m.sessions, m.cursor, m.workspace.focused = sampleSessions(), 0, false
		line := strings.Split(m.View().Content, "\n")[30]
		r, g, b, _ := m.tokens.colors.dim.RGBA()
		want := fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8)
		if testForegroundAt(line, 0) != want || testBackgroundAt(line, 0) != "default" {
			t.Fatal("curated key lost dim role or gained filled chip")
		}
		m.sessions[0].TerminalState = session.TerminalExited
		plain := ansi.Strip(strings.Split(m.View().Content, "\n")[30])
		if strings.Contains(plain, "Enter") || !strings.Contains(plain, "S 设置  ? 帮助") {
			t.Fatal("unavailable session advertised open or lost globals")
		}
		m.sidebarMode = "grouped"
		m.cursor = -1
		if plain = ansi.Strip(strings.Split(m.View().Content, "\n")[30]); !strings.Contains(plain, "Tab 折叠") {
			t.Fatal("expanded group context missing")
		}
		m.sidebarCollapsed = true
		if plain = ansi.Strip(strings.Split(m.View().Content, "\n")[30]); !strings.Contains(plain, "Tab 展开") {
			t.Fatal("collapsed group context missing")
		}
		m.sidebarMode = "flat"
		m.sessions = nil
		if plain = strings.TrimSpace(ansi.Strip(strings.Split(m.View().Content, "\n")[30])); plain != "n 新建  g 分组  S 设置  ? 帮助" {
			t.Fatalf("empty-list context: %s", plain)
		}
	}
}

func TestWorkspaceFooterModesRetainBoundedFramesAndTerminalOwnership(t *testing.T) {
	for _, mode := range []string{"full", "compact", "minimal", "curated"} {
		for _, width := range []int{1, 10, 22, 27, 49, 50, 63, 69, 70, 80, 100, 101, 120, 160} {
			for _, focused := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%d/focus%t", mode, width, focused), func(t *testing.T) {
					m := NewWithOptions(context.Background(), newGroupedFixture(), Options{Footer: mode, SidebarMode: "flat"})
					defer m.Close()
					m.width, m.height = width, 32
					m.sessions, m.cursor, m.workspace.focused = sampleSessions(), 0, focused
					view := m.View().Content
					lines := strings.Split(view, "\n")
					if len(lines) != 32 || !utf8.ValidString(view) {
						t.Fatal("invalid frame")
					}
					for _, line := range lines {
						if lipgloss.Width(line) != width {
							t.Fatal("frame column mismatch")
						}
					}
					footer := m.renderManagementFooter()
					if focused {
						footer = m.renderTerminalFooter()
					}
					if lipgloss.Width(footer) > width {
						t.Fatal("footer overflow was hidden by final frame crop")
					}
					plain := ansi.Strip(lines[30])
					if focused && width >= 22 && !strings.Contains(plain, "F6") {
						t.Fatal("footer preference hid terminal input ownership")
					}
					if !focused && width >= 50 && mode == "curated" && !strings.HasSuffix(strings.TrimSpace(plain), "S 设置  ? 帮助") {
						t.Fatal("curated globals cropped")
					}
				})
			}
		}
	}
}

func TestWorkspaceFooterChoiceSurvivesOtherSettingsAndLeavesTerminalInputLiteral(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.toml")
	s := directDragFixture()
	h := newWorkspaceHarness(t, s, Options{Footer: "curated", SidebarMode: "flat", SaveUISettings: func(_ context.Context, v config.UISettings) error { return config.SaveUI(path, v) }})
	h.until(func() bool { return h.m.workspace.conn != nil })
	s.mu.Lock()
	original := s.streams["first"]
	s.mu.Unlock()
	before := h.m.workspace.tabs[h.m.workspace.active].session
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	openThemeSettings(t, h)
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	h.until(func() bool { return !h.m.preferences.busy })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.until(func() bool { return h.m.densitySettings == nil && h.m.workspace.conn != nil })
	saved, err := config.LoadUI(path)
	if err != nil || saved.Footer != "curated" || saved.Theme != "light" || saved.SidebarDensity != "minimal" {
		t.Fatalf("combined save lost footer or other choices: %+v %v", saved, err)
	}
	if !sameRun(before, h.m.workspace.tabs[h.m.workspace.active].session) {
		t.Fatal("footer setting changed original Session/Run")
	}
	original.mu.Lock()
	defer original.mu.Unlock()
	count := 0
	for _, v := range original.writes {
		if v == "text:S" {
			count++
		} else if strings.HasPrefix(v, "text:") || strings.HasPrefix(v, "key:") {
			t.Errorf("settings leaked input: %s", v)
		}
	}
	if count != 1 {
		t.Fatalf("terminal-focus S was not delivered exactly once: %d", count)
	}
}

func TestWorkspaceFooterCanForceCompactOrMinimalAtWideWidths(t *testing.T) {
	for _, mode := range []string{"compact", "minimal"} {
		t.Run(mode, func(t *testing.T) {
			m := NewWithOptions(context.Background(), newGroupedFixture(), Options{Footer: mode, SidebarMode: "flat"})
			defer m.Close()
			m.width, m.height = 160, 32
			m.sessions, m.cursor, m.workspace.focused = sampleSessions(), 0, false
			line := strings.Split(m.View().Content, "\n")[30]
			plain := ansi.Strip(line)
			i := strings.Index(plain, "Enter")
			if i < 0 {
				t.Fatal("forced tier lost Enter")
			}
			r, g, b, _ := m.tokens.colors.accent.RGBA()
			if testBackgroundAt(line, ansi.StringWidth(plain[:i])-1) == fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8) {
				t.Fatal("forced tier retained full-tier padded key chip")
			}
			if strings.Contains(plain, "打开") != (mode == "compact") {
				t.Fatalf("forced tier labels incorrect: %s", plain)
			}
			for _, key := range []string{"F6", "?", "q"} {
				if !strings.Contains(plain, key) {
					t.Fatalf("forced tier lost %s", key)
				}
			}
		})
	}
}
