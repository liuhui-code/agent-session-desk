package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestWorkspaceNewTabButtonOpensLaunchDialogFromTerminalFocus(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	row := ansi.Strip(strings.Split(h.m.View().Content, "\n")[1])
	idx := strings.LastIndex(row, "+")
	if idx < 0 {
		t.Fatal("Tab bar has no visible new-terminal button")
	}
	h.send(tea.MouseClickMsg{X: ansi.StringWidth(row[:idx]), Y: 1, Button: tea.MouseLeft})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "新建终端") })
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.starts) != 0 {
		t.Fatal("new Tab button started a process before explicit form submission")
	}
}

func TestWorkspaceCloseButtonRemovesInactiveViewWithoutInterruptingActiveInput(t *testing.T) {
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
	h.send(tea.KeyPressMsg{Code: 'a', Text: "a"})
	row := ansi.Strip(strings.Split(h.m.View().Content, "\n")[1])
	start := strings.Index(row, "评审")
	offset := strings.Index(row[start:], "×")
	if offset < 0 {
		close(gate)
		t.Fatal("inactive Tab has no visible close button")
	}
	x := ansi.StringWidth(row[:start+offset])
	h.send(tea.MouseClickMsg{X: x, Y: 1, Button: tea.MouseLeft})
	row = ansi.Strip(strings.Split(h.m.View().Content, "\n")[1])
	if strings.Contains(row, "评审") || !strings.Contains(h.m.View().Content, "shell marker first") {
		close(gate)
		t.Fatal("inactive close changed active terminal or kept old Tab")
	}
	close(gate)
	h.until(func() bool {
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Join(stream.writes, ",") == "text:a"
	})
	service.mu.Lock()
	defer service.mu.Unlock()
	if strings.Join(service.opened, ",") != "first" || len(service.starts) != 0 {
		t.Fatal("inactive close reconnected or started a process")
	}
}

func TestWorkspaceLongTabTitleKeepsCloseAndNewButtonsReachable(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) {
		rows := sampleSessions()
		rows[0].Title = strings.Repeat("超长标题", 100)
		return rows, nil
	}}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	row := ansi.Strip(strings.Split(h.m.View().Content, "\n")[1])
	if !strings.Contains(row, "×") || !strings.Contains(row, "+") {
		t.Fatal("long title hides primary Tab controls")
	}
	if !strings.Contains(row, "…") {
		t.Fatal("long Tab title has no truncation indication")
	}
}
