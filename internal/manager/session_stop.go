package manager

import (
	"context"
	"fmt"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// StopReviewed binds a human confirmation to its original named Run. Discovery,
// metadata and terminal effects are separate resources, not one transaction.
func (m *Manager) StopReviewed(ctx context.Context, expected session.Session) error {
	view, err := m.Workspace(ctx)
	if err != nil {
		return fmt.Errorf("revalidate stop target: %w", err)
	}
	matches := 0
	var target session.Session
	for _, v := range view.Sessions {
		if v.ID == expected.ID {
			matches++
			target = v
		}
	}
	if matches != 1 || target.RunID != expected.RunID || target.BackendID != expected.BackendID || target.Command != expected.Command || target.Title != expected.Title || target.TitleRevision != expected.TitleRevision || (target.TerminalState != session.TerminalAvailable && target.TerminalState != session.TerminalExited) {
		return fmt.Errorf("reviewed Session/Run changed or unavailable; no stop attempted; refresh and review again")
	}
	return m.Stop(ctx, target)
}
