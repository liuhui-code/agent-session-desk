package backend

import (
	"context"
	"strings"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// Query identity and state together so a target fallback cannot turn a different
// pane's output into evidence for this Run. This is still not an atomic process
// identity contract across arbitrary external respawn/server replacement.
func (b *Backend) observePane(ctx context.Context, name, pane string) (session.TerminalState, *session.PaneExit) {
	out, err := b.run(ctx, maxControlBytes, b.paneArgs(name, pane, "display-message", "-p", "#{pane_id}|#{pane_dead}")...)
	if err != nil || out.truncated {
		return session.TerminalUnknown, nil
	}
	switch strings.TrimSpace(out.text) {
	case pane + "|0":
		return session.TerminalAvailable, nil
	case pane + "|1":
		r, err := b.readRecord(ctx, name)
		if err != nil || r.PaneID != pane {
			return session.TerminalUnknown, nil
		}
		return session.TerminalExited, &session.PaneExit{PaneID: pane, Source: b.kind + ".pane_dead", ObservedAt: time.Now().UTC()}
	default:
		return session.TerminalUnknown, nil
	}
}
