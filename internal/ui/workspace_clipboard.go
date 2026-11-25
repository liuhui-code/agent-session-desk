package ui

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type workspaceClipboard struct {
	generation uint64
	text       string
	err        error
}
type workspaceClipboardSent struct{ token uint64 }
type hostClipboardRequest struct {
	token, generation uint64
	notice            string
}
type hostClipboardState struct {
	serial   uint64
	inflight *hostClipboardRequest
}

func (m *Model) readWorkspaceClipboard() tea.Cmd {
	w := m.workspace
	if m.clipboard.inflight != nil || w.clipboardReading || w.visibleCopy != nil {
		return nil
	}
	source, ok := w.conn.(terminal.ClipboardSource)
	if !ok {
		return nil
	}
	generation, ctx := w.generation, w.ctx
	w.clipboardReading = true
	return func() tea.Msg { text, e := source.NextClipboard(ctx); return workspaceClipboard{generation, text, e} }
}
func (m *Model) submitHostClipboard(text string, generation uint64, notice string) tea.Cmd {
	if m.clipboard.serial == math.MaxUint64 {
		m.status = "Copy request IDs exhausted; reopen the UI"
		return nil
	}
	m.clipboard.serial++
	token := m.clipboard.serial
	m.clipboard.inflight = &hostClipboardRequest{token, generation, notice}
	m.status = "Sending copy request…"
	return tea.Sequence(tea.SetClipboard(text), func() tea.Msg { return workspaceClipboardSent{token} })
}
func (m *Model) drainWorkspaceClipboard() tea.Cmd {
	w := m.workspace
	if m.clipboard.inflight != nil {
		return nil
	}
	if op := w.visibleCopy; op != nil && op.ready {
		w.visibleCopy = nil
		if op.generation == w.generation && m.visibleCopyTargetCurrent(op.target) {
			notice := fmt.Sprintf("Copy request sent (%d visible rows, %s)", strings.Count(op.payload, "\n")+1, label(op.target.Title))
			return m.submitHostClipboard(op.payload, op.generation, notice)
		}
		m.status = "Session/Run changed; the old screen was not sent"
	}
	if pending := w.pendingClipboard; pending != nil {
		w.pendingClipboard = nil
		if pending.generation == w.generation && w.conn != nil {
			return m.submitHostClipboard(pending.text, pending.generation, "Copy request sent to the terminal")
		}
	}
	return m.readWorkspaceClipboard()
}
func (m *Model) workspaceClipboardEvent(msg tea.Msg) (tea.Cmd, bool) {
	if cmd, handled := m.visiblePaneEvent(msg); handled {
		return cmd, true
	}
	w := m.workspace
	switch msg := msg.(type) {
	case workspaceClipboard:
		if msg.generation != w.generation || w.conn == nil {
			return nil, true
		}
		w.clipboardReading = false
		if msg.err != nil {
			m.status = "Copy channel unavailable; reconnect"
			return nil, true
		}
		if len(msg.text) == 0 || len(msg.text) > terminal.MaxClipboardBytes || !utf8.ValidString(msg.text) || strings.IndexByte(msg.text, 0) >= 0 {
			return m.readWorkspaceClipboard(), true
		}
		if m.clipboard.inflight != nil || (w.visibleCopy != nil && w.visibleCopy.ready) {
			w.pendingClipboard = &msg
			return nil, true
		}
		return m.submitHostClipboard(msg.text, w.generation, "Copy request sent to the terminal"), true
	case workspaceClipboardSent:
		current := m.clipboard.inflight
		if current == nil || msg.token != current.token {
			return nil, true
		}
		m.clipboard.inflight = nil
		if current.generation == w.generation {
			m.status = current.notice
		}
		return m.drainWorkspaceClipboard(), true
	}
	return nil, false
}
