package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestDefaultGroupedSidebarFoldsWithoutClosingTabOrChangingTerminalInput(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	if !strings.Contains(ansi.Strip(h.m.View().Content), "▾ 我的会话 (2)") {
		t.Fatal("default sidebar lost the pinned upstream group header")
	}
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "draft-preserved"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyUp})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "▸ 我的会话 (2)") })
	view := ansi.Strip(h.m.View().Content)
	if strings.Contains(view, "可连接") || !strings.Contains(view, "shell marker first") {
		t.Fatal("fold did not hide children or closed the active terminal")
	}
	h.send(tea.KeyPressMsg{Code: 'v', Mod: tea.ModAlt})
	h.until(func() bool { return !strings.Contains(ansi.Strip(h.m.View().Content), "我的会话") })
	h.send(tea.KeyPressMsg{Code: 'v', Mod: tea.ModAlt})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "▸ 我的会话 (2)") })
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "▾ 我的会话 (2)") })
	stream.mu.Lock()
	defer stream.mu.Unlock()
	writes := strings.Join(stream.writes, ",")
	if strings.Count(writes, "text:draft-preserved") != 1 || strings.Contains(writes, "key:") || strings.Contains(writes, "text:v") {
		t.Fatalf("group navigation leaked or lost input: %v", stream.writes)
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.opened) != 1 || len(service.starts) != 0 {
		t.Fatal("view changes created or reconnected terminal processes")
	}
}

func TestGroupedHeaderMouseClickFoldsInsteadOfSelectingFirstSession(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	row := -1
	for y, line := range strings.Split(ansi.Strip(h.m.View().Content), "\n") {
		if strings.Contains(line, "▾ 我的会话") {
			row = y
			break
		}
	}
	if row < 0 {
		t.Fatal("default group not visible")
	}
	h.send(tea.MouseClickMsg{Button: tea.MouseLeft, X: 2, Y: row})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "▸ 我的会话") })
	if !strings.Contains(h.m.View().Content, "shell marker first") {
		t.Fatal("header click disconnected the terminal")
	}
}

func TestGroupedMetadataClickOpensTheExactSessionBelowTheHeader(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	row := -1
	for y, line := range strings.Split(ansi.Strip(h.m.View().Content), "\n") {
		if strings.HasPrefix(strings.TrimLeft(line, " ▶┃"), "评审") {
			row = y + 1
			break
		}
	}
	if row < 0 {
		t.Fatal("second card missing")
	}
	h.send(tea.MouseClickMsg{Button: tea.MouseLeft, X: 3, Y: row})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
}

func TestSidebarNavigationAndDensityShareOneSerializedPreferenceWriter(t *testing.T) {
	started := make(chan config.UISettings, 3)
	release := make(chan struct{}, 3)
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{SaveUISettings: func(ctx context.Context, s config.UISettings) error {
		started <- s
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'v', Mod: tea.ModAlt})
	select {
	case s := <-started:
		if s.SidebarMode != "flat" || s.SidebarDensity != "compact" {
			t.Fatalf("first write %+v", s)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("navigation save not started")
	}
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "侧栏密度") })
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	select {
	case s := <-started:
		t.Fatalf("concurrent whole-document write: %+v", s)
	default:
	}
	release <- struct{}{}
	h.until(func() bool { return h.m.sidebarMode == "flat" })
	select {
	case s := <-started:
		if s.SidebarMode != "flat" || s.SidebarDensity != "full" {
			t.Fatalf("density save lost navigation: %+v", s)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("density save not started")
	}
	release <- struct{}{}
	h.until(func() bool {
		return !strings.Contains(h.m.View().Content, "侧栏密度") && strings.Contains(h.m.View().Content, "会话 平铺")
	})
}

func TestGroupedVariableHeightWindowKeepsTheSelectedTargetVisibleAndClickable(t *testing.T) {
	for _, density := range []string{"full", "compact", "minimal", "auto"} {
		t.Run(density, func(t *testing.T) {
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) {
				var items []session.Session
				for i := 0; i < 30; i++ {
					id := fmt.Sprintf("target-%02d", i)
					items = append(items, session.Session{ID: id, RunID: id, Title: id, TerminalState: session.TerminalAvailable})
				}
				return items, nil
			}}}
			h := newWorkspaceHarness(t, service, Options{SidebarDensity: density})
			h.until(func() bool { return h.m.workspace.conn != nil })
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			for i := 0; i < 15; i++ {
				h.send(tea.KeyPressMsg{Code: tea.KeyDown})
			}
			for _, size := range [][2]int{{120, 14}, {63, 10}, {49, 9}, {27, 8}, {80, 12}} {
				h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
				view := ansi.Strip(h.m.View().Content)
				if !strings.Contains(view, "▶ target-15") {
					t.Fatalf("selected row lost at %v:\n%s", size, view)
				}
				for _, line := range strings.Split(view, "\n") {
					if ansi.StringWidth(line) != size[0] {
						t.Fatal("viewport width exceeded")
					}
				}
				clickWorkspaceListRow(t, h, "target-15")
				if !strings.Contains(ansi.Strip(h.m.View().Content), "▶ target-15") {
					t.Fatal("visible row click selected a different target")
				}
			}
			h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
			h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker target-15") })
		})
	}
}

func TestManagementQuitWaitsForAnAcceptedSidebarPreference(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{SaveUISettings: func(ctx context.Context, s config.UISettings) error {
		close(started)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	// Drain the resize caused by focus change, so the terminal-input fence
	// cannot accidentally hide a missing preference-save fence.
	h.until(func() bool { return h.m.workspace.pending == 0 && !h.m.workspace.busy })
	h.send(tea.KeyPressMsg{Code: 'v', Mod: tea.ModAlt})
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("save not started")
	}
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if h.m.closed {
		t.Fatal("quit cancelled an accepted preference before its result")
	}
	close(release)
	h.until(func() bool { return h.m.closed })
}

func TestSidebarSaveFailureCancelsWaitingQuitAndKeepsTheErrorVisible(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{SaveUISettings: func(ctx context.Context, _ config.UISettings) error {
		close(started)
		select {
		case <-release:
			return errors.New("test preference disk failure")
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.until(func() bool { return h.m.workspace.pending == 0 && !h.m.workspace.busy })
	h.send(tea.KeyPressMsg{Code: 'v', Mod: tea.ModAlt})
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("save not started")
	}
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	close(release)
	h.until(func() bool { return h.m.sidebarPreferenceError != nil })
	if h.m.closed || h.m.workspace.transition != workspaceIdle || h.m.sidebarMode != "grouped" {
		t.Fatal("failed preference silently quit or changed the saved mode")
	}
	if !strings.Contains(h.m.View().Content, "test preference disk failure") {
		t.Fatal("save failure not visible")
	}
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if !h.m.closed {
		t.Fatal("explicit quit after reviewing failure did not exit")
	}
}

func TestRepeatedLeftWhileSavingKeepsTheDefaultGroupCollapsed(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{SaveUISettings: func(ctx context.Context, s config.UISettings) error {
		if !s.DefaultGroupCollapsed {
			t.Error("repeated Left reversed collapse")
		}
		close(started)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '1', Text: "1"})
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("save not started")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	close(release)
	h.until(func() bool { return !h.m.preferences.busy })
	if !h.m.sidebarCollapsed || !h.m.preferences.saved.DefaultGroupCollapsed {
		t.Fatal("repeated Left expanded the default group")
	}
}

func TestTerminalAltVStillGoesToTheOriginalTerminal(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	h.send(tea.KeyPressMsg{Code: 'v', Mod: tea.ModAlt})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if !strings.Contains(strings.Join(stream.writes, ","), "key:M-v") || h.m.sidebarMode != "grouped" || h.m.preferences.busy {
		t.Fatalf("terminal key was intercepted by navigation: %v", stream.writes)
	}
}
