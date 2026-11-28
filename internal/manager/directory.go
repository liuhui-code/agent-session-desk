package manager

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// CurrentDirectory is an on-demand, identity-checked pane observation. It never
// substitutes the manager's own cwd or a previous Run's launch directory.
func (m *Manager) CurrentDirectory(ctx context.Context, target session.Session) (string, error) {
	live, err := m.messageTarget(ctx, target.ID, target.RunID)
	if err != nil {
		return "", err
	}
	if live.BackendID != target.BackendID || live.Command != target.Command {
		return "", fmt.Errorf("original terminal changed")
	}
	source, ok := m.Terminal.(interface {
		CurrentDirectory(context.Context, session.Session) (string, error)
	})
	if !ok {
		return "", fmt.Errorf("pane directory unavailable")
	}
	dir, err := source.CurrentDirectory(ctx, target)
	if err != nil {
		return "", err
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("pane directory unavailable")
	}
	return dir, nil
}
