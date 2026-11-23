package ui

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func clickWorkspaceListRow(t *testing.T, h *workspaceHarness, title string) {
	t.Helper()
	for y, line := range strings.Split(ansi.Strip(h.m.View().Content), "\n") {
		if y < 2 {
			continue
		} // The header and Tab row are not session-list rows.
		plain := strings.TrimLeft(line, " ▶┃")
		if strings.HasPrefix(plain, title) {
			h.send(tea.MouseClickMsg{X: 3, Y: y, Button: tea.MouseLeft})
			return
		}
	}
	t.Fatalf("no visible list row for %q", title)
}

func TestWorkspaceListClickSelectsThenEnterOpensExactRun(t *testing.T) {
	for _, width := range []int{120, 60, 40} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
			h := newWorkspaceHarness(t, service)
			h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
			h.send(tea.WindowSizeMsg{Width: width, Height: 24})
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			clickWorkspaceListRow(t, h, "评审")
			if !strings.Contains(ansi.Strip(h.m.View().Content), "▶ 评审") || h.m.View().Cursor != nil {
				t.Fatal("click did not select the visible session in management focus")
			}
			service.mu.Lock()
			opened := strings.Join(service.opened, ",")
			service.mu.Unlock()
			if opened != "first" {
				t.Fatalf("single click opened a terminal: %s", opened)
			}
			h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
			h.until(func() bool {
				return strings.Contains(h.m.View().Content, "shell marker second") && h.m.View().Cursor != nil
			})
			h.send(tea.KeyPressMsg{Code: 'z', Text: "z"})
			h.until(func() bool {
				service.mu.Lock()
				second := service.streams["second"]
				service.mu.Unlock()
				second.mu.Lock()
				defer second.mu.Unlock()
				return strings.Contains(strings.Join(second.writes, ","), "text:z")
			})
			service.mu.Lock()
			first := service.streams["first"]
			service.mu.Unlock()
			first.mu.Lock()
			defer first.mu.Unlock()
			if strings.Contains(strings.Join(first.writes, ","), "text:z") {
				t.Fatal("input reached the previous Run")
			}
		})
	}
}

func TestWorkspaceListMouseUsesVisibleRowsAndIgnoresIndicators(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) {
		rows := sampleSessions()[:1]
		for i := 1; i < 30; i++ {
			rows = append(rows, session.Session{ID: fmt.Sprint(i), RunID: fmt.Sprint(i), Title: fmt.Sprintf("task-%02d", i), TerminalState: session.TerminalAvailable})
		}
		return rows, nil
	}}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.WindowSizeMsg{Width: 120, Height: 14})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	for i := 0; i < 15; i++ {
		h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	before := ansi.Strip(h.m.View().Content)
	clickWorkspaceListRow(t, h, "task-14")
	after := ansi.Strip(h.m.View().Content)
	if !strings.Contains(after, "▶ task-14") {
		t.Fatal("scrolled click selected a different row")
	}
	beforeY, afterY := -1, -1
	for y, line := range strings.Split(before, "\n") {
		if strings.Contains(line, "task-14") {
			beforeY = y
		}
	}
	for y, line := range strings.Split(after, "\n") {
		if strings.Contains(line, "task-14") {
			afterY = y
		}
	}
	if beforeY != afterY {
		t.Fatal("clicked list row jumped to another viewport position")
	}
	for _, marker := range []string{"上方", "下方"} {
		found := false
		for y, line := range strings.Split(after, "\n") {
			if strings.Contains(line, marker) {
				found = true
				h.send(tea.MouseClickMsg{X: 3, Y: y, Button: tea.MouseLeft})
				if !strings.Contains(ansi.Strip(h.m.View().Content), "▶ task-14") {
					t.Fatalf("%s indicator selected a session", marker)
				}
				break
			}
		}
		if !found {
			t.Fatalf("missing %s indicator", marker)
		}
	}
}

func TestWorkspaceListDoubleClickOpensWithoutInjectingEnter(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	clickWorkspaceListRow(t, h, "评审")
	clickWorkspaceListRow(t, h, "评审")
	h.until(func() bool {
		return strings.Contains(h.m.View().Content, "shell marker second") && h.m.View().Cursor != nil
	})
	service.mu.Lock()
	second := service.streams["second"]
	service.mu.Unlock()
	second.mu.Lock()
	defer second.mu.Unlock()
	for _, write := range second.writes {
		if strings.HasPrefix(write, "key:") || strings.HasPrefix(write, "text:") {
			t.Fatalf("double click injected child input: %s", write)
		}
	}
}

func TestWorkspaceListDoubleClickDoesNotActivateAReplacementRun(t *testing.T) {
	var replaced atomic.Bool
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) {
		rows := sampleSessions()
		if replaced.Load() {
			rows[1].RunID += "-replacement"
			rows[1].Title = "评审重启"
		}
		return rows, nil
	}}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	clickWorkspaceListRow(t, h, "评审")
	replaced.Store(true)
	h.send(pollMsg{generation: h.m.pollGeneration})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "评审重启") })
	clickWorkspaceListRow(t, h, "评审重启")
	if h.m.View().Cursor != nil || !strings.Contains(ansi.Strip(h.m.View().Content), "▶ 评审重启") {
		t.Fatal("two different Runs were treated as one double click")
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if strings.Join(service.opened, ",") != "first" {
		t.Fatal("replacement Run opened without its own double click or Enter")
	}
}

func TestWorkspaceListDoubleClickWhileConnectingKeepsManagementFocus(t *testing.T) {
	gate := make(chan struct{})
	service := &workspaceFixture{openGate: gate, fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "正在连接原终端") })
	// The list is visible in the compact sidebar even while the active Tab opens.
	clickWorkspaceListRow(t, h, "评审")
	clickWorkspaceListRow(t, h, "评审")
	close(gate)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	if h.m.View().Cursor != nil {
		t.Fatal("refused switch gave input focus to the previous Run")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		return strings.Contains(h.m.View().Content, "shell marker second") && h.m.View().Cursor != nil
	})
}
