package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestMinimalWorkspaceNewActionOpensLaunchDialogFromManagement(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
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
	h.run(h.m.closeWorkspaceView(1))
	if len(h.m.workspace.tabs) != 1 || !strings.Contains(h.m.View().Content, "shell marker first") {
		close(gate)
		t.Fatal("closing inactive logical view changed active terminal")
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
	view := ansi.Strip(h.m.View().Content)
	if strings.Contains(view, "×") || strings.Contains(view, "+") || strings.Contains(view, "Agent Session Desk") {
		t.Fatal("long title restored removed duplicate controls")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { return h.m.launch != nil })
}
