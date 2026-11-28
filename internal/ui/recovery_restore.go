package ui

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) recoveryBlocked(d *recoveryCenter) string {
	if d.submitted {
		return "Outcome may be uncertain; close and reopen recovery review to inspect the latest attempt"
	}
	if d.review.Profile == nil {
		return "No Agent recovery declaration; no conversation restore can be guessed"
	}
	if m.recoveryUI.restore == nil {
		return "Recovery execution unavailable in this host"
	}
	if m.listErr != nil || !m.diagnostic.discoveryObserved {
		return "Backend state unknown; refresh observations before recovery"
	}
	p := d.review.Profile
	if p.Revision < 1 || p.Declaration.SessionID != d.review.Source.ID || p.Declaration.Validate() != nil {
		return "Recovery declaration is invalid; no process will start"
	}
	if !session.ValidID(d.review.Source.ID) || !session.ValidID(d.review.Source.RunID) {
		return "Recovery requires complete original identity"
	}
	if d.review.Attempt != nil && d.review.Attempt.State == "launch_unknown" {
		return "Previous recovery launch is unknown; inspect it before any new attempt"
	}
	for _, s := range m.sessions {
		if s.ID == d.review.Source.ID {
			return "Original terminal still exists; reattach or explicitly stop it before restoring"
		}
	}
	w := m.workspace
	if w.busy || w.transition != workspaceIdle {
		return "Native view is busy; wait before restoring"
	}
	if w.conn != nil && len(w.tabs) > 0 && w.tabs[w.active].session.ID == d.review.Source.ID {
		return "Original native client is still connected; no recovery will start"
	}
	return ""
}
func (m *Model) beginRestoreReview(d *recoveryCenter) tea.Cmd {
	if reason := m.recoveryBlocked(d); reason != "" {
		d.message = reason
		return nil
	}
	d.confirm = true
	d.ended = false
	d.startSelected = false
	d.offset = 0
	d.message = "Starting a process does not prove that the Agent restored its conversation"
	return nil
}
func (m *Model) submitRecovery(d *recoveryCenter) tea.Cmd {
	if m.diagnosticContentWidth() < m.tokens.diagnostic.confirmWidth || m.height < m.tokens.diagnostic.confirmHeight {
		d.ended = false
		d.message = "Enlarge the window and confirm again before starting recovery"
		return nil
	}
	if !d.ended {
		d.message = "Confirm that the previous Run has ended first"
		return nil
	}
	if reason := m.recoveryBlocked(d); reason != "" {
		d.message = reason
		return nil
	}
	d.busy = true
	d.submitted = true
	d.message = "Starting one recovery attempt; no automatic retry"
	review, restore, ctx := d.review, m.recoveryUI.restore, m.ctx
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 65*time.Second)
		defer cancel()
		outcome, err := restore(ctx, review)
		return recoveryLaunched{d, outcome, err}
	}
}
func (m *Model) recoveryLaunchResult(result recoveryLaunched) tea.Cmd {
	d := m.recoveryUI.center
	if d != result.center {
		return nil
	}
	d.busy = false
	d.confirm = false
	d.ended = false
	a, target := result.outcome.Attempt, result.outcome.Session
	if a.ID != "" {
		d.review.Attempt = &a
	}
	if result.err != nil {
		if !result.outcome.Attempted && a.ID == "" {
			d.submitted = false
			d.message = "Recovery not started; close and refresh review before retry"
			return nil
		}
		d.message = "Recovery outcome unconfirmed; close and reopen review to inspect status. No automatic retry"
		return m.loadSessions()
	}
	p := d.review.Profile
	if p == nil || !session.ValidID(a.ID) || a.SessionID != d.review.Source.ID || a.RunID == d.review.Source.RunID || a.ConversationID != p.Declaration.ConversationID || a.Revision != p.Revision || target.ID != a.SessionID || target.RunID != a.RunID || !session.ValidID(target.RunID) || target.BackendID != "asd_"+target.ID+"_"+target.RunID {
		d.message = "Recovery identity or receipt unconfirmed; no automatic attach or retry"
		return m.loadSessions()
	}
	if target.TerminalState != session.TerminalAvailable {
		d.message = "Recovery launch recorded; terminal not interactive. Inspect the Agent receipt"
		return m.loadSessions()
	}
	m.recoveryUI.center = nil
	cmd := m.acceptLaunchedSession(target)
	m.status = "Recovery process started; conversation restore not yet confirmed"
	switch a.State {
	case "restored":
		m.status = "Agent reported the original conversation restored"
	case "failed", "not_found":
		m.status = fmt.Sprintf("Agent reported %s; inspect its terminal", a.State)
	}
	return tea.Batch(cmd, m.loadSessions())
}
