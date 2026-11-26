package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"time"
)

func (m *Model) loadPreview() tea.Cmd {
	if m.workspace != nil {
		return nil
	}
	s, ok := m.selected()
	if !ok || m.previewBusy || m.attaching || m.closed || m.listErr != nil || !m.previewVisible() {
		return nil
	}
	if s.TerminalState != session.TerminalAvailable && s.TerminalState != session.TerminalExited {
		m.invalidatePreview()
		return nil
	}
	m.previewBusy = true
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	m.previewCancel = cancel
	generation := m.previewGeneration
	lines := min(maxPreviewLines, max(1, m.height))
	return func() tea.Msg {
		defer cancel()
		preview, err := m.service.Preview(ctx, s, lines)
		return previewMsg{generation: generation, sessionID: s.ID, runID: s.RunID, preview: preview, err: err}
	}
}
