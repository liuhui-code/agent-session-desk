package ui

import tea "charm.land/bubbletea/v2"

func (m *Model) workspaceClick(msg tea.MouseClickMsg) tea.Cmd {
	w := m.workspace
	if msg.Button != tea.MouseLeft || w.transition != workspaceIdle {
		return nil
	}
	w.split.dragging = false
	if msg.Y == 1 {
		button := m.workspaceNewTabCell()
		if msg.X >= button.x && msg.X < button.x+button.width {
			return m.navigateWorkspace(workspaceLaunch)
		}
		for _, cell := range m.workspaceTabCells() {
			if msg.X >= cell.x && msg.X < cell.x+cell.width {
				if msg.X == cell.closeX {
					return m.closeWorkspaceView(cell.index)
				}
				w.focused = true
				if cell.index == w.active {
					return m.workspaceSize()
				}
				return m.activateWorkspace(cell.index)
			}
		}
	}
	g := m.workspaceGeometry()
	if m.workspaceDividerHit(msg.X, msg.Y) {
		w.split.dragging = true
		w.split.moved = false
		w.split.dragStart = msg.X
		return nil
	}
	if msg.X >= g.x && msg.X < g.x+g.width && msg.Y >= g.y && msg.Y < g.y+g.height {
		if !w.focused {
			w.focused = true
			return m.workspaceSize()
		}
	}
	if g.left > 0 && msg.X >= 0 && msg.X < g.left && msg.Y >= g.y-1 && msg.Y < g.y+g.height {
		w.focused = false
		return m.workspaceSize()
	}
	return nil
}
