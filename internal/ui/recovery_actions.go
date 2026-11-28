package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) newSessionFromRecovery(d *recoveryCenter) tea.Cmd {
	if d.submitted || d.busy || d.review.Launch == nil {
		d.message = "No complete original launch snapshot; no command was guessed"
		return nil
	}
	if m.groupOrder != nil || m.sessionOrder != nil || m.workspace.transition != workspaceIdle {
		d.message = "Finish the current navigation before creating a Session"
		return nil
	}
	if _, ok := m.service.(launchService); !ok {
		d.message = "New terminals unavailable in this host"
		return nil
	}
	for _, entry := range d.review.Launch.Env {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(strings.ToUpper(key), "ASD_RECOVERY_") {
			d.message = "Recovery handoff launches cannot be cloned; restore with fresh credentials instead"
			return nil
		}
	}
	seed := *d.review.Launch
	m.launchSeed = &seed
	m.recoveryUI.center = nil
	return m.navigateWorkspace(workspaceLaunchEdit)
}

func (m *Model) reattachRecoverySource(d *recoveryCenter) tea.Cmd {
	w := m.workspace
	if w.busy || w.transition != workspaceIdle {
		d.message = "Native view busy; wait before reconnecting"
		return nil
	}
	if m.listErr != nil {
		d.message = "Backend observation unknown; refresh before reconnecting"
		return nil
	}
	source := d.review.Source
	var target session.Session
	for _, row := range m.sessions {
		if activityKey(row) == activityKey(source) && row.TerminalState == session.TerminalAvailable {
			target = row
			break
		}
	}
	if target.ID == "" {
		d.message = "Original terminal not observed as available; no restart was sent"
		return nil
	}
	index := -1
	for i, tab := range w.tabs {
		if activityKey(tab.session) == activityKey(target) {
			index = i
			break
		}
	}
	if index < 0 {
		w.nextID++
		w.tabs = append(w.tabs, workspaceTab{viewID: w.nextID, session: target, live: true})
		index = len(w.tabs) - 1
	}
	m.recoveryUI.center = nil
	w.focused = true
	m.selectWorkspaceSession(target)
	return m.activateWorkspace(index)
}
