package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/diagnostics"
	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type recoveryUIState struct {
	center  *recoveryCenter
	review  func(context.Context, session.Session) (recovery.Review, error)
	restore func(context.Context, recovery.Review) (recovery.Outcome, error)
}
type recoveryCenter struct {
	rows                                                   []session.Session
	width, height                                          int
	cursor, offset                                         int
	busy, detail, confirm, ended, startSelected, submitted bool
	review                                                 recovery.Review
	message                                                string
	cancel                                                 context.CancelFunc
}
type recoveryHistoryLoaded struct {
	center *recoveryCenter
	rows   []session.HistoryRecord
	err    error
}
type recoveryReviewed struct {
	center *recoveryCenter
	source session.Session
	review recovery.Review
	err    error
}
type recoveryLaunched struct {
	center  *recoveryCenter
	outcome recovery.Outcome
	err     error
}

func (m *Model) beginRecoveryCenter() tea.Cmd {
	if m.workspace == nil || (m.workspace.focused && len(m.workspace.tabs) > 0) || m.workspace.transition != workspaceIdle {
		return nil
	}
	if m.recoveryUI.review == nil {
		m.status = "Recovery review unavailable in this host"
		return nil
	}
	g := m.workspaceGeometry()
	d := &recoveryCenter{width: g.width, height: g.height}
	m.recoveryUI.center = d
	return m.loadRecoveryHistory(d)
}
func (m *Model) loadRecoveryHistory(d *recoveryCenter) tea.Cmd {
	if d.busy || d.submitted {
		return nil
	}
	d.busy = true
	d.detail = false
	d.confirm = false
	d.message = "Loading recent session metadata…"
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	d.cancel = cancel
	service := m.service
	return func() tea.Msg {
		rows, err := service.History(ctx, historyLimit)
		return recoveryHistoryLoaded{d, rows, err}
	}
}
func (m *Model) inspectRecovery(d *recoveryCenter) tea.Cmd {
	if len(d.rows) == 0 || d.busy {
		return nil
	}
	source := d.rows[d.cursor]
	d.busy = true
	d.message = "Loading recovery review…"
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	d.cancel = cancel
	review := m.recoveryUI.review
	return func() tea.Msg { r, err := review(ctx, source); return recoveryReviewed{d, source, r, err} }
}
func (m *Model) closeRecoveryCenter() tea.Cmd {
	d := m.recoveryUI.center
	if d != nil && d.cancel != nil {
		d.cancel()
	}
	m.recoveryUI.center = nil
	// Resize coordination is explicit; inspecting metadata never reconnects.
	if m.workspace != nil && m.workspace.conn != nil {
		g := m.workspaceGeometry()
		if d != nil && g.width == d.width && g.height == d.height {
			return nil
		}
		return m.workspaceSize()
	}
	return nil
}
func (m *Model) recoveryEvent(msg tea.Msg) (tea.Cmd, bool) {
	d := m.recoveryUI.center
	switch result := msg.(type) {
	case recoveryLaunched:
		return m.recoveryLaunchResult(result), true
	case recoveryHistoryLoaded:
		if d != result.center {
			return nil, true
		}
		d.busy = false
		if d.cancel != nil {
			d.cancel()
			d.cancel = nil
		}
		if result.err != nil {
			d.message = "Metadata unavailable · " + diagnostics.Cause(result.err)
			return nil, true
		}
		d.rows = nil
		seen := map[string]bool{}
		for _, row := range result.rows[:min(len(result.rows), historyLimit)] {
			if !seen[row.Session.ID] {
				seen[row.Session.ID] = true
				d.rows = append(d.rows, row.Session)
			}
		}
		d.cursor = min(d.cursor, max(0, len(d.rows)-1))
		d.message = "Enter inspects metadata; no terminal action is automatic"
		return nil, true
	case recoveryReviewed:
		if d != result.center {
			return nil, true
		}
		d.busy = false
		if d.cancel != nil {
			d.cancel()
			d.cancel = nil
		}
		if result.err != nil {
			d.message = "Recovery review unavailable · " + diagnostics.Cause(result.err)
			return nil, true
		}
		if activityKey(result.source) != activityKey(result.review.Source) {
			d.message = "Latest Run changed; refresh the list before reviewing"
			return nil, true
		}
		d.review = result.review
		d.detail = true
		d.confirm = false
		d.offset = 0
		d.message = "Review only; choose an explicit action"
		return nil, true
	}
	if d == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		return nil, true
	case tea.PasteMsg, tea.KeyReleaseMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
		return nil, true
	case AttentionActivate, GlobalActivate:
		if m.attentionFocus != nil {
			m.attentionFocus()
		}
		d.message = "Close recovery review before navigating to a pending item"
		return nil, true
	case tea.KeyPressMsg:
		if d.busy {
			if msg.String() == "esc" && !d.submitted {
				return m.closeRecoveryCenter(), true
			}
			return nil, true
		}
		switch msg.String() {
		case "esc":
			if d.confirm {
				d.confirm = false
				d.ended = false
				d.offset = 0
				return nil, true
			}
			if d.detail {
				d.detail = false
				d.offset = 0
				d.review = recovery.Review{}
				return nil, true
			}
			return m.closeRecoveryCenter(), true
		case "up", "k":
			if d.detail {
				d.offset = max(0, d.offset-1)
			} else {
				d.cursor = max(0, d.cursor-1)
			}
		case "down", "j":
			if d.detail {
				d.offset++
			} else {
				d.cursor = min(max(0, len(d.rows)-1), d.cursor+1)
			}
		case "pgup":
			d.offset = max(0, d.offset-m.diagnosticPageSize())
		case "pgdown":
			d.offset += m.diagnosticPageSize()
		case "home":
			d.offset = 0
			d.cursor = 0
		case "end":
			if d.detail {
				d.offset = len(m.recoveryDetailLines(d))
			} else {
				d.cursor = max(0, len(d.rows)-1)
			}
		case "n":
			if d.detail && !d.confirm {
				return m.newSessionFromRecovery(d), true
			}
		case "a":
			if d.detail && !d.confirm {
				return m.reattachRecoverySource(d), true
			}
		case " ", "space":
			if d.confirm {
				d.ended = !d.ended
			}
		case "tab", "left", "right":
			if d.confirm {
				d.startSelected = !d.startSelected
			}
		case "enter":
			if d.confirm {
				if !d.startSelected {
					d.confirm = false
					d.ended = false
					d.offset = 0
					return nil, true
				}
				return m.submitRecovery(d), true
			}
			if !d.detail {
				return m.inspectRecovery(d), true
			}
		case "r":
			if d.detail {
				return m.beginRestoreReview(d), true
			}
			if !d.detail {
				return m.loadRecoveryHistory(d), true
			}
		}
		return nil, true
	}
	return nil, false
}
