package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"strings"
	"testing"
)

func TestWorkspacePastePreservesUnicodeAndNewlinesWithoutEnter(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.PasteMsg{Content: "中文 😀\nsecond line"})
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Join(stream.writes, ",") == "paste:中文 😀\nsecond line"
	})
}

func TestWorkspaceOpeningHistoryDrainsPreviouslyTypedInput(t *testing.T) {
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
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'h', Text: "h"})
	close(gate)
	h.until(func() bool { return h.m.history.open && h.m.workspace.conn == nil })
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if len(stream.writes) == 0 || stream.writes[0] != "text:a" {
		t.Fatalf("history cancelled accepted input: %v", stream.writes)
	}
}

func TestWorkspacePollRefreshesOnlyVisibleActiveTerminal(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(pollMsg{h.m.pollGeneration})
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "refresh")
	})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '?', Text: "?"})
	h.send(pollMsg{h.m.pollGeneration})
	h.until(func() bool { return h.m.help && h.m.workspace.conn == nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.until(func() bool { return h.m.workspace.conn != nil })
}

func TestWorkspaceSidebarSelectionOpensOriginalViewWithoutSendingInput(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.MouseClickMsg{X: 5, Y: h.m.workspaceGeometry().y + 3, Button: tea.MouseLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	if h.m.View().MouseMode != tea.MouseModeAllMotion {
		t.Fatal("mouse events not enabled")
	}
}

func TestWorkspaceCloseWaitsForInputBufferedDuringConnect(t *testing.T) {
	gate := make(chan struct{})
	service := &workspaceFixture{openGate: gate, fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.busy })
	h.send(tea.KeyPressMsg{Code: 'a', Text: "a"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	close(gate)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	if stream == nil {
		t.Fatal("Tab close cancelled buffered input before opening original terminal")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if !strings.Contains(strings.Join(stream.writes, ","), "text:a") {
		t.Fatalf("buffered input lost: %v", stream.writes)
	}
}
