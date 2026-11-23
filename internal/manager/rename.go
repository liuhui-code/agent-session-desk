package manager

import (
	"context"
	"fmt"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Manager) Rename(ctx context.Context, expected session.Session, value string) (session.Session, error) {
	title, err := session.NormalizeRenameTitle(value)
	if err != nil {
		return session.Session{}, err
	}
	items, err := m.List(ctx)
	if err != nil {
		return session.Session{}, err
	}
	for _, item := range items {
		if item.ID == expected.ID && item.RunID == expected.RunID && item.BackendID == expected.BackendID && item.Command == expected.Command {
			return m.metadata.Rename(ctx, expected, title)
		}
	}
	return session.Session{}, fmt.Errorf("original Session/Run is no longer discovered; no rename applied")
}
