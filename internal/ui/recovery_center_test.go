package ui

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestRecoveryCenterReviewsWithoutChangingOriginalTerminalOrStartingAgent(t *testing.T) {
	source := sampleSessions()[0]
	svc := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }, history: func(context.Context, int) ([]session.HistoryRecord, error) {
		return []session.HistoryRecord{{Session: source}}, nil
	}}}
	h := newWorkspaceHarness(t, svc, Options{SidebarMode: "flat", ReviewRecovery: func(context.Context, session.Session) (recovery.Review, error) {
		return recovery.Review{Source: source}, nil
	}})
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.listBusy })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	conn, generation := h.m.workspace.conn, h.m.workspace.generation
	h.send(tea.KeyPressMsg{Code: 'R', Text: "R"})
	h.until(func() bool {
		return strings.Contains(h.m.View().Content, "Recover sessions") && !strings.Contains(h.m.View().Content, "Loading")
	})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "No Agent recovery declaration") })
	h.send(tea.PasteMsg{Content: "do not execute"})
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	if len(svc.starts) != 0 || h.m.workspace.conn != conn || h.m.workspace.generation != generation {
		t.Fatal("review launched or reconnected a terminal")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if strings.Contains(h.m.View().Content, "Recover sessions") || h.m.workspace.conn != conn {
		t.Fatal("closing recovery center changed original client")
	}
}

func TestRecoveryNewSessionReviewsOriginalLaunchAndPreservesArgumentBoundaries(t *testing.T) {
	source := sampleSessions()[0]
	launch := session.Launch{Name: "original", Executable: "agent.exe", Dir: t.TempDir(), Args: []string{"", "literal $VALUE", "x y"}, Env: []string{"PRIVATE_TEST=keep"}}
	svc := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }, history: func(context.Context, int) ([]session.HistoryRecord, error) {
		return []session.HistoryRecord{{Session: source}}, nil
	}}}
	h := newWorkspaceHarness(t, svc, Options{SidebarMode: "flat", ReviewRecovery: func(context.Context, session.Session) (recovery.Review, error) {
		return recovery.Review{Source: source, Launch: &launch}, nil
	}})
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.listBusy })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: 'R', Text: "R"})
	h.until(func() bool { return h.m.recoveryUI.center != nil && !h.m.recoveryUI.center.busy })
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return h.m.recoveryUI.center.detail })
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if !strings.Contains(h.m.View().Content, "New terminal") || len(svc.starts) != 0 {
		t.Fatal("new Session did not open an explicit launch review")
	}
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.until(func() bool { return h.m.launch == nil && len(svc.starts) == 1 })
	svc.mu.Lock()
	got := svc.starts[0]
	svc.mu.Unlock()
	if got.Executable != launch.Executable || got.Dir != launch.Dir || !reflect.DeepEqual(got.Args, launch.Args) || !reflect.DeepEqual(got.Env, launch.Env) {
		t.Fatal("saved launch boundaries changed")
	}
}

func TestRecoveryRestoreRequiresEndedConsentAndExplicitStartWithOneAttempt(t *testing.T) {
	source := session.Session{ID: strings.Repeat("1", 32), RunID: strings.Repeat("2", 32), BackendID: "asd_" + strings.Repeat("1", 32) + "_" + strings.Repeat("2", 32), Title: "Ended Agent"}
	profile := recovery.Profile{Revision: 1, Declaration: recovery.Declaration{Protocol: "asd.agent-recovery", SchemaVersion: 1, SessionID: source.ID, RunID: source.RunID, AgentID: "synthetic", ConversationID: "conversation-42", Resume: recovery.Resume{Executable: `C:\Tools\agent.exe`, Args: []string{"resume", "conversation-42"}, Cwd: `C:\Work`}}}
	calls := make(chan struct{}, 2)
	gate := make(chan struct{})
	svc := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }, history: func(context.Context, int) ([]session.HistoryRecord, error) {
		return []session.HistoryRecord{{Session: source}}, nil
	}}}
	h := newWorkspaceHarness(t, svc, Options{SidebarMode: "flat", ReviewRecovery: func(context.Context, session.Session) (recovery.Review, error) {
		return recovery.Review{Source: source, Profile: &profile}, nil
	}, RestoreRecovery: func(ctx context.Context, r recovery.Review) (recovery.Outcome, error) {
		calls <- struct{}{}
		select {
		case <-gate:
		case <-ctx.Done():
			return recovery.Outcome{}, ctx.Err()
		}
		run := strings.Repeat("3", 32)
		return recovery.Outcome{Attempt: recovery.Attempt{ID: strings.Repeat("4", 32), SessionID: source.ID, RunID: run, ConversationID: profile.Declaration.ConversationID, Revision: 1, State: "started"}, Session: session.Session{ID: source.ID, RunID: run, BackendID: "asd_" + source.ID + "_" + run, Title: source.Title, TerminalState: session.TerminalAvailable}}, nil
	}})
	h.until(func() bool { return !h.m.listBusy })
	h.send(tea.KeyPressMsg{Code: 'R', Text: "R"})
	h.until(func() bool { return h.m.recoveryUI.center != nil && !h.m.recoveryUI.center.busy })
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return h.m.recoveryUI.center.detail })
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if !strings.Contains(h.m.View().Content, "I confirm the previous Run ended") {
		t.Fatal("no ended-consent form")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter}) // Default Cancel.
	if len(calls) != 0 {
		t.Fatal("default Enter launched recovery")
	}
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if len(calls) != 0 {
		t.Fatal("unchecked ended consent launched recovery")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeySpace})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	select {
	case <-calls:
	case <-time.After(time.Second):
		t.Fatal("explicit confirmed recovery not submitted")
	}
	if len(calls) != 0 {
		t.Fatal("duplicate attempt while starting")
	}
	close(gate)
	h.until(func() bool { return h.m.recoveryUI.center == nil })
	if !strings.Contains(h.m.status, "not yet confirmed") {
		t.Fatal("started process was confused with restored conversation")
	}
}

func TestRecoveryCannotStartWhenConfirmationControlsAreClipped(t *testing.T) {
	source := session.Session{ID: strings.Repeat("1", 32), RunID: strings.Repeat("2", 32), BackendID: "asd_" + strings.Repeat("1", 32) + "_" + strings.Repeat("2", 32), Title: "Ended"}
	profile := recovery.Profile{Revision: 1, Declaration: recovery.Declaration{Protocol: "asd.agent-recovery", SchemaVersion: 1, SessionID: source.ID, RunID: source.RunID, AgentID: "synthetic", ConversationID: "conversation-42", Resume: recovery.Resume{Executable: `C:\Tools\agent.exe`, Args: []string{}, Cwd: `C:\Work`}}}
	calls := 0
	svc := &workspaceFixture{fakeService: fakeService{history: func(context.Context, int) ([]session.HistoryRecord, error) {
		return []session.HistoryRecord{{Session: source}}, nil
	}}}
	h := newWorkspaceHarness(t, svc, Options{SidebarMode: "flat", ReviewRecovery: func(context.Context, session.Session) (recovery.Review, error) {
		return recovery.Review{Source: source, Profile: &profile}, nil
	}, RestoreRecovery: func(context.Context, recovery.Review) (recovery.Outcome, error) {
		calls++
		return recovery.Outcome{}, context.DeadlineExceeded
	}})
	h.until(func() bool { return !h.m.listBusy })
	h.send(tea.KeyPressMsg{Code: 'R', Text: "R"})
	h.until(func() bool { return h.m.recoveryUI.center != nil && !h.m.recoveryUI.center.busy })
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return h.m.recoveryUI.center.detail })
	h.m.listErr = nil
	h.m.diagnostic.discoveryObserved = true
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if !h.m.recoveryUI.center.confirm {
		t.Fatal("confirmation not opened")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeySpace})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.WindowSizeMsg{Width: 12, Height: 4})
	_, cmd := h.m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		h.send(cmd())
	}
	if calls != 0 {
		t.Fatal("clipped confirmation controls authorized execution")
	}
}
