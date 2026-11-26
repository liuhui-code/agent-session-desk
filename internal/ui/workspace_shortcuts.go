package ui

import (
	tea "charm.land/bubbletea/v2"
)

func (m *Model) workspaceRenameKey() tea.Cmd {
	w := m.workspace
	if w.transition != workspaceIdle || w.busy || w.err != nil || m.listErr != nil {
		return nil
	}
	if m.groupOrder != nil || m.sessionOrder != nil {
		m.status = "Saving order; wait before renaming"
		return nil
	}
	if w.focused && w.active >= 0 && w.active < len(w.tabs) {
		m.selectWorkspaceSession(w.tabs[w.active].session)
	}
	if !w.focused && m.sidebarGroupSelected() {
		return m.beginGroupRename()
	}
	return m.beginSessionRename()
}
