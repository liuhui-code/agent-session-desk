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
	state, exit, _ := b.observePaneProgram(ctx, name, pane, false)
	return state, exit
}
func (b *Backend) observePaneProgram(ctx context.Context, name, pane string, program bool) (session.TerminalState, *session.PaneExit, string) {
	format := "#{pane_id}|#{pane_dead}"
	if program {
		format += "|#{pane_current_command}"
	}
	out, err := b.run(ctx, maxControlBytes, b.paneArgs(name, pane, "display-message", "-p", format)...)
	if err != nil || out.truncated {
		if program {
			state, exit := b.observePane(ctx, name, pane)
			return state, exit, ""
		}
		return session.TerminalUnknown, nil, ""
	}
	parts := strings.SplitN(strings.TrimSpace(out.text), "|", 3)
	if len(parts) < 2 || parts[0] != pane {
		return session.TerminalUnknown, nil, ""
	}
	current := ""
	if len(parts) == 3 {
		current = session.ProgramLabel(parts[2])
	}
	switch parts[0] + "|" + parts[1] {
	case pane + "|0":
		return session.TerminalAvailable, nil, current
	case pane + "|1":
		r, err := b.readRecord(ctx, name)
		if err != nil || r.PaneID != pane {
			return session.TerminalUnknown, nil, ""
		}
		return session.TerminalExited, &session.PaneExit{PaneID: pane, Source: b.kind + ".pane_dead", ObservedAt: time.Now().UTC()}, ""
	default:
		return session.TerminalUnknown, nil, ""
	}
}
