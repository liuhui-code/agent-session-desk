package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type managementMessageProbe struct{ *groupedFixture }

func (s *managementMessageProbe) PendingMessage(_ context.Context, v session.Session) (session.Message, error) {
	return session.Message{ID: strings.Repeat("a", 32), Target: v, Body: "backend only", State: "queued"}, nil
}
func (s *managementMessageProbe) DeliverMessage(context.Context, string, session.Session, func(context.Context, string) error) (session.Message, error) {
	panic("frontend must not call message delivery")
}

func TestMessagingDoesNotAddManagementKeybindings(t *testing.T) {
	service := &managementMessageProbe{newGroupedFixture()}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: 'm', Text: "m"})
	if strings.Contains(h.m.View().Content, "Review queued message") {
		t.Fatal("lowercase m introduced a frontend messaging operation")
	}
	h.send(tea.KeyPressMsg{Code: 'M', Text: "M"})
	h.until(func() bool { return h.m.groupDialog != nil })
	if !h.m.groupDialog.move {
		t.Fatal("uppercase M no longer moves the selected session")
	}
}
