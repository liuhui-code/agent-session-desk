package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

func withoutGroupFolds(value string, removed map[string]bool) string {
	var ids []string
	for _, id := range strings.Split(value, ",") {
		if id != "" && !removed[id] {
			ids = append(ids, id)
		}
	}
	return strings.Join(ids, ",")
}

// Only IDs still referenced by a saved/in-flight document need a suppression
// marker. Each document already has a bounded fold set; no deletion log grows.
func (p *uiPreferenceState) retainForgottenFolds() {
	keep := map[string]bool{}
	for _, value := range []string{p.saved.CollapsedGroups, p.writing.CollapsedGroups} {
		for _, id := range strings.Split(value, ",") {
			if id != "" {
				keep[id] = true
			}
		}
	}
	for id := range p.forgottenFolds {
		if !keep[id] {
			delete(p.forgottenFolds, id)
		}
	}
}

func (m *Model) forgetRemovedGroupFolds(removed map[string]bool) tea.Cmd {
	p := &m.preferences
	if p.forgottenFolds == nil {
		p.forgottenFolds = map[string]bool{}
	}
	for _, value := range []string{p.saved.CollapsedGroups, p.writing.CollapsedGroups} {
		for _, id := range strings.Split(value, ",") {
			if removed[id] {
				p.forgottenFolds[id] = true
			}
		}
	}
	for id := range removed {
		delete(m.collapsedGroups, id)
	}
	p.wanted.CollapsedGroups = withoutGroupFolds(p.wanted.CollapsedGroups, removed)
	p.accepted.CollapsedGroups = withoutGroupFolds(p.accepted.CollapsedGroups, removed)
	// Cleanup updates the last explicit document without adopting unrelated
	// temporary Tab/search reveals from the currently visible wanted state.
	return m.writeUIPreferences()
}
