package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestWorkspaceHelpOpensSourceStyleShortCard(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '?', Text: "?"})
	h.until(func() bool { return h.m.help && h.m.workspace.conn == nil })
	view := ansi.Strip(h.m.View().Content)
	if !strings.Contains(view, "Quick start") || !strings.Contains(view, "╭") || strings.Contains(view, "Getting started\n") {
		t.Fatalf("help did not open the centered source-style short card:\n%s", view)
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	if !strings.Contains(ansi.Strip(h.m.View().Content), "All shortcuts") {
		t.Fatal("Tab did not reveal the full shortcut list")
	}
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if h.m.help || h.m.closed {
		t.Fatal("q must only close help, without also quitting the manager")
	}
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
}

func TestWorkspaceHelpConsumesMouseAndClosingKeysOnce(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	for _, key := range []tea.KeyPressMsg{{Code: 'n', Text: "n"}, {Code: 'q', Text: "q"}, {Code: tea.KeyF6}, {Code: tea.KeyPgDown, Mod: tea.ModCtrl}, {Code: 'c', Mod: tea.ModCtrl}, {Code: tea.KeyEnter}} {
		h.send(tea.KeyPressMsg{Code: '?', Text: "?"})
		h.until(func() bool { return h.m.help && h.m.workspace.conn == nil })
		h.send(tea.MouseClickMsg{X: 5, Y: 1, Button: tea.MouseLeft})
		h.send(tea.MouseWheelMsg{X: 30, Y: 10, Button: tea.MouseWheelDown})
		if !h.m.help {
			t.Fatal("mouse dismissed help")
		}
		h.send(key)
		if h.m.closed || h.m.help || h.m.launch != nil || h.m.workspace.focused || h.m.workspace.active != 0 {
			t.Fatalf("closing help also executed %s", key.String())
		}
		h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.starts) > 0 {
		t.Fatal("help closing key created a process")
	}
	for _, id := range service.opened {
		if id != "first" {
			t.Fatal("help changed the target Run")
		}
	}
	for _, stream := range service.streams {
		stream.mu.Lock()
		writes := append([]string(nil), stream.writes...)
		stream.mu.Unlock()
		for _, write := range writes {
			if strings.HasPrefix(write, "text:") || strings.HasPrefix(write, "key:") || strings.HasPrefix(write, "paste:") {
				t.Fatalf("help leaked terminal input: %s", write)
			}
		}
	}
}

func TestWorkspaceHelpBothPagesStayInsideResizedViewport(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	t.Cleanup(m.Close)
	m.help = true
	for _, width := range []int{1, 5, 22, 37, 49, 63, 79, 99, 100, 120} {
		for _, height := range []int{1, 4, 8, 12, 16, 25, 32} {
			m.width, m.height = width, height
			for _, all := range []bool{false, true} {
				m.helpPage = helpPage{all: all, offset: 9999}
				view := m.View()
				lines := strings.Split(view.Content, "\n")
				if len(lines) != height || view.Cursor != nil {
					t.Fatalf("%dx%d invalid help height/cursor", width, height)
				}
				for _, line := range lines {
					if lipgloss.Width(line) != width {
						t.Fatalf("%dx%d help row overflow", width, height)
					}
				}
			}
		}
	}
}

func TestWorkspaceHelpWaitsForAcceptedTerminalInput(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	gate := make(chan struct{})
	stream.mu.Lock()
	stream.gate = gate
	stream.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '?', Text: "?"})
	if h.m.help || h.m.workspace.conn == nil {
		close(gate)
		t.Fatal("help disconnected accepted input before the write completed")
	}
	close(gate)
	h.until(func() bool { return h.m.help && h.m.workspace.conn == nil })
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if len(stream.writes) == 0 || stream.writes[0] != "text:d" {
		t.Fatalf("accepted input lost or moved: %v", stream.writes)
	}
}

func TestWorkspaceHelpLastPageMatchesSourceTrailingRow(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	t.Cleanup(m.Close)
	m.help, m.helpPage.all = true, true
	for _, size := range [][2]int{{120, 32}, {63, 25}, {49, 25}} {
		m.width, m.height = size[0], size[1]
		height := func() int {
			first, last := -1, -1
			for i, line := range strings.Split(ansi.Strip(m.View().Content), "\n") {
				if strings.Contains(line, "╭") {
					first = i
				}
				if strings.Contains(line, "╰") {
					last = i
				}
			}
			return last - first + 1
		}
		m.helpPage.offset = 0
		first := height()
		m.helpPage.offset = 9999
		if last := height(); last != first-1 {
			t.Fatalf("%dx%d source omits the final bottom indicator row: first=%d last=%d", size[0], size[1], first, last)
		}
	}
}

func TestWorkspaceFullHelpScrollsToTheLastRowAndBack(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '?', Text: "?"})
	h.until(func() bool { return h.m.help && h.m.workspace.conn == nil })
	h.send(tea.WindowSizeMsg{Width: 49, Height: 25})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	initial := ansi.Strip(h.m.View().Content)
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	last := ansi.Strip(h.m.View().Content)
	if !h.m.help || initial == last || !strings.Contains(last, "Closing a tab does not stop its run") || strings.Contains(last, "下方还有") {
		t.Fatalf("full help tail is unreachable or still claims more rows:\n%s", last)
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown})
	if ansi.Strip(h.m.View().Content) != last {
		t.Fatal("scroll advanced past the last page")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyHome})
	if ansi.Strip(h.m.View().Content) != initial {
		t.Fatal("Home did not return to the same first page")
	}
	h.send(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if !h.m.help || ansi.Strip(h.m.View().Content) == initial {
		t.Fatal("j did not scroll the help rows")
	}
	h.send(tea.KeyPressMsg{Code: 'k', Text: "k"})
	if ansi.Strip(h.m.View().Content) != initial {
		t.Fatal("k did not return to the first row")
	}
}
