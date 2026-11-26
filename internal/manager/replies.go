package manager

import (
	"context"
	"fmt"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Manager) BindAgent(ctx context.Context, c session.AgentCredential) (session.Session, error) {
	if err := c.Validate(); err != nil {
		return session.Session{}, err
	}
	v, err := m.messageTarget(ctx, c.SessionID, c.RunID)
	if err != nil {
		return v, err
	}
	return v, m.metadata.BindAgent(ctx, c, v)
}

func (m *Manager) authenticateAgent(ctx context.Context, c session.AgentCredential) error {
	raw, err := m.metadata.AgentIdentity(ctx, c)
	if err != nil {
		return err
	}
	live, err := m.messageTarget(ctx, raw.ID, raw.RunID)
	if err != nil {
		return err
	}
	if live.BackendID != raw.BackendID || live.Command != raw.Command {
		return fmt.Errorf("original agent Run identity changed")
	}
	return nil
}

func (m *Manager) Inbox(ctx context.Context, c session.AgentCredential) ([]session.Message, error) {
	if err := m.authenticateAgent(ctx, c); err != nil {
		return nil, err
	}
	return m.metadata.Inbox(ctx, c)
}

func (m *Manager) Reply(ctx context.Context, c session.AgentCredential, r session.ReplyRequest) (session.Reply, error) {
	if err := r.Validate(); err != nil {
		return session.Reply{}, err
	}
	if err := m.authenticateAgent(ctx, c); err != nil {
		return session.Reply{}, err
	}
	return m.metadata.RecordReply(ctx, c, r)
}

func (m *Manager) Replies(ctx context.Context, id string) ([]session.Reply, error) {
	return m.metadata.Replies(ctx, id)
}
