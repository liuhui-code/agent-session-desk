package ui

import tea "charm.land/bubbletea/v2"

type workspaceTabSlot struct{ index, x, width int }
type workspaceTabDrag struct {
	viewID uint64
	slots  []workspaceTabSlot
}

func (m *Model) pressWorkspaceTab(index int) tea.Cmd {
	w := m.workspace
	if w.busy {
		m.status = "当前 Tab 正在连接；连接完成后可切换"
		return nil
	}
	drag := workspaceTabDrag{viewID: w.tabs[index].viewID}
	// Freeze the visible slots for this gesture. Different title widths must
	// not shift the drop thresholds back under a stationary pointer.
	for _, cell := range m.workspaceTabCells() {
		drag.slots = append(drag.slots, workspaceTabSlot{cell.index, cell.x, cell.width})
	}
	w.focused = true
	var cmd tea.Cmd
	if index == w.active {
		cmd = m.workspaceSize()
	} else {
		cmd = m.activateWorkspace(index)
	}
	w.tabDrag = drag
	return cmd
}

func (m *Model) dragWorkspaceTab(x, y int) {
	w := m.workspace
	if y != 1 || x < 0 || x >= m.workspaceTabBudget() || w.transition > workspaceSwitch {
		w.tabDrag = workspaceTabDrag{}
		return
	}
	slots := w.tabDrag.slots
	for _, slot := range slots {
		if x >= slot.x && x < slot.x+slot.width {
			m.moveWorkspaceTab(w.tabDrag.viewID, slot.index)
			return
		}
	}
	if len(slots) > 0 && x >= slots[len(slots)-1].x+slots[len(slots)-1].width {
		m.moveWorkspaceTab(w.tabDrag.viewID, slots[len(slots)-1].index)
	}
}

func (m *Model) workspaceTabMotion(msg tea.MouseMotionMsg) bool {
	if m.workspace.tabDrag.viewID == 0 {
		return false
	}
	if msg.Button != tea.MouseLeft {
		m.workspace.tabDrag = workspaceTabDrag{}
		return true
	}
	m.dragWorkspaceTab(msg.X, msg.Y)
	return true
}

func (m *Model) workspaceTabRelease(msg tea.MouseReleaseMsg) bool {
	if m.workspace.tabDrag.viewID == 0 {
		return false
	}
	if msg.Button == tea.MouseLeft {
		m.dragWorkspaceTab(msg.X, msg.Y)
	}
	m.workspace.tabDrag = workspaceTabDrag{}
	return true
}
