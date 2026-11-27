package ui

import (
	"context"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
	"strings"
	"testing"
)

func TestWorkspaceOpenPublishesFirstScreenTogether(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	cmd := h.m.activateWorkspaceNow(1)
	if cmd == nil {
		t.Fatal("no switch")
	}
	h.send(cmd())
	if !strings.Contains(h.m.View().Content, "shell marker second") {
		t.Fatal("successful open published an empty frame before its already available screen")
	}
	if strings.Contains(h.m.View().Content, "shell marker first") {
		t.Fatal("previous Run remained under new target")
	}
}

func TestWorkspaceReturnShowsItsOwnLastScreenWhileRevalidating(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.run(h.m.activateWorkspaceNow(1))
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	cmd := h.m.activateWorkspaceNow(0)
	view := h.m.View()
	if !strings.Contains(view.Content, "shell marker first") || strings.Contains(view.Content, "shell marker second") || strings.Contains(view.Content, "Connecting to the original terminal") {
		t.Fatal("return replaced the target's own last screen with connecting page")
	}
	if view.Cursor != nil || h.m.workspace.conn != nil || !strings.Contains(view.Content, "last observed") {
		t.Fatal("cached screen acquired live input authority or lacked freshness label")
	}
	h.send(cmd())
	if !strings.Contains(h.m.View().Content, "shell marker first") || strings.Contains(h.m.View().Content, "last observed") {
		t.Fatal("fresh first frame did not replace read-only snapshot atomically")
	}
}

func TestWorkspaceSnapshotBudgetAndIdentity(t *testing.T) {
	h, _ := attentionUIFixture(t)
	w := h.m.workspace
	for i := range 12 {
		w.tabs[w.active].session.RunID = fmt.Sprint(i)
		w.screen = terminal.Screen{Width: 80, Height: 24, Rows: []terminal.Row{{Runs: []terminal.Run{{Text: strings.Repeat("x", 30000)}}}}}
		w.rememberScreen()
		total := 0
		for _, v := range w.snapshots {
			total += v.bytes
		}
		if len(w.snapshots) > workspaceSnapshotLimit || total > workspaceSnapshotBytes {
			t.Fatal("snapshot cache grew beyond its global budget")
		}
	}
	saved := w.lastScreen()
	if saved.Width == 0 {
		t.Fatal("current Run missing")
	}
	w.tabs[w.active].session.RunID = "replacement"
	if w.lastScreen().Width != 0 {
		t.Fatal("snapshot inherited by another Run")
	}
	w.tabs[w.active].session.RunID = "11"
	w.tabs[w.active].session.TerminalState = session.TerminalExited
	if w.lastScreen().Width != 0 {
		t.Fatal("unavailable terminal acquired a cached connecting screen")
	}
	w.tabs[w.active].session.TerminalState = session.TerminalAvailable
	before := len(w.snapshots)
	w.screen.Rows[0].Runs[0].Text = strings.Repeat("x", workspaceSnapshotBytes)
	w.rememberScreen()
	if len(w.snapshots) != before {
		t.Fatal("oversized screen was cached")
	}
	w.closeViews()
	if w.snapshots != nil {
		t.Fatal("UI close retained cache")
	}
}

func TestConnectingPausesOptionalRefreshWork(t *testing.T) {
	h, _ := attentionUIFixture(t)
	h.m.workspace.busy = true
	if h.m.sampleActivity() != nil {
		t.Fatal("sampling started ahead of connection")
	}
	h.m.pollWorkspace()
	if h.m.listBusy {
		t.Fatal("poll discovery started ahead of connection")
	}
	ctx, cancel := context.WithCancel(context.Background())
	h.m.activity.cancel = cancel
	h.m.pauseActivity()
	if ctx.Err() == nil {
		t.Fatal("inflight optional sampling was not cancelled")
	}
}
