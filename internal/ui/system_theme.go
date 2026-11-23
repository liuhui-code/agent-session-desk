package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/appearance"
)

type systemThemeState struct {
	choice, resolved, source string
	known, failed, blocked   bool
	open                     func(context.Context) (appearance.Connection, error)
	op                       *systemThemeOperation
}

type systemThemeOperation struct {
	ctx              context.Context
	cancel           context.CancelFunc
	conn             appearance.Connection
	opening, closing bool
	invalid          bool
	reading          bool
	sequence         uint64
}

type systemThemeOpened struct {
	op   *systemThemeOperation
	conn appearance.Connection
	err  error
}
type systemThemeRead struct {
	op       *systemThemeOperation
	sequence uint64
	value    appearance.Observation
	err      error
}
type systemThemeClosed struct {
	op  *systemThemeOperation
	err error
}

// Opening and closing are serialized even if a user switches away and back
// while an OS call is still completing. No theme poll or terminal I/O is used.
func (m *Model) syncSystemTheme() tea.Cmd {
	s := &m.systemTheme
	if op := s.op; op != nil {
		if op.invalid && !op.opening && !op.closing {
			op.closing = true
			return func() tea.Msg { return systemThemeClosed{op, op.conn.Close()} }
		}
		return nil
	}
	if s.choice != "system" || s.failed || s.blocked || m.closed {
		return nil
	}
	ctx, cancel := context.WithCancel(m.ctx)
	op := &systemThemeOperation{ctx: ctx, cancel: cancel, opening: true}
	s.op, s.known = op, false
	open := s.open
	if open == nil {
		open = appearance.Open
	}
	return func() tea.Msg {
		conn, err := open(ctx)
		return systemThemeOpened{op, conn, err}
	}
}

func readSystemTheme(op *systemThemeOperation) tea.Cmd {
	if op.reading || op.invalid || op.closing || op.conn == nil {
		return nil
	}
	op.sequence++
	op.reading = true
	sequence := op.sequence
	return func() tea.Msg {
		v, err := op.conn.Next(op.ctx)
		return systemThemeRead{op, sequence, v, err}
	}
}

func (m *Model) systemThemeEvent(msg tea.Msg) (tea.Cmd, bool) {
	s := &m.systemTheme
	switch v := msg.(type) {
	case systemThemeOpened:
		if v.op != s.op || !v.op.opening {
			return nil, true // Commands emit one receipt; duplicate receipts are inert.
		}
		v.op.opening, v.op.conn = false, v.conn
		if v.op.invalid && v.conn == nil {
			s.op = nil
			return m.syncSystemTheme(), true
		}
		if v.op.invalid {
			return m.syncSystemTheme(), true
		}
		if v.err != nil || v.conn == nil {
			s.known, s.failed = false, true
			v.op.invalid = true
			v.op.cancel()
			if v.conn == nil {
				s.op = nil
				return nil, true
			}
		}
		if v.op.invalid {
			return m.syncSystemTheme(), true
		}
		return readSystemTheme(v.op), true
	case systemThemeRead:
		if v.op != s.op || !v.op.reading || v.sequence != v.op.sequence {
			return nil, true
		}
		v.op.reading = false
		if v.op.invalid || s.choice != "system" {
			return m.syncSystemTheme(), true
		}
		if v.err != nil {
			s.known, s.failed = false, true
			v.op.invalid = true
			v.op.cancel()
			return m.syncSystemTheme(), true
		}
		s.known = v.value.Err == nil && (v.value.Theme == "dark" || v.value.Theme == "light") && (v.value.Source == "terminal" || v.value.Source == "windows")
		if s.known {
			s.resolved, s.source = v.value.Theme, v.value.Source
			m.drawUITheme()
		}
		return readSystemTheme(v.op), true
	case systemThemeClosed:
		if v.op != s.op || !v.op.closing {
			return nil, true
		}
		s.op = nil
		if v.err != nil {
			s.blocked, s.known = true, false
		}
		return m.syncSystemTheme(), true
	}
	return nil, false
}

func (m *Model) retrySystemTheme() tea.Cmd {
	s := &m.systemTheme
	if s.choice != "system" || s.blocked {
		return nil
	}
	s.failed = false
	if s.op != nil {
		s.op.invalid = true
		s.op.cancel()
	}
	return m.syncSystemTheme()
}
