// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/search.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: scoped live Session/Run targets, stable v2 input and no agent status,
// Recall or group assumptions. See third_party/agent-deck/LICENSE.
package ui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const searchQueryLimit = 100 // input bound, matching the pinned local search

type sessionSearch struct {
	input   textinput.Model
	results []session.Session
	cursor  int
	notice  string
}

func (m *Model) openSessionSearch() tea.Cmd {
	d := &sessionSearch{input: textinput.New()}
	d.input.CharLimit = searchQueryLimit
	d.input.Placeholder = "Search sessions…"
	d.input.Prompt = m.tokens.search.prompt
	d.input.SetVirtualCursor(false)
	d.input.SetStyles(searchInputStyles(m.tokens.forDisplay(m.colorProfile)))
	d.input.KeyMap.Paste.SetEnabled(false)
	m.search = d
	m.workspace.focused = false
	m.syncSessionSearch(false)
	m.sizeSessionSearch()
	return d.input.Focus()
}

func (m *Model) syncSessionSearch(preserve bool) {
	d := m.search
	if d == nil {
		return
	}
	var previous session.Session
	hadSelection := preserve && d.cursor >= 0 && d.cursor < len(d.results)
	if hadSelection {
		previous = d.results[d.cursor]
	}
	d.results = filterSessionSearch(m.sessions, d.input.Value())
	d.cursor = 0
	if preserve {
		d.cursor = -1
		for i, s := range d.results {
			if hadSelection && sameRun(s, previous) {
				d.cursor = i
				break
			}
		}
		if hadSelection && d.cursor < 0 {
			d.notice = "Run changed; select again"
		}
	}
}

func (m *Model) sessionSearchEvent(msg tea.Msg) (tea.Cmd, bool) {
	d := m.search
	if d == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		m.sizeSessionSearch()
		return nil, true
	case tea.KeyReleaseMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
		return nil, true
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.search = nil
			return m.connectWorkspace(), true
		case "up", "ctrl+k":
			d.cursor = max(0, d.cursor-1)
			return nil, true
		case "down", "ctrl+j":
			d.cursor = min(max(0, len(d.results)-1), d.cursor+1)
			return nil, true
		case "enter":
			if d.cursor < 0 || d.cursor >= len(d.results) {
				return nil, true
			}
			target := d.results[d.cursor]
			for i, s := range m.sessions {
				if sameRun(s, target) {
					if s.TerminalState != session.TerminalAvailable {
						d.notice = "Target terminal is not interactive"
						return nil, true
					}
					m.cursor, m.search = i, nil
					m.revealSidebarSession()
					return m.workspaceKey(tea.KeyPressMsg{Code: tea.KeyEnter}), true
				}
			}
			m.syncSessionSearch(true)
			return nil, true
		case "tab", "shift+tab", "ctrl+pgup", "ctrl+pgdown", "f6":
			return nil, true
		}
		if len(d.input.Value())+len(msg.Text) > searchQueryLimit*4 {
			return nil, true
		}
		before := d.input.Value()
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(msg)
		if d.input.Value() != before {
			d.notice = ""
			m.syncSessionSearch(false)
		}
		return cmd, true
	case tea.PasteMsg:
		if strings.ContainsAny(msg.Content, "\r\n\t") || len(d.input.Value())+len(msg.Content) > searchQueryLimit*4 {
			d.notice = "Search requires bounded single-line text"
			return nil, true
		}
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(msg)
		d.notice = ""
		m.syncSessionSearch(false)
		return cmd, true
	}
	return nil, false
}
