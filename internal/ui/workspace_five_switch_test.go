package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestFiveSessionCycleKeepsTargetScreenBeforeSavingOutgoing(t *testing.T) {
	rows := sampleSessions()
	for i := 2; i < 5; i++ {
		rows = append(rows, session.Session{ID: fmt.Sprint(i), RunID: fmt.Sprint("run-", i), BackendID: fmt.Sprint("backend-", i), Title: fmt.Sprint(i), TerminalState: session.TerminalAvailable})
	}
	svc := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return rows, nil }}}
	h := newWorkspaceHarness(t, svc)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	for i := 1; i < 5; i++ {
		h.run(h.m.activateWorkspaceNow(i))
		h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	}
	for n := 0; n < 15; n++ {
		i := n % 5
		cmd := h.m.activateWorkspaceNow(i)
		view := h.m.View()
		if strings.Contains(view.Content, "Connecting to the original terminal") || !strings.Contains(view.Content, "shell marker "+rows[i].ID) {
			t.Fatalf("cycle %d evicted incoming original %s before lookup", n, rows[i].ID)
		}
		if view.Cursor != nil || h.m.workspace.conn != nil {
			t.Fatal("last screen gained native authority")
		}
		total := 0
		for _, v := range h.m.workspace.snapshots {
			total += v.bytes
		}
		if total > workspaceSnapshotBytes || len(h.m.workspace.snapshots) > workspaceSnapshotLimit {
			t.Fatal("retention budget increased")
		}
		h.run(cmd)
		h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	}
}
