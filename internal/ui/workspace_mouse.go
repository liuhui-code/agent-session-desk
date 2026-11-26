// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/home.go, handleMouse, at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: v2 messages, shared
// flat-list window, exact Session/Run double-click identity and product Tabs.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type workspaceListClick struct {
	at      time.Time
	index   int
	session session.Session
}

func (m *Model) workspaceClick(msg tea.MouseClickMsg) tea.Cmd {
	return tea.Batch(m.finishWorkspaceSplit(), m.workspaceClickNow(msg))
}

func (m *Model) workspaceClickNow(msg tea.MouseClickMsg) tea.Cmd {
	w := m.workspace
	w.programHover.latest = nil
	previous := w.listClick
	w.listClick = workspaceListClick{}
	if w.transition != workspaceIdle {
		return nil
	}
	w.split.dragging = false
	w.split.hovered = false
	if cmd, handled := m.startProgramPointer(msg); handled {
		return cmd
	}
	if msg.Button != tea.MouseLeft {
		return nil
	}
	if cmd, handled := m.startCopyPointer(msg); handled {
		return cmd
	}
	g := m.workspaceGeometry()
	if m.workspaceDividerHit(msg.X, msg.Y) {
		w.split.dragging = true
		w.split.moved = false
		w.split.dragStart = msg.X
		return nil
	}
	listTop := g.listY
	listHeight := g.listHeight
	if msg.X >= g.listX && msg.X < g.listX+g.listWidth && msg.Y >= listTop && msg.Y < listTop+listHeight {
		window := m.sidebarNavigationWindow(g.listWidth, listHeight)
		w.focused = false
		m.status = ""
		for _, hit := range window.rows {
			if msg.Y < listTop+hit.top || msg.Y >= listTop+hit.top+hit.height {
				continue
			}
			m.viewOffset = window.start
			if hit.index < 0 {
				return tea.Batch(m.toggleSidebarGroup(hit.groupID), m.workspaceSize())
			}
			index := hit.index
			m.cursor = index
			m.invalidatePreview()
			now := time.Now()
			elapsed := now.Sub(previous.at)
			if previous.index == index && sameRun(previous.session, m.sessions[index]) && elapsed >= 0 && elapsed < m.tokens.workspace.doubleClickInterval {
				return m.workspaceKey(tea.KeyPressMsg{Code: tea.KeyEnter})
			}
			w.listClick = workspaceListClick{at: now, index: index, session: m.sessions[index]}
			return m.workspaceSize()
		}
		return m.workspaceSize()
	}
	if msg.X >= g.x && msg.X < g.x+g.width && msg.Y >= g.y && msg.Y < g.y+g.height {
		if !w.focused {
			w.focused = true
			return m.workspaceSize()
		}
	}
	if g.left > 0 && msg.X >= g.listX && msg.X < g.listX+g.left && msg.Y >= g.listY && msg.Y < g.listY+g.listHeight {
		w.focused = false
		return m.workspaceSize()
	}
	return nil
}
