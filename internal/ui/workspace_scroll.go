// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/home.go, mouse wheel routing, at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: v2 messages and
// native psmux scrollback, shared product list/terminal geometry, FIFO writes.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package ui

import tea "charm.land/bubbletea/v2"

func (m *Model) workspaceWheel(msg tea.MouseWheelMsg) tea.Cmd {
	return tea.Batch(m.finishProgramPointer(), m.finishWorkspaceSplit(), m.workspaceWheelNow(msg))
}

func (m *Model) workspaceWheelNow(msg tea.MouseWheelMsg) tea.Cmd {
	w := m.workspace
	w.copyPointer = nil
	if w.transition != workspaceIdle {
		return nil
	}
	w.split.dragging = false
	w.split.hovered = false
	w.listClick = workspaceListClick{}
	delta := 1
	switch msg.Button {
	case tea.MouseWheelUp:
		delta = -1
	case tea.MouseWheelDown:
	default:
		return nil
	}
	g := m.workspaceGeometry()
	listTop := g.y - g.above - m.styles.terminalFrame.GetBorderTopSize() + m.tokens.layout.headingRows
	listHeight := max(0, g.listHeight-m.tokens.layout.headingRows)
	if msg.X >= 0 && msg.X < g.listWidth && msg.Y >= listTop && msg.Y < listTop+listHeight {
		w.focused = false
		m.status = ""
		return tea.Batch(m.move(delta), m.workspaceSize())
	}
	width, height := min(g.width, w.screen.Width), min(g.height, w.screen.Height)
	if w.conn == nil || w.busy || w.err != nil || msg.X < g.x || msg.X >= g.x+width || msg.Y < g.y || msg.Y >= g.y+height {
		return nil
	}
	return m.workspaceWrite(workspaceInput{kind: "scroll", x: msg.X - g.x, y: msg.Y - g.y, up: delta < 0})
}
