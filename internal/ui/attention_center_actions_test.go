package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type centerFixture struct {
	*attentionFixture
	snoozes int
	err     error
	gate    <-chan struct{}
}

func (s *centerFixture) CreateGroup(context.Context, string, string, string) (session.Group, error) {
	return session.Group{}, fmt.Errorf("unused")
}
func (s *centerFixture) MoveToGroup(context.Context, session.Session, string) (session.Session, error) {
	return session.Session{}, fmt.Errorf("unused")
}

func (s *centerFixture) Workspace(context.Context) (session.Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return session.Workspace{Sessions: sampleSessions(), Attention: []session.AttentionItem{s.item}}, nil
}
func (s *centerFixture) SnoozeAttention(ctx context.Context, expected session.AttentionItem, d time.Duration) (session.AttentionItem, error) {
	s.mu.Lock()
	s.snoozes++
	gate, err := s.gate, s.err
	s.mu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return session.AttentionItem{}, ctx.Err()
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		return session.AttentionItem{}, err
	}
	if expected.ID != s.item.ID || expected.ReminderRevision != s.item.ReminderRevision {
		return session.AttentionItem{}, fmt.Errorf("stale")
	}
	s.item.SnoozedUntil = nil
	if d != 0 {
		until := time.Now().Add(d)
		s.item.SnoozedUntil = &until
	}
	s.item.ReminderRevision++
	return s.item, nil
}
func newCenterHarness(t *testing.T) (*workspaceHarness, *centerFixture) {
	base := &attentionFixture{}
	base.item = session.AttentionItem{AttentionEvent: session.AttentionEvent{ID: strings.Repeat("a", 32), Version: 1, Kind: "approval", State: "pending", Title: "Allow operation?"}, Source: sampleSessions()[1], TerminalAvailable: true, CreatedAt: time.Now()}
	s := &centerFixture{attentionFixture: base}
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil && len(h.m.attention.items) == 1 })
	h.send(tea.KeyPressMsg{Code: 'a', Mod: tea.ModAlt})
	h.until(func() bool { return h.m.attention.center != nil })
	return h, s
}
func TestAttentionCenterSnoozeFilterResumeAndExplicitOpen(t *testing.T) {
	h, s := newCenterHarness(t)
	h.send(tea.KeyPressMsg{Code: 's', Text: "s"})
	h.until(func() bool { return !h.m.attention.center.busy && h.m.attention.items[0].SnoozedUntil != nil })
	if s.visits != 0 || h.m.attention.items[0].ViewedAt != nil {
		t.Fatal("snooze viewed/answered prompt")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	if !strings.Contains(ansi.Strip(h.m.View().Content), "0 items") {
		t.Fatal("Due filter includes snoozed request")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	if !strings.Contains(ansi.Strip(h.m.View().Content), "1 items") {
		t.Fatal("Snoozed filter lost pending request")
	}
	h.send(tea.KeyPressMsg{Code: 'u', Text: "u"})
	h.until(func() bool { return !h.m.attention.center.busy && h.m.attention.items[0].SnoozedUntil == nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		return h.m.attention.center == nil && h.m.workspace.active == 1 && h.m.workspace.conn != nil && !h.m.attention.busy
	})
	if h.m.attention.items[0].State != "pending" || s.visits != 1 {
		t.Fatal("explicit view resolved or duplicated prompt")
	}
}
func TestAttentionCenterOneWriteAndUncertainResultNeverRetries(t *testing.T) {
	h, s := newCenterHarness(t)
	gate := make(chan struct{})
	s.gate = gate
	s.err = fmt.Errorf("commit uncertain")
	h.send(tea.KeyPressMsg{Code: 's', Text: "s"})
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	h.send(tea.PasteMsg{Content: "y\n"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if h.m.attention.center == nil || !h.m.attention.center.busy {
		t.Fatal("write barrier was bypassed")
	}
	close(gate)
	h.until(func() bool { return !h.m.attention.center.busy })
	h.send(tea.KeyPressMsg{Code: 's', Text: "s"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	s.mu.Lock()
	calls, visits := s.snoozes, s.visits
	s.mu.Unlock()
	if calls != 1 || visits != 0 || !strings.Contains(ansi.Strip(h.m.View().Content), "unconfirmed") {
		t.Fatal("uncertain write retried or navigated")
	}
}
