package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestWorkspaceSidebarDefaultCardsSeparateTerminalMetadata(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) {
		rows := sampleSessions()
		rows[0].Command = `C:\Synthetic Tools\PowerShell\pwsh.exe`
		return rows, nil
	}}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.WindowSizeMsg{Width: 120, Height: 24})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	view := ansi.Strip(h.m.View().Content)
	lines := strings.Split(view, "\n")
	found := false
	for y, line := range lines {
		if strings.HasPrefix(line, "  ┃ 开发") {
			found = true
			if y+1 >= len(lines) || !strings.HasPrefix(lines[y+1], "    ╰ 可连接 · pwsh.exe") {
				t.Fatalf("terminal metadata did not occupy its own card line:\n%s", view)
			}
		}
	}
	if !found {
		t.Fatalf("active terminal card marker missing:\n%s", view)
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	if !strings.Contains(ansi.Strip(h.m.View().Content), "▶ 评审") {
		t.Fatal("management selection must remain distinct from active terminal")
	}
}

func TestWorkspaceSidebarMetadataClickSelectsItsOwnSession(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) {
		rows := sampleSessions()
		rows[0].Command = "pwsh-first.exe"
		rows[1].Command = "pwsh-second.exe"
		return rows, nil
	}}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.WindowSizeMsg{Width: 120, Height: 24})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	for y, line := range strings.Split(ansi.Strip(h.m.View().Content), "\n") {
		if strings.Contains(line, "pwsh-second.exe") {
			h.send(tea.MouseClickMsg{X: 3, Y: y, Button: tea.MouseLeft})
			if !strings.Contains(ansi.Strip(h.m.View().Content), "▶ 评审") {
				t.Fatal("metadata click did not select its own card")
			}
			return
		}
	}
	t.Fatal("missing second card metadata")
}

func TestWorkspaceSidebarShowsBothCardsWhenTheyExactlyFit(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.WindowSizeMsg{Width: 60, Height: 24})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	if !strings.Contains("\n"+ansi.Strip(h.m.View().Content), "\n    评审") {
		t.Fatal("below-indicator reservation hid a card although both cards fit")
	}
}

func TestWorkspaceSidebarKeepsSelectedCardVisibleAcrossWidthAndHeight(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) {
		rows := sampleSessions()
		for i := 2; i < 30; i++ {
			rows = append(rows, session.Session{ID: fmt.Sprint(i), RunID: fmt.Sprint(i), Title: fmt.Sprintf("task-%02d", i), TerminalState: session.TerminalAvailable})
		}
		return rows, nil
	}}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	for _, width := range []int{27, 28, 40, 60, 79, 80, 120} {
		for _, height := range []int{8, 9, 10, 12, 14, 24, 32} {
			h.send(tea.WindowSizeMsg{Width: width, Height: height})
			view := ansi.Strip(h.m.View().Content)
			if !strings.Contains(view, "▶ task-29") {
				t.Fatalf("selected card hidden at %dx%d:\n%s", width, height, view)
			}
			for _, line := range strings.Split(view, "\n") {
				if ansi.StringWidth(line) > width {
					t.Fatalf("card exceeded %d columns: %q", width, line)
				}
			}
		}
	}
}
