// Navigation adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Sources: internal/ui/home.go, home_flatitems.go and embedded_sidebar.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: stable persisted product
// groups over discovered Session/Run rows; no agent state or terminal effects.
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) groupedSidebar() bool { return m.workspace != nil && m.sidebarMode == "grouped" }
func (m *Model) sidebarGroupSelected() bool {
	return m.groupedSidebar() && (m.cursor < 0 || len(m.sessions) == 0)
}
func (m *Model) sidebarTitle() string {
	if m.groupedSidebar() {
		return "会话 分组"
	}
	return "会话 平铺"
}
func (m *Model) moveGroupedSidebar(delta int) tea.Cmd {
	items := m.sidebarItems()
	if len(items) == 0 {
		return nil
	}
	selected := m.sidebarSelectedItem(items)
	m.selectSidebarItem(items[min(max(0, selected+delta), len(items)-1)])
	m.invalidatePreview()
	return nil
}
func (m *Model) toggleDefaultGroup() tea.Cmd {
	return m.toggleSidebarGroup(session.DefaultGroupID)
}
func (m *Model) toggleSidebarGroup(id string) tea.Cmd {
	m.cursor, m.selectedGroupID = -1, id
	collapsed := false
	if id == session.DefaultGroupID {
		collapsed = m.preferences.wanted.DefaultGroupCollapsed
	} else {
		for _, value := range strings.Split(m.preferences.wanted.CollapsedGroups, ",") {
			if value == id {
				collapsed = true
			}
		}
	}
	m.wantedGroupCollapse(id, !collapsed)
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
	case "left":
		g := m.selectedGroup()
		if m.sidebarGroupSelected() && !m.groupCollapsed(g.ID) {
			m.wantedGroupCollapse(g.ID, true)
			return m.saveUIPreferences(), true
		}
		if m.sidebarGroupSelected() && g.ParentID != "" {
			g = m.groupByID(g.ParentID)
		}
		m.cursor, m.selectedGroupID = -1, g.ID
		return nil, true
	case "enter", "right", "l", "tab":
		if m.sidebarGroupSelected() {
			return m.toggleSidebarGroup(m.selectedGroup().ID), true
		}
	}
	if len(msg.String()) == 1 && msg.String() >= "1" && msg.String() <= "9" {
		n := int(msg.String()[0] - '1')
		for _, g := range m.sidebarGroups() {
			if g.ParentID != "" {
				continue
			}
			if n == 0 {
				m.cursor, m.selectedGroupID, m.viewOffset = -1, g.ID, 0
				return nil, true
			}
			n--
		}
		return nil, true
	}
	return nil, false
}

// Revealing a Tab/search target changes only presentation. A later explicit
// preference change carries the visible expansion; no terminal input is sent.
func (m *Model) revealSidebarSession() {
	if m.groupedSidebar() {
		if v, ok := m.selected(); ok {
			m.revealGroup(session.GroupOf(v))
		}
	}
}
