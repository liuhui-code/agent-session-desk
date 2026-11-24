package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestWorkspaceOldListCannotInvalidateJustCreatedTerminal(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	oldGeneration := h.m.listGeneration
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { return h.m.launch != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: "pwsh.exe"})
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker created-1") })
	// This snapshot was requested before Start finished and arrives afterwards.
	h.send(sessionsMsg{generation: oldGeneration, sessions: sampleSessions()})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker created-1") })
}

func TestWorkspaceDiscoveryBeforeStartResultDoesNotDuplicateTab(t *testing.T) {
	gate := make(chan struct{})
	service := &workspaceFixture{startGate: gate, fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return !h.m.listBusy })
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: "pwsh.exe"})
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.until(func() bool { service.mu.Lock(); defer service.mu.Unlock(); return len(service.created) == 1 })
	h.send(pollMsg{generation: h.m.pollGeneration})
	h.until(func() bool { return len(h.m.sessions) == 1 })
	close(gate)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker created-1") })
	if len(h.m.workspace.tabs) != 1 || h.m.workspace.tabs[0].session.ID != "created-1" {
		t.Fatal("Start result duplicated the discovered logical view")
	}
}

func TestWorkspaceClosingLastTabReturnsManagementFocus(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	for _, id := range []string{"first", "second"} {
		h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker "+id) })
		if h.m.workspace.focused {
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
		}
		h.send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	}
	if h.m.workspace.focused || h.m.View().Cursor != nil {
		t.Fatal("empty workspace still claims terminal focus")
	}
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "新建终端") })
}

func TestWorkspaceEmptyDiscoveryAndF6KeepManagementFocus(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return !h.m.listBusy })
	if h.m.workspace.focused {
		t.Fatal("empty discovery exposes nonexistent terminal focus")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	if h.m.workspace.focused {
		t.Fatal("F6 focuses a nonexistent terminal")
	}
}

func TestWorkspaceDisconnectKeepsLastObservedScreenReadOnly(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	_ = stream.Close()
	h.until(func() bool {
		return strings.Contains(h.m.View().Content, "只读") && strings.Contains(h.m.View().Content, "shell marker first")
	})
	if h.m.workspace.focused || h.m.View().Cursor != nil {
		t.Fatal("disconnected screen still has terminal input focus")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	if h.m.workspace.focused {
		t.Fatal("F6 focuses a disconnected terminal")
	}
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	h.until(func() bool { return h.m.closed })
}
