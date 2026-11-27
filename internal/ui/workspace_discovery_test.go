package ui

import (
	"context"
	"errors"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"testing"
)

func TestDiscoveryAbsenceDoesNotCloseHealthyNativeClient(t *testing.T) {
	svc := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, svc)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.listBusy })
	w := h.m.workspace
	conn, generation := w.conn, w.generation
	original := w.tabs[w.active].session
	h.m.sessions = nil
	h.m.workspaceSessions()
	if w.conn != conn || w.generation != generation || w.err != nil {
		t.Fatalf("enumeration absence disconnected native client: %v", w.err)
	}
	h.m.sessions = []session.Session{original}
	h.m.workspaceSessions()
	if w.conn != conn || w.err != nil || !w.tabs[w.active].live {
		t.Fatal("original did not reconcile without reconnect")
	}
	original.TerminalState = session.TerminalExited
	h.m.sessions = []session.Session{original}
	h.m.workspaceSessions()
	if w.conn != nil || w.err == nil {
		t.Fatal("explicit original exit did not disconnect")
	}
}

func TestDiscoveryAbsenceStillAcceptsNativeStreamFailure(t *testing.T) {
	svc := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, svc)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.listBusy })
	w := h.m.workspace
	h.m.sessions = nil
	h.m.workspaceSessions()
	h.m.Update(workspaceFrame{generation: w.generation, err: errors.New("original stream ended")})
	if w.conn != nil || w.err == nil {
		t.Fatal("actual stream failure was hidden by discovery uncertainty")
	}
}
