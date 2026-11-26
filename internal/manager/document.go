package manager

import (
	"context"
	"fmt"

	"github.com/liuhui-code/agent-session-desk/internal/document"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Manager) ResolveMarkdown(ctx context.Context, target session.Session, ref string) (string, error) {
	live, err := m.messageTarget(ctx, target.ID, target.RunID)
	if err != nil {
		return "", err
	}
	if live.BackendID != target.BackendID || live.Command != target.Command {
		return "", fmt.Errorf("original terminal changed")
	}
	relative, err := document.NeedsDirectory(ref)
	if err != nil {
		return "", err
	}
	directory := ""
	if relative {
		source, ok := m.Terminal.(interface {
			CurrentDirectory(context.Context, session.Session) (string, error)
		})
		if !ok {
			return "", fmt.Errorf("pane directory unknown; use an absolute Markdown path")
		}
		directory, err = source.CurrentDirectory(ctx, target)
		if err != nil {
			return "", err
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return document.Resolve(ref, directory)
}
