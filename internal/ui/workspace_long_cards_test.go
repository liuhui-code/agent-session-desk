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

func TestLongCardTitleUsesSourceColumnCropWithoutInventingAnEllipsis(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	defer m.Close()
	m.sessions = []session.Session{{ID: "one", RunID: "run-one", Title: "abcdefghijklmnopqrstuvwxyz0123456789"}}
	m.cursor = 0
	m.workspace.focused = false
	got := ansi.Strip(m.workspaceSidebarCard(0, 28, 1, 0)[0])
	// The pinned card renderer fits the identity line with an empty tail.
	if got != "  abcdefghijklmnopqrstuvwxyz" {
		t.Fatalf("card differs from source column crop: %q", got)
	}
}

func TestLongExecutablePathShowsItsActualBasenameBeforeBoundingDisplay(t *testing.T) {
	for _, path := range []string{`C:\` + strings.Repeat(`deep\`, 1000) + `actual-agent.exe`, "/" + strings.Repeat("deep/", 1000) + "actual-agent.exe"} {
		m := New(context.Background(), &workspaceFixture{})
		m.sessions = []session.Session{{ID: "one", RunID: "run-one", Title: "One", Command: path, TerminalState: session.TerminalAvailable}}
		m.workspace.focused = false
		m.cursor = 0
		got := ansi.Strip(strings.Join(m.workspaceSidebarCard(0, 42, 2, 0), "\n"))
		if !strings.Contains(got, "actual-agent.exe") || m.sessions[0].Command != path {
			m.Close()
			t.Fatalf("display used a truncated directory or changed launch data: %q", got)
		}
		m.Close()
	}
}

func TestLongUnicodeCardsKeepFixedRowsAndSelectedBackground(t *testing.T) {
	for _, title := range []string{strings.Repeat("中文任务", 30), strings.Repeat("e\u0301", 80), strings.Repeat("1️⃣", 50), strings.Repeat("👩‍💻", 30)} {
		for _, width := range []int{27, 28, 29, 42, 63} {
			for _, level := range []int{1, 3, 8} {
				for _, rows := range []int{1, 2, 3} {
					m := New(context.Background(), &workspaceFixture{})
					m.sessions = []session.Session{{ID: "one", RunID: "original", Title: title, Command: strings.Repeat("1️⃣", 30) + ".exe", TerminalState: session.TerminalAvailable}}
					m.cursor = 0
					m.workspace.focused = false
					lines := m.workspaceSidebarCard(0, width, rows, level)
					if len(lines) != rows {
						m.Close()
						t.Fatal("title/metadata expanded fixed card height")
					}
					r, g, b, _ := m.tokens.colors.focusBackground.RGBA()
					want := fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8)
					for y, line := range lines {
						if ansi.StringWidth(line) != width {
							m.Close()
							t.Fatalf("card %d/%d/%d row%d width%d", width, level, rows, y, ansi.StringWidth(line))
						}
						for x := 0; x < width; x++ {
							if testBackgroundAt(line, x) != want {
								m.Close()
								t.Fatalf("long selected card background hole at %d/%d", x, y)
							}
						}
					}
					m.Close()
				}
			}
		}
	}
}

func TestLongCardViewportMetadataClickAndOpenKeepOriginalRunAcrossDensities(t *testing.T) {
	for _, density := range []string{"minimal", "compact", "full", "auto"} {
		t.Run(density, func(t *testing.T) {
			s := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) {
				items := []session.Session{}
				for i := 0; i < 30; i++ {
					id := fmt.Sprintf("%032x", i+1)
					items = append(items, session.Session{ID: id, RunID: "run-" + id, BackendID: "backend-" + id, Title: strings.Repeat("相同前缀1️⃣e\u0301", 12) + fmt.Sprint(i), Command: "/synthetic/" + strings.Repeat("long-program-", 30) + ".exe", TerminalState: session.TerminalAvailable})
				}
				return items, nil
			}}}
			h := newWorkspaceHarness(t, s, Options{SidebarMode: "flat", SidebarDensity: density})
			h.until(func() bool { return h.m.workspace.conn != nil })
			s.mu.Lock()
			original := s.streams[h.m.workspace.tabs[0].session.ID]
			s.mu.Unlock()
			h.send(tea.KeyPressMsg{Code: 'd', Text: "protected-draft"})
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			for i := 0; i < 15; i++ {
				h.send(tea.KeyPressMsg{Code: tea.KeyDown})
			}
			target, _ := h.m.selected()
			for _, size := range [][2]int{{120, 24}, {120, 14}, {79, 14}, {63, 24}, {49, 14}, {27, 14}, {120, 24}} {
				h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
				g := h.m.workspaceGeometry()
				height := g.listHeight
				window := h.m.sidebarNavigationWindow(g.listWidth, height)
				found := false
				for _, hit := range window.rows {
					if hit.index != h.m.cursor {
						continue
					}
					found = true
					y := g.y - g.above - h.m.styles.terminalFrame.GetBorderTopSize() + hit.top + hit.height - 1
					h.send(tea.MouseClickMsg{X: 3, Y: y, Button: tea.MouseLeft})
					selected, ok := h.m.selected()
					if !ok || !sameStopTarget(selected, target) || h.m.workspace.focused {
						t.Fatal("long metadata click changed selected identity or activated terminal")
					}
					break
				}
				if !found {
					t.Fatal("selected long card disappeared after resize")
				}
				for _, line := range strings.Split(h.m.View().Content, "\n") {
					if ansi.StringWidth(line) != size[0] {
						t.Fatal("long card overflowed viewport")
					}
				}
			}
			s.mu.Lock()
			opened := len(s.opened)
			s.mu.Unlock()
			if opened != 1 {
				t.Fatal("single metadata click opened a terminal")
			}
			h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
			h.until(func() bool {
				return h.m.workspace.conn != nil && sameRun(h.m.workspace.tabs[h.m.workspace.active].session, target)
			})
			original.mu.Lock()
			defer original.mu.Unlock()
			writes := strings.Join(original.writes, ",")
			if strings.Count(writes, "text:protected-draft") != 1 || strings.Contains(writes, "key:Enter") {
				t.Fatal("selection lost/replayed/submitted the original draft")
			}
		})
	}
}

func TestExecutableDisplaySanitizesEscapesBeforeFindingTheBasename(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	defer m.Close()
	// An OSC URI contains separators that are not part of the inert executable.
	// Finding the basename inside that control sequence would expose false.exe.
	command := "/synthetic/actual-agent.exe\x1b]8;;https://synthetic.invalid/path/false.exe\a"
	m.sessions = []session.Session{{ID: "one", RunID: "original", Title: "One", Command: command, TerminalState: session.TerminalAvailable}}
	m.cursor = 0
	m.workspace.focused = false
	rendered := strings.Join(m.workspaceSidebarCard(0, 42, 2, 0), "\n")
	plain := ansi.Strip(rendered)
	if !strings.Contains(plain, "actual-agent.exe") || strings.Contains(plain, "false.exe") || strings.Contains(rendered, "\x1b]8") || m.sessions[0].Command != command {
		t.Fatalf("control sequence changed executable label: %q", plain)
	}
}
