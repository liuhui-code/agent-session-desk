package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestWorkspaceWheelScrollsOnlyTheTerminalUnderPointer(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	g := h.m.workspaceGeometry()
	h.send(tea.MouseWheelMsg{X: g.x + 3, Y: g.y + 1, Button: tea.MouseWheelUp})
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "scroll:true,3,1")
	})
	if strings.Contains(ansi.Strip(h.m.View().Content), "› 评审") {
		t.Fatal("terminal wheel moved the list selection")
	}
}

func TestWorkspaceScrollModeShowsHistoryAndNativePrompt(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}, frame: func(v session.Session, w, h int) terminal.Screen {
		return terminal.Screen{Width: w, Height: h, CopyMode: true, ScrollbackOffset: 9, Status: "Search: rare-token", Rows: []terminal.Row{{Runs: []terminal.Run{{Text: "shell marker " + v.ID, Width: len("shell marker " + v.ID)}}}}}
	}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	view := ansi.Strip(h.m.View().Content)
	if !strings.Contains(view, "滚动历史") || !strings.Contains(view, "Search: rare-token") || h.m.View().Cursor != nil {
		t.Fatal("history mode is indistinguishable from live input or hides its native prompt")
	}
}

func TestWorkspaceListWheelSelectsWithoutSendingTerminalKeys(t *testing.T) {
	for _, width := range []int{120, 60, 40} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
			h := newWorkspaceHarness(t, service)
			h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
			h.send(tea.WindowSizeMsg{Width: width, Height: 24})
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			var y int
			for row, line := range strings.Split(ansi.Strip(h.m.View().Content), "\n") {
				if strings.HasPrefix(line, "› 开发") {
					y = row
					break
				}
			}
			h.send(tea.MouseWheelMsg{X: 3, Y: y, Button: tea.MouseWheelDown})
			if !strings.Contains(ansi.Strip(h.m.View().Content), "› 评审") || h.m.View().Cursor != nil {
				t.Fatal("list wheel did not navigate in management focus")
			}
			h.until(func() bool { return h.m.workspace.pending == 0 })
			service.mu.Lock()
			stream := service.streams["first"]
			service.mu.Unlock()
			stream.mu.Lock()
			defer stream.mu.Unlock()
			for _, write := range stream.writes {
				if strings.HasPrefix(write, "scroll:") || strings.HasPrefix(write, "key:") {
					t.Fatalf("list wheel reached child terminal: %s", write)
				}
			}
		})
	}
}

func TestWorkspaceWheelIsIgnoredOutsideContentAndInDialog(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.MouseWheelMsg{X: 5, Y: 1, Button: tea.MouseWheelUp})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "新建终端") })
	h.send(tea.MouseWheelMsg{X: 70, Y: 10, Button: tea.MouseWheelUp})
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	for _, write := range stream.writes {
		if strings.HasPrefix(write, "scroll:") {
			t.Fatal("hidden terminal received a wheel event")
		}
	}
}

func TestWorkspaceScrollDrainsBeforeSwitchAndDoesNotReachNewRun(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	first := service.streams["first"]
	service.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	g := h.m.workspaceGeometry()
	h.send(tea.MouseWheelMsg{X: g.x + 3, Y: g.y + 1, Button: tea.MouseWheelUp})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: 'z', Text: "z"})
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
		return a == "scroll:true,3,1" && b == "text:z"
	})
}
