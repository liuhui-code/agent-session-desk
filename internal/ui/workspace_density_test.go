package ui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestWorkspaceDensitySettingsApplyFullCardsWithoutTerminalInput(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "侧栏密度") })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown}) // Theme is now the first settings row.
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.until(func() bool { return !strings.Contains(ansi.Strip(h.m.View().Content), "侧栏密度") })
	view := ansi.Strip(h.m.View().Content)
	if !strings.Contains("\n"+view, "\n    │ 可连接") || !strings.Contains("\n"+view, "\n    ╰ ") {
		t.Fatalf("full card did not render two metadata lines:\n%s", view)
	}
	service.mu.Lock()
	first := service.streams["first"]
	service.mu.Unlock()
	first.mu.Lock()
	defer first.mu.Unlock()
	for _, write := range first.writes {
		if strings.HasPrefix(write, "text:") || strings.HasPrefix(write, "key:") {
			t.Fatalf("density settings leaked terminal input: %s", write)
		}
	}
}

func TestWorkspaceDensitySettingsSelectedChoiceRemainsVisibleInShortNarrowView(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "侧栏密度") })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown}) // Theme is now the first settings row.
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	h.until(func() bool { return h.m.sidebarDensity == "auto" })
	h.send(tea.WindowSizeMsg{Width: 27, Height: 8})
	view := ansi.Strip(h.m.View().Content)
	if !strings.Contains(view, ">Auto") {
		t.Fatalf("selected choice cropped from short settings view:\n%s", view)
	}
}

func TestWorkspaceDensitySavesSerializeLatestChoiceBeforeClosing(t *testing.T) {
	started := make(chan string, 4)
	release := make(chan struct{}, 2)
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{SaveSidebarDensity: func(ctx context.Context, value string) error {
		started <- value
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "侧栏密度") })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown}) // Theme is now the first settings row.
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	select {
	case got := <-started:
		if got != "full" {
			t.Fatalf("first save: %s", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("save did not start")
	}
	for i := 0; i < 3; i++ {
		h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if !strings.Contains(h.m.View().Content, "正在保存") {
		t.Fatal("closing discarded an accepted pending choice")
	}
	select {
	case value := <-started:
		t.Fatalf("concurrent preference save: %s", value)
	default:
	}
	release <- struct{}{}
	h.until(func() bool { return h.m.sidebarDensity == "full" })
	select {
	case got := <-started:
		if got != "auto" {
			t.Fatalf("latest choice not coalesced: %s", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("latest save did not start")
	}
	release <- struct{}{}
	h.until(func() bool {
		return strings.Contains(h.m.View().Content, "shell marker first") && !strings.Contains(h.m.View().Content, "侧栏密度")
	})
	if h.m.sidebarDensity != "auto" {
		t.Fatal("closed settings before applying the last successful save")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "text:S")
	})
	if strings.Contains(h.m.View().Content, "侧栏密度") {
		t.Fatal("terminal S opened management settings")
	}
}

func TestWorkspaceDensitySaveFailureKeepsAppliedChoiceAndVisibleError(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{SaveSidebarDensity: func(context.Context, string) error { return errors.New("synthetic write refusal") }})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "侧栏密度") })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown}) // Theme is now the first settings row.
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "未保存") })
	h.send(tea.WindowSizeMsg{Width: 27, Height: 8})
	if h.m.sidebarDensity != "compact" || !strings.Contains(h.m.View().Content, "未保存") {
		t.Fatal("failed save changed applied density or hid error")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if strings.Contains(h.m.View().Content, "侧栏密度") {
		t.Fatal("save failure trapped the settings dialog")
	}
}

func TestWorkspaceDensityOpeningWaitsForAcceptedTerminalWrite(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.until(func() bool { return h.m.workspace.pending == 0 })
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	gate := make(chan struct{})
	stream.mu.Lock()
	stream.gate = gate
	stream.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'z', Text: "z"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	if strings.Contains(h.m.View().Content, "侧栏密度") {
		t.Fatal("settings disconnected before accepted input drained")
	}
	close(gate)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "侧栏密度") })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown}) // Theme is now the first settings row.
	h.send(tea.PasteMsg{Content: "never sent"})
	h.send(tea.MouseClickMsg{X: 90, Y: 10, Button: tea.MouseLeft})
	h.send(pollMsg{generation: h.m.pollGeneration})
	h.until(func() bool { return !h.m.listBusy })
	service.mu.Lock()
	opened := len(service.opened)
	service.mu.Unlock()
	if opened != 1 {
		t.Fatal("discovery reopened a hidden terminal while settings owned focus")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	var inputs []string
	for _, write := range stream.writes {
		if strings.HasPrefix(write, "text:") || strings.HasPrefix(write, "key:") || strings.HasPrefix(write, "paste:") {
			inputs = append(inputs, write)
		}
	}
	if strings.Join(inputs, ",") != "text:z" {
		t.Fatalf("input lost or settings pasted into terminal: %v", stream.writes)
	}
}

func TestWorkspaceDensityAutoUsesCurrentBudgetWithoutOscillation(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{SidebarDensity: "auto", SidebarMode: "flat"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	for _, tc := range []struct{ width, height, metadata int }{{28, 6, 4}, {28, 5, 2}, {28, 3, 0}, {27, 10, 0}, {28, 6, 4}, {28, 5, 2}, {28, 3, 0}} {
		view := ansi.Strip(h.m.renderSessionList(tc.width, tc.height))
		if view != ansi.Strip(h.m.renderSessionList(tc.width, tc.height)) {
			t.Fatal("auto density oscillated without a geometry change")
		}
		lines := strings.Split(view, "\n")
		metadata := 0
		for _, line := range lines {
			if strings.HasPrefix(line, "  │ ") || strings.HasPrefix(line, "  ╰ ") {
				metadata++
			}
			if ansi.StringWidth(line) > tc.width {
				t.Fatal("card exceeded width")
			}
		}
		if metadata != tc.metadata || len(lines) > tc.height {
			t.Fatalf("%+v: unexpected card budget:\n%s", tc, view)
		}
	}
}

func TestWorkspaceDensitySpaceCyclesFromAutoToFull(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service, Options{SidebarDensity: "auto"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "侧栏密度") })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown}) // Theme is now the first settings row.
	h.send(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	h.until(func() bool { return h.m.sidebarDensity == "full" })
	if !strings.Contains(ansi.Strip(h.m.View().Content), ">Full") {
		t.Fatal("Space did not wrap to first radio option")
	}
}
