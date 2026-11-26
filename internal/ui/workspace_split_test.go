package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestWorkspaceSplitKeysOnlyAdjustInManagementFocus(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: '<', Text: "<"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	if h.m.workspaceGeometry().left != 40 {
		t.Fatal("default management split changed")
	}
	h.send(tea.KeyPressMsg{Code: '<', Text: "<"})
	if h.m.workspaceGeometry().left != 46 {
		t.Fatal("management < did not shrink terminal by five percent")
	}
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "size:67,30")
	})
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if strings.Count(strings.Join(stream.writes, ","), "text:<") != 1 {
		t.Fatal("management split key leaked to terminal")
	}
}

func TestWorkspaceDividerDragChangesSplitWithoutTerminalInput(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.bodyX + g.left + 1, Y: g.y + 1, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: 60, Y: g.y + 1, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: 60, Y: g.y + 1, Button: tea.MouseLeft})
	if h.m.workspaceGeometry().left != 58 {
		t.Fatal("divider did not follow drag to fifty percent")
	}
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "size:55,31")
	})
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	for _, write := range stream.writes {
		if strings.HasPrefix(write, "text:") || strings.HasPrefix(write, "key:") {
			t.Fatal("divider drag injected child input")
		}
	}
}

func TestWorkspaceSplitBoundsResetAndClickWithoutMovement(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.bodyX + g.left + 1, Y: g.y + 1, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: g.bodyX + g.left + 1, Y: g.y + 1, Button: tea.MouseLeft})
	if h.m.workspaceGeometry().left != g.left {
		t.Fatal("plain divider click changed split")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	for range 30 {
		h.send(tea.KeyPressMsg{Code: '<', Text: "<"})
	}
	if h.m.workspaceGeometry().left != 104 {
		t.Fatal("maximum split bound not applied")
	}
	for range 30 {
		h.send(tea.KeyPressMsg{Code: '>', Text: ">"})
	}
	if h.m.workspaceGeometry().left != 11 {
		t.Fatal("minimum split bound not applied")
	}
	h.send(tea.KeyPressMsg{Code: '=', Text: "="})
	if h.m.workspaceGeometry().left != 40 {
		t.Fatal("reset did not restore management default")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	if h.m.workspaceGeometry().left != 20 {
		t.Fatal("reset did not restore automatic compact sidebar")
	}
}
