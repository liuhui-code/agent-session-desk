package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// This is a read bound, not a layout token or a complete-history guarantee.
const historyLimit = 100

type historyState struct {
	open, detail, busy bool
	rows               []session.HistoryRecord
	cursor, offset     int
	err                error
	generation         uint64
	cancel             context.CancelFunc
}
type historyMsg struct {
	generation uint64
	rows       []session.HistoryRecord
	err        error
}

func (m *Model) openHistory() tea.Cmd {
	m.history.open = true
	m.history.detail = false
	m.history.offset = 0
	m.help = false
	m.invalidatePreview()
	m.history.generation++
	return m.loadHistory()
}
func (m *Model) cancelHistory() {
	m.history.generation++
	if m.history.cancel != nil {
		m.history.cancel()
		m.history.cancel = nil
	}
}
func (m *Model) closeHistory() tea.Cmd {
	m.cancelHistory()
	m.history.open = false
	m.history.detail = false
	m.history.rows = nil
	m.history.err = nil
	m.history.cursor = 0
	return m.loadPreview()
}
func (m *Model) loadHistory() tea.Cmd {
	if !m.history.open || m.history.busy || m.closed {
		return nil
	}
	m.history.busy = true
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	m.history.cancel = cancel
	generation := m.history.generation
	return func() tea.Msg {
		rows, err := m.service.History(ctx, historyLimit)
		return historyMsg{generation, rows, err}
	}
}
func (m *Model) historyResult(msg historyMsg) tea.Cmd {
	m.history.busy = false
	if m.history.cancel != nil {
		m.history.cancel()
		m.history.cancel = nil
	}
	if msg.generation != m.history.generation || !m.history.open {
		return m.loadHistory()
	}
	chosen := ""
	if len(m.history.rows) > 0 {
		chosen = m.history.rows[m.history.cursor].Session.RunID
	}
	m.history.rows = msg.rows
	m.history.err = msg.err
	m.history.cursor = 0
	found := false
	for i, row := range msg.rows {
		if row.Session.RunID == chosen {
			m.history.cursor = i
			found = true
			break
		}
	}
	if !found {
		m.history.detail = false
		m.history.offset = 0
	}
	return nil
}
func (m *Model) historyKey(key string) tea.Cmd {
	if m.history.busy && key != "esc" && key != "h" {
		return nil
	}
	if m.history.detail && key != "esc" && key != "h" && key != "r" {
		m.scrollHistoryDetail(key)
		return nil
	}
	switch key {
	case "r":
		m.history.err = nil
		return m.loadHistory()
	case "enter":
		if !m.history.busy && m.history.err == nil && len(m.history.rows) > 0 {
			m.history.detail = true
			m.history.offset = 0
		}
	case "esc", "h":
		if m.history.detail {
			m.history.detail = false
		} else {
			return m.closeHistory()
		}
	case "up", "k":
		m.history.cursor = max(0, m.history.cursor-1)
	case "down", "j":
		m.history.cursor = min(max(0, len(m.history.rows)-1), m.history.cursor+1)
	case "home", "g":
		m.history.cursor = 0
	case "end", "G":
		m.history.cursor = max(0, len(m.history.rows)-1)
	case "pgdown":
		m.history.cursor = min(max(0, len(m.history.rows)-1), m.history.cursor+m.historyPageSize())
	case "pgup":
		m.history.cursor = max(0, m.history.cursor-m.historyPageSize())
	}
	return nil
}

func (m *Model) historyPageSize() int {
	return max(1, m.height-m.tokens.chromeRows()-m.tokens.layout.headingRows)
}

func (m *Model) scrollHistoryDetail(key string) {
	h := &m.history
	last := max(0, len(m.historyDetailLines())-m.historyPageSize())
	switch key {
	case "up", "k":
		h.offset--
	case "down", "j":
		h.offset++
	case "pgup":
		h.offset -= m.historyPageSize()
	case "pgdown":
		h.offset += m.historyPageSize()
	case "home", "g":
		h.offset = 0
	case "end", "G":
		h.offset = last
	}
	h.offset = max(0, min(last, h.offset))
}
