package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type sessionStopFixture struct {
	*workspaceFixture
	stopped   []session.Session
	stopErr   error
	stopGate  <-chan struct{}
	completed []session.Session
}

func (s *sessionStopFixture) List(ctx context.Context) ([]session.Session, error) {
	rows, err := s.workspaceFixture.List(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []session.Session{}
	for _, v := range rows {
		stopped := false
		for _, old := range s.completed {
			stopped = stopped || sameStopTarget(v, old)
		}
		if !stopped {
			result = append(result, v)
		}
	}
	return result, err
}

func (s *sessionStopFixture) StopReviewed(ctx context.Context, target session.Session) error {
	s.mu.Lock()
	s.stopped = append(s.stopped, target)
	gate, err := s.stopGate, s.stopErr
	s.mu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err == nil {
		s.mu.Lock()
		s.completed = append(s.completed, target)
		s.mu.Unlock()
	}
	return err
}
func (s *sessionStopFixture) count() int { s.mu.Lock(); defer s.mu.Unlock(); return len(s.stopped) }

func TestSessionStopDefaultsToCancelAndStopsSelectedNotActiveRunOnce(t *testing.T) {
	s := &sessionStopFixture{workspaceFixture: &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}}
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "停止会话") })
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return h.m.workspace.conn != nil })
	if s.count() != 0 || h.m.workspace.tabs[0].session.ID != "first" {
		t.Fatal("default Enter stopped original terminal")
	}
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.until(func() bool { return s.count() == 1 && !strings.Contains(h.m.View().Content, "停止会话") })
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped[0].ID != "second" || s.stopped[0].RunID != sampleSessions()[1].RunID || len(h.m.workspace.tabs) != 2 || h.m.workspace.active != 0 {
		t.Fatal("stop changed original identity or unrelated Tab")
	}
}
