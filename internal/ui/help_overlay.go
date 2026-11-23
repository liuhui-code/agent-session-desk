// Help interaction adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/help.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: scoped product keys, native Tab/input fences and stable v2 events.
package ui

import tea "charm.land/bubbletea/v2"

type helpPage struct {
	all    bool
	offset int
}

func (m *Model) helpEvent(msg tea.Msg) (tea.Cmd, bool) {
	if !m.help {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "tab" {
			m.helpPage.all = !m.helpPage.all
			m.helpPage.offset = 0
			return nil, true
		}
		if m.helpPage.all {
			_, _, _, limit := m.fullHelpPage()
			offset := min(m.helpPage.offset, limit)
			switch msg.String() {
			case "j", "down":
				offset++
			case "k", "up":
				offset--
			case "ctrl+d", "pgdown":
				offset += m.tokens.help.pageStep
			case "ctrl+u", "pgup":
				offset -= m.tokens.help.pageStep
			case "home", "g":
				offset = 0
			case "end", "G":
				offset = limit
			default:
				m.help = false
				return m.connectWorkspace(), true
			}
			m.helpPage.offset = max(0, min(offset, limit))
			return nil, true
		}
		m.help = false
		return m.connectWorkspace(), true
	case tea.PasteMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
		return nil, true
	}
	return nil, false
}
