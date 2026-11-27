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

	if l.Executable != "" {
		var err error
		l, err = session.ResolveLaunch(l)
		if err != nil {
			return session.Session{}, err
		}
	}
	if l.GroupID != "" {
		groups, err := m.metadata.Groups(ctx)
		if err != nil {
			return session.Session{}, err
		}
		found := false
		for _, g := range groups {
			if g.ID == l.GroupID {
				found = true
			}
		}
		if !found {
			return session.Session{}, fmt.Errorf("target group not found; terminal not started")
		}
	}
	if err := m.metadata.CheckWritable(ctx); err != nil {
		return session.Session{}, fmt.Errorf("metadata not writable; terminal not started: %w", err)
	}
	v, err := m.Terminal.Start(ctx, l)
	if err != nil {
		return v, err
	}
	if l.GroupID == "" {
		err = m.metadata.Save(ctx, v)
	} else {
		err = m.metadata.SaveInGroup(ctx, v, l.GroupID)
	}
	if err != nil {
		return v, fmt.Errorf("terminal created (session_id=%s run_id=%s); metadata commit failed; do not retry start: %w", v.ID, v.RunID, err)
	}
	if l.Executable != "" {
		if err = m.metadata.SaveLaunch(ctx, v, l); err != nil {
			return v, fmt.Errorf("terminal created (session_id=%s run_id=%s); launch description commit failed; do not retry start: %w", v.ID, v.RunID, err)
		}
	}
	if l.GroupID != "" {
		view, e := m.metadata.Workspace(ctx, []session.Session{v})
		if e != nil {
			return v, fmt.Errorf("terminal created (session_id=%s run_id=%s); metadata read failed; do not retry start: %w", v.ID, v.RunID, e)
		}
		v = view.Sessions[0]
	}
	return v, nil
}

func (m *Manager) List(ctx context.Context) ([]session.Session, error) {
	view, err := m.Workspace(ctx)
	return view.Sessions, err
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
