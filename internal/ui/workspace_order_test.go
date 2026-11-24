package ui

import (
	"context"
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// Logical view order is intentionally independent of the one visible sidebar.
func tabOrderRow(h *workspaceHarness) string {
	var names []string
	for _, tab := range h.m.workspace.tabs {
		names = append(names, tab.session.Title)
	}
	return strings.Join(names, " ")
}

func TestWorkspaceTabOrderKeepsPendingSwitchBoundToOriginalRun(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	first := service.streams["first"]
	service.mu.Unlock()
	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	defer release()
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'a', Text: "a"})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.m.moveWorkspaceTab(h.m.workspace.tabs[0].viewID, 1)
	h.send(tea.KeyPressMsg{Code: 'b', Text: "b"})
	release()
	h.until(func() bool {
		service.mu.Lock()
		second := service.streams["second"]
		service.mu.Unlock()
		if second == nil {
			return false
		}
		second.mu.Lock()
		defer second.mu.Unlock()
		return strings.Contains(strings.Join(second.writes, ","), "text:b") && strings.Contains(h.m.View().Content, "shell marker second")
	})
	first.mu.Lock()
	writes := strings.Join(first.writes, ",")
	first.mu.Unlock()
	if !strings.Contains(writes, "text:a") || strings.Contains(writes, "text:b") || !strings.Contains(h.m.View().Content, "shell marker second") {
		t.Fatalf("pending switch observation: first writes=%q; screen=%s", writes, ansi.Strip(h.m.View().Content))
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: ']', Text: "]"})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgUp, Mod: tea.ModCtrl})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	defer service.mu.Unlock()
	if strings.Join(service.opened, ",") != "first,second,first" || len(service.starts) != 0 {
		t.Fatal("navigation after sorting did not follow the original Run identities")
	}
}

func TestWorkspaceUnequalLogicalTitlesKeepKeyboardOrderAtEdges(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) {
		rows := sampleSessions()
		rows[0].Title, rows[1].Title = "短", "很长的评审标题"
		return rows, nil
	}}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: ']', Text: "]"})
	row := tabOrderRow(h)
	for range 3 {
		h.send(tea.KeyPressMsg{Code: ']', Text: "]"})
	}
	if tabOrderRow(h) != row {
		t.Fatal("right edge reordered after reaching the last position")
	}
	h.send(tea.KeyPressMsg{Code: '[', Text: "["})
	row = tabOrderRow(h)
	h.send(tea.KeyPressMsg{Code: '[', Text: "["})
	if tabOrderRow(h) != row || strings.Index(row, "短") > strings.Index(row, "很长") {
		t.Fatal("left edge wrapped or could not restore the original order")
	}
}

func TestWorkspaceManagementReordersActiveTabWithoutReconnecting(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: ']', Text: "]"})
	row := tabOrderRow(h)
	if !strings.Contains(row, "评审") || !strings.Contains(row, "开发") || strings.Index(row, "评审") > strings.Index(row, "开发") {
		t.Fatal("management ] did not move the active Tab to the right")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '[', Text: "["})
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "text:[")
	})
	if tabOrderRow(h) != row || !strings.Contains(h.m.View().Content, "shell marker first") {
		t.Fatal("terminal [ reordered the Tab or changed its terminal")
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if strings.Join(service.opened, ",") != "first" || len(service.starts) != 0 {
		t.Fatal("reordering reconnected or started a Run")
	}
}

func TestRemovedTabRowHasNoPhantomDragTargets(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	row := tabOrderRow(h)
	x := ansi.StringWidth(row[:strings.Index(row, "开发")])
	destination := ansi.StringWidth(row[:strings.Index(row, "评审")])
	h.send(tea.MouseClickMsg{X: x, Y: 1, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: destination, Y: 1, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: destination, Y: 1, Button: tea.MouseLeft})
	row = tabOrderRow(h)
	if strings.Index(row, "评审") < strings.Index(row, "开发") {
		t.Fatal("removed top row still reordered a Tab")
	}
	if !h.m.workspace.focused {
		h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	}
	h.send(tea.KeyPressMsg{Code: 'a', Text: "a"})
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "text:a")
	})
	service.mu.Lock()
	defer service.mu.Unlock()
	if strings.Join(service.opened, ",") != "first" || len(service.starts) != 0 {
		t.Fatal("drag reconnected the active terminal or started a Run")
	}
}

func TestWorkspaceTabDragStopsAfterWheelOrOtherInput(t *testing.T) {
	for _, input := range []struct {
		name string
		msg  tea.Msg
	}{
		{"wheel", tea.MouseWheelMsg{X: 0, Y: 1, Button: tea.MouseWheelDown}},
		{"paste", tea.PasteMsg{Content: "safe"}},
		{"key", tea.KeyPressMsg{Code: tea.KeyF6}},
		{"resize", tea.WindowSizeMsg{Width: 120, Height: 32}},
		{"left-tab-row", tea.MouseMotionMsg{X: 0, Y: 2, Button: tea.MouseLeft}},
		{"released", tea.MouseReleaseMsg{X: 1, Y: 1, Button: tea.MouseLeft}},
	} {
		t.Run(input.name, func(t *testing.T) {
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
			h := newWorkspaceHarness(t, service)
			h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
			row := tabOrderRow(h)
			x := ansi.StringWidth(row[:strings.Index(row, "开发")])
			destination := ansi.StringWidth(row[:strings.Index(row, "评审")])
			h.send(tea.MouseClickMsg{X: x, Y: 1, Button: tea.MouseLeft})
			h.send(input.msg)
			h.send(tea.MouseMotionMsg{X: destination, Y: 1, Button: tea.MouseLeft})
			if tabOrderRow(h) != row {
				t.Fatal("stale gesture reordered a Tab after another input")
			}
		})
	}
}
