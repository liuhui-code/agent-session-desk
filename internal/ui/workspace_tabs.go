package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

// Tab order is view state. Keep the live connection and both target identities
// while moving a view, including during a write or an identity-bound switch.
func (m *Model) moveWorkspaceTab(viewID uint64, destination int) {
	w := m.workspace
	if len(w.tabs) < 2 || destination < 0 || destination >= len(w.tabs) {
		return
	}
	activeID := w.tabs[w.active].viewID
	for source, tab := range w.tabs {
		if tab.viewID != viewID || source == destination {
			continue
		}
		if source < destination {
			copy(w.tabs[source:destination], w.tabs[source+1:destination+1])
		} else {
			copy(w.tabs[destination+1:source+1], w.tabs[destination:source])
		}
		w.tabs[destination] = tab
		for index := range w.tabs {
			if w.tabs[index].viewID == activeID {
				w.active = index
				break
			}
		}
		m.status = "Tab order updated; run and input targets are unchanged"
		return
	}
}

func (m *Model) closeWorkspaceTabNow() tea.Cmd {
	w := m.workspace
	if len(w.tabs) == 0 {
		w.focused = false
		return nil
	}
	w.transition = workspaceIdle
	w.disconnect()
	w.tabs = append(w.tabs[:w.active], w.tabs[w.active+1:]...)
	w.active = min(w.active, max(0, len(w.tabs)-1))
	w.err = nil
	w.screen = terminal.Screen{}
	w.buffered = nil
	w.transition = workspaceIdle
	m.status = "View closed; the run continues. Reopen it from the sidebar."
	if len(w.tabs) == 0 {
		w.focused = false
		m.status = "Navigation focus; n New, Enter Reopen. The original run continues."
		return nil
	}
	return m.activateWorkspaceNow(w.active)
}

func (m *Model) closeWorkspaceTab() tea.Cmd {
	w := m.workspace
	if w.pending > 0 || (w.busy && len(w.buffered) > 0) {
		w.transition = workspaceClose
		m.status = "Finishing input before closing the view…"
		return nil
	}
	return m.closeWorkspaceTabNow()
}

func (m *Model) closeWorkspaceView(index int) tea.Cmd {
	w := m.workspace
	if index < 0 || index >= len(w.tabs) {
		return nil
	}
	if index == w.active {
		return m.closeWorkspaceTab()
	}
	w.tabs = append(w.tabs[:index], w.tabs[index+1:]...)
	if index < w.active {
		w.active--
	}
	m.status = "View closed; the run continues. Reopen it from the sidebar."
	return nil
}
