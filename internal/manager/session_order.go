package manager

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func (m *Manager) ReorderSession(ctx context.Context, expected session.SessionOrder, delta int) (session.Workspace, error) {
	view, err := m.Workspace(ctx)
	if err != nil {
		return session.Workspace{}, err
	}
	if len(expected.Siblings) == 0 || view.OrderStamps[session.GroupOf(expected.Siblings[0])] != expected.Stamp || !session.SameOrderSnapshot(session.OrderPeers(view.Sessions, expected.TargetID), expected.Siblings) {
		return session.Workspace{}, store.ErrSessionOrderConflict
	}
	return m.metadata.ReorderSession(ctx, expected, delta, view.Sessions)
}
