package backend

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// CurrentDirectory reads the original pane's reported directory on demand.
// It does not use the manager's cwd, launch a shell, or modify terminal input.
func (b *Backend) CurrentDirectory(ctx context.Context, target session.Session) (string, error) {
	before, err := b.validateTarget(ctx, target)
	if err != nil {
		return "", err
	}
	if before.Command != target.Command || b.paneState(ctx, target.BackendID, before.PaneID) != session.TerminalAvailable {
		return "", fmt.Errorf("original pane unavailable")
	}
	out, err := b.run(ctx, 4096, b.paneArgs(target.BackendID, before.PaneID, "display-message", "-p", "#{pane_current_path}")...)
	if err != nil {
		return "", err
	}
	directory := strings.TrimSuffix(strings.TrimSuffix(out.text, "\n"), "\r")
	if out.truncated || !utf8.ValidString(directory) || strings.ContainsAny(directory, "\x00\r\n") || !filepath.IsAbs(directory) {
		return "", fmt.Errorf("pane directory unknown; use an absolute Markdown path")
	}
	after, err := b.validateTarget(ctx, target)
	if err != nil || after != before || b.paneState(ctx, target.BackendID, before.PaneID) != session.TerminalAvailable {
		return "", fmt.Errorf("original pane changed during directory lookup")
	}
	return directory, nil
}
