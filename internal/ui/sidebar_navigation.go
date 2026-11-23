// Navigation adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Sources: internal/ui/home.go, home_flatitems.go and embedded_sidebar.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: one implicit default
// group over discovered Session/Run rows; no agent state or terminal effects.
package ui

import tea "charm.land/bubbletea/v2"

func (m *Model) groupedSidebar() bool { return m.workspace != nil && m.sidebarMode == "grouped" }
func (m *Model) sidebarGroupSelected() bool {
	return m.groupedSidebar() && (m.cursor < 0 || m.sidebarCollapsed || len(m.sessions) == 0)
}
func (m *Model) sidebarTitle() string {
	if m.groupedSidebar() {
		return "会话 分组"
	}
	return "会话 平铺"
}
func (m *Model) moveGroupedSidebar(delta int) tea.Cmd {
	last := len(m.sessions) - 1
	if m.sidebarCollapsed {
		last = -1
	}
	previous := m.cursor
	m.cursor = min(max(-1, m.cursor+delta), last)
	if previous != m.cursor {
		m.invalidatePreview()
	}
	return nil
}
func (m *Model) toggleDefaultGroup() tea.Cmd {
	m.cursor = -1
	m.preferences.wanted.DefaultGroupCollapsed = !m.preferences.wanted.DefaultGroupCollapsed
	return m.saveUIPreferences()
}
func (m *Model) sidebarNavigationKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	if m.workspace.transition != workspaceIdle {
		return nil, false
	}
	if msg.Code == 'v' && msg.Mod == tea.ModAlt {
		if m.preferences.wanted.SidebarMode == "flat" {
			m.preferences.wanted.SidebarMode = "grouped"
		} else {
			m.preferences.wanted.SidebarMode = "flat"
		}
		return m.saveUIPreferences(), true
	}
	if !m.groupedSidebar() {
		return nil, false
	}
	switch msg.String() {
	case "1":
		m.cursor, m.viewOffset = -1, 0
		return nil, true
	case "left":
		if m.sidebarGroupSelected() && !m.sidebarCollapsed {
			m.preferences.wanted.DefaultGroupCollapsed = true
			return m.saveUIPreferences(), true
		}
		m.cursor = -1
		return nil, true
	case "enter", "right", "l", "tab":
		if m.sidebarGroupSelected() {
			return m.toggleDefaultGroup(), true
		}
	}
	return nil, false
}

// Revealing a Tab/search target changes only presentation. A later explicit
// preference change carries the visible expansion; no terminal input is sent.
func (m *Model) revealSidebarSession() {
	if m.groupedSidebar() && m.sidebarCollapsed {
		m.sidebarCollapsed = false
		m.preferences.wanted.DefaultGroupCollapsed = false
		m.viewOffset = 0
	}
}
