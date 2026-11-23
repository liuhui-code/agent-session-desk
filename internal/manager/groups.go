package manager

import (
	"context"
	"fmt"
	"sort"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// Workspace records immutable discovery before projecting mutable names and
// membership together. No terminal callback runs inside a metadata transaction.
func (m *Manager) Workspace(ctx context.Context) (session.Workspace, error) {
	rows, err := m.Terminal.List(ctx)
	if err != nil {
		return session.Workspace{}, err
	}
	for _, v := range rows {
		if err = m.metadata.Save(ctx, v); err != nil {
			return session.Workspace{}, fmt.Errorf("record backend discovery: %w", err)
		}
	}
	view, err := m.metadata.Workspace(ctx, rows)
	if err != nil {
		return session.Workspace{}, err
	}
	sort.SliceStable(view.Sessions, func(i, j int) bool {
		a, b := view.Sessions[i], view.Sessions[j]
		if a.Title == b.Title {
			return a.ID < b.ID
		}
		return a.Title < b.Title
	})
	return view, nil
}

func (m *Manager) CreateGroup(ctx context.Context, parentID, name, dir string) (session.Group, error) {
	return m.metadata.CreateGroup(ctx, parentID, name, dir)
}

func (m *Manager) RenameGroup(ctx context.Context, expected session.Group, name string) (session.Group, error) {
	return m.metadata.RenameGroup(ctx, expected, name)
}

func (m *Manager) ReorderGroup(ctx context.Context, expected session.GroupOrder, delta int) ([]session.Group, error) {
	return m.metadata.ReorderGroup(ctx, expected, delta)
}

func (m *Manager) MoveToGroup(ctx context.Context, expected session.Session, target string) (session.Session, error) {
	view, err := m.Workspace(ctx)
	if err != nil {
		return session.Session{}, err
	}
	for _, v := range view.Sessions {
		if v.ID == expected.ID && v.RunID == expected.RunID && v.BackendID == expected.BackendID && v.Command == expected.Command {
			// Name edits are independent; retain their current value while the
			// original membership revision still governs this move's CAS.
			v.GroupID, v.GroupRevision = expected.GroupID, expected.GroupRevision
			return m.metadata.MoveToGroup(ctx, v, target)
		}
	}
	return session.Session{}, fmt.Errorf("original Session/Run is no longer discovered; no move applied")
}
