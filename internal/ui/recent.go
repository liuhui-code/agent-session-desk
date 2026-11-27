package ui

import tea "charm.land/bubbletea/v2"

func (m *Model) rememberWorkspace() {
	w := m.workspace
	if w.active < 0 || w.active >= len(w.tabs) {
		return
	}
	key := activityKey(w.tabs[w.active].session)
	out := []string{key}
	for _, v := range w.recent {
		if v != key && len(out) < 16 {
			out = append(out, v)
		}
	}
	w.recent = out
}
func (m *Model) previousWorkspace() tea.Cmd {
	w := m.workspace
	if len(w.tabs) == 0 {
		return nil
	}
	current := activityKey(w.tabs[w.active].session)
	for _, k := range w.recent {
		if k == current {
			continue
		}
		for i, t := range w.tabs {
			if t.live && activityKey(t.session) == k {
				return m.activateWorkspace(i)
			}
		}
	}
	return nil
}

type pollMsg struct{ generation uint64 }
