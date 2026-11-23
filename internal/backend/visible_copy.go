package backend

import (
	"context"
	"fmt"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

// CaptureVisible is a fresh, bounded current-pane snapshot. It does not include
// history, resize, write input or use the UI/preview cache. The pre/post checks
// discard known target changes; they are not an atomic backend snapshot.
func (b *Backend) CaptureVisible(ctx context.Context, s session.Session) (session.Preview, error) {
	original, err := b.validateTarget(ctx, s)
	if err != nil {
		return session.Preview{}, err
	}
	if original.Command != s.Command {
		return session.Preview{}, fmt.Errorf("original executable identity changed; visible capture refused")
	}
	if b.paneState(ctx, s.BackendID, original.PaneID) == session.TerminalUnknown {
		return session.Preview{}, fmt.Errorf("original pane could not be verified before visible capture")
	}
	out, err := b.run(ctx, terminal.MaxFrameBytes, b.paneArgs(s.BackendID, original.PaneID, "capture-pane", "-p", "-e")...)
	if err != nil {
		return session.Preview{}, err
	}
	if out.truncated || !terminal.ValidCaptureEncoding(out.text) {
		return session.Preview{}, fmt.Errorf("visible capture exceeded bounds or has invalid UTF-8; not copied")
	}
	after, err := b.validateTarget(ctx, s)
	if err != nil {
		return session.Preview{}, err
	}
	if after != original || b.paneState(ctx, s.BackendID, original.PaneID) == session.TerminalUnknown {
		return session.Preview{}, fmt.Errorf("original terminal changed during visible capture; discarded")
	}
	return session.Preview{Text: out.text}, nil
}
