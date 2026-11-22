package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// Discovery can see the backend effect before Start finishes its metadata
// commit. Acknowledgement must reconcile that same Run, not create a second
// list row or view of the newly launched terminal.
func (m *Model) acceptLaunchedSession(s session.Session) tea.Cmd {
	found := false
	for i, v := range m.sessions {
		if sameRun(v, s) {
			m.sessions[i] = s
			found = true
			break
		}
	}
	if !found {
		m.sessions = append(m.sessions, s)
	}
	w := m.workspace
	w.initialized = true
	w.focused = true
	for i, tab := range w.tabs {
		if sameRun(tab.session, s) {
			w.tabs[i].session = s
			w.tabs[i].live = true
			return m.activateWorkspace(i)
		}
	}
	w.nextID++
	w.tabs = append(w.tabs, workspaceTab{viewID: w.nextID, session: s, live: true})
	return m.activateWorkspace(len(w.tabs) - 1)
}
