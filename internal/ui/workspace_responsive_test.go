package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestWorkspaceNarrowManagementShowsStackedListAndTerminal(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.WindowSizeMsg{Width: 60, Height: 24})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	view := ansi.Strip(h.m.View().Content)
	if !strings.Contains(view, "┃ 开发") || !strings.Contains(view, "shell marker first") {
		t.Fatal("narrow management hides session navigation or terminal")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.View().Cursor == nil || strings.Contains(ansi.Strip(h.m.View().Content), "┃ 开发") {
		t.Fatal("terminal focus does not expand back to full content area")
	}
}

func TestWorkspaceVeryNarrowManagementSelectsFromSingleList(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.WindowSizeMsg{Width: 40, Height: 24})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	view := ansi.Strip(h.m.View().Content)
	if !strings.Contains(view, "┃ 开发") || strings.Contains(view, "shell marker first") {
		t.Fatal("single-list management layout not shown below fifty columns")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		return strings.Contains(h.m.View().Content, "shell marker second") && h.m.View().Cursor != nil
	})
}

func TestWorkspaceNarrowSwitchKeepsAcceptedInputWhenF6Pressed(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.WindowSizeMsg{Width: 40, Height: 24})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	service.mu.Lock()
	first := service.streams["first"]
	service.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'a', Text: "a"})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: 'b', Text: "b"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	close(gate)
	h.until(func() bool {
		service.mu.Lock()
		second := service.streams["second"]
		service.mu.Unlock()
		if second == nil {
			return false
		}
		first.mu.Lock()
		a := strings.Join(first.writes, ",")
		first.mu.Unlock()
		second.mu.Lock()
		b := strings.Join(second.writes, ",")
		second.mu.Unlock()
		return strings.Contains(a, "text:a") && !strings.Contains(a, "text:b") && strings.Contains(b, "text:b") && !strings.Contains(b, "text:a")
	})
}
