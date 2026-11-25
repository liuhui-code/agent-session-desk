package ui

import tea "charm.land/bubbletea/v2"

func (m *Model) toggleWorkspaceSidebar() tea.Cmd {
	w := m.workspace
	w.sidebarHidden = !w.sidebarHidden
	if w.sidebarHidden {
		w.sidebarFocusBeforeHide = w.focused
		w.focused = true
	} else {
		w.focused = w.sidebarFocusBeforeHide
	}
	w.split.hovered = false
	w.listClick = workspaceListClick{}
	return m.workspaceSize()
}
