package ui

import (
	"context"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestRecoveryBlocksLiveUnknownAndUncertainAttempts(t *testing.T) {
	source := session.Session{ID: strings.Repeat("1", 32), RunID: strings.Repeat("2", 32)}
	profile := recovery.Profile{Revision: 1, Declaration: recovery.Declaration{Protocol: "asd.agent-recovery", SchemaVersion: 1, SessionID: source.ID, RunID: source.RunID, AgentID: "fixture", ConversationID: "c", Resume: recovery.Resume{Executable: `C:\agent.exe`, Cwd: `C:\Work`, Args: []string{}}}}
	for _, scenario := range []string{"live", "exited", "unknown", "launch_unknown", "busy"} {
		t.Run(scenario, func(t *testing.T) {
			m := NewWithOptions(context.Background(), fakeService{}, Options{RestoreRecovery: func(context.Context, recovery.Review) (recovery.Outcome, error) {
				t.Fatal("guard launched")
				return recovery.Outcome{}, nil
			}})
			defer m.Close()
			m.workspace = &workspaceState{}
			m.diagnostic.discoveryObserved = true
			d := &recoveryCenter{review: recovery.Review{Source: source, Profile: &profile}}
			switch scenario {
			case "live", "exited":
				row := source
				row.TerminalState = session.TerminalAvailable
				if scenario == "exited" {
					row.TerminalState = session.TerminalExited
				}
				m.sessions = []session.Session{row}
			case "unknown":
				m.diagnostic.discoveryObserved = false
			case "launch_unknown":
				d.review.Attempt = &recovery.Attempt{State: "launch_unknown"}
			case "busy":
				m.workspace.busy = true
			}
			if m.recoveryBlocked(d) == "" {
				t.Fatal("unsafe recovery allowed")
			}
		})
	}
}
func TestRecoveryCloneRejectsOldRecoveryCredentials(t *testing.T) {
	svc := &workspaceFixture{}
	m := NewWithOptions(context.Background(), svc, Options{})
	defer m.Close()
	d := &recoveryCenter{review: recovery.Review{Launch: &session.Launch{Env: []string{"ASD_RECOVERY_CONTEXT_FILE=private"}}}}
	m.recoveryUI.center = d
	if cmd := m.newSessionFromRecovery(d); cmd != nil || m.launchSeed != nil || m.recoveryUI.center != d {
		t.Fatal("old handoff cloned")
	}
}
func TestRecoveryLateReadDoesNotReopenClosedPanel(t *testing.T) {
	m := NewWithOptions(context.Background(), fakeService{}, Options{})
	defer m.Close()
	d := &recoveryCenter{}
	m.recoveryEvent(recoveryReviewed{center: d})
	m.recoveryEvent(recoveryHistoryLoaded{center: d})
	if m.recoveryUI.center != nil {
		t.Fatal("late read reopened panel")
	}
}

func TestRecoveryReattachRetainsExactConnectedRunWithoutStart(t *testing.T) {
	svc := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, svc)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.listBusy })
	w := h.m.workspace
	original, conn, generation := w.tabs[w.active].session, w.conn, w.generation
	d := &recoveryCenter{detail: true, review: recovery.Review{Source: original}}
	h.m.recoveryUI.center = d
	w.focused = false
	w.busy = true
	if h.m.reattachRecoverySource(d) != nil || w.focused || h.m.recoveryUI.center != d {
		t.Fatal("busy reattach changed focus")
	}
	w.busy = false
	h.run(h.m.reattachRecoverySource(d))
	if h.m.recoveryUI.center != nil || !w.focused || w.conn != conn || w.generation != generation || len(svc.starts) != 0 {
		t.Fatal("reattach restarted original")
	}
}
