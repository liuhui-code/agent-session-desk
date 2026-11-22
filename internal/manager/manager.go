// Package manager coordinates terminal effects and metadata commits. They are
// separate resources: no SQL transaction can make process creation atomic.
package manager

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type Terminal interface {
	List(context.Context) ([]session.Session, error)
	Start(context.Context, session.Launch) (session.Session, error)
	Stop(context.Context, session.Session) error
	Preview(context.Context, session.Session, int) (session.Preview, error)
	Attach(context.Context, session.Session) (*exec.Cmd, error)
}

type Manager struct {
	Terminal
	metadata *store.Store
}

func New(terminal Terminal, metadata *store.Store) *Manager {
	return &Manager{Terminal: terminal, metadata: metadata}
}
func (m *Manager) Start(ctx context.Context, l session.Launch) (session.Session, error) {
	if err := m.metadata.CheckWritable(ctx); err != nil {
		return session.Session{}, fmt.Errorf("metadata not writable; terminal not started: %w", err)
	}
	v, err := m.Terminal.Start(ctx, l)
	if err != nil {
		return v, err
	}
	if err := m.metadata.Save(ctx, v); err != nil {
		return v, fmt.Errorf("terminal created (session_id=%s run_id=%s); metadata commit failed; do not retry start: %w", v.ID, v.RunID, err)
	}
	return v, nil
}

func (m *Manager) List(ctx context.Context) ([]session.Session, error) {
	rows, err := m.Terminal.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range rows {
		if err = m.metadata.Save(ctx, v); err != nil {
			return nil, fmt.Errorf("record backend discovery: %w", err)
		}
	}
	return rows, nil
}

func (m *Manager) Stop(ctx context.Context, v session.Session) error {
	if err := m.Terminal.Stop(ctx, v); err != nil {
		return err
	}
	if err := m.metadata.MarkStopped(ctx, v); err != nil {
		return fmt.Errorf("backend session stopped (session_id=%s run_id=%s); metadata update failed: %w", v.ID, v.RunID, err)
	}
	return nil
}

func (m *Manager) History(ctx context.Context, limit int) ([]session.HistoryRecord, error) {
	return m.metadata.History(ctx, limit)
}
