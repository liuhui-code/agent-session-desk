package ui

import tea "charm.land/bubbletea/v2"

// Management navigation waits for writes already accepted by this view. It
// never retries an uncertain write or transfers it to another Run.
func (m *Model) navigateWorkspace(target workspaceTransition) tea.Cmd {
	if m.groupOrder != nil || m.sessionOrder != nil {
		m.status = "Saving order; wait before continuing"
		return nil
	}
	w := m.workspace
	if w.transition != workspaceIdle {
		return nil
	}
	if target == workspaceLaunch || target == workspaceLaunchEdit {
		if _, ok := m.service.(groupService); ok {
			m.launchGroup = m.selectedGroup()
		}
	}
	w.transition = target
	if w.pending > 0 || (w.busy && len(w.buffered) > 0) {
		m.status = "Finishing input before opening navigation…"
		return nil
	}
	return m.finishWorkspaceNavigation()
}
func (m *Model) finishWorkspaceNavigation() tea.Cmd {
	w := m.workspace
	target := w.transition
	w.transition = workspaceIdle
	w.disconnect()
	m.status = ""
	switch target {
	case workspaceHistory:
		return m.openHistory()
	case workspaceHelp:
		m.help = true
		m.helpPage = helpPage{}
	case workspaceAttach:
		return m.attach()
	case workspaceLaunch:
		if m.loadQuickLaunch != nil {
			return m.startQuickLaunch()
		}
		return m.openLaunchDialog()
	case workspaceLaunchEdit:
		return m.openLaunchDialog()
	case workspaceSettings:
		m.openDensitySettings()
	case workspaceSearch:
		return m.openSessionSearch()
	case workspaceRename:
		return m.openSessionRename()
	case workspaceGroup:
		return m.openGroupDialog()
	case workspaceRemoval:
		return m.openGroupRemoval()
	case workspaceStop:
		return m.openSessionStop()
	case workspaceAttention:
		return m.openAttention()
	case workspaceAttentionCenter:
		m.attention.center, m.attention.centerIntent = m.attention.centerIntent, nil
		m.workspace.focused = false
	}
	return nil
}

func (m *Model) refreshWorkspace() tea.Cmd {
	w := m.workspace
	if m.snippets.dialog != nil || w == nil || w.conn == nil || w.busy || w.pending > 0 || w.transition != workspaceIdle || m.attaching || m.history.open || m.help || m.launch != nil {
		return nil
	}
	g := m.workspaceGeometry()
	if g.width < 1 || g.height < 1 {
		return nil
	}
	return m.workspaceWrite(workspaceInput{kind: "refresh"})
}

// Defer only optional polling; explicit metadata reconciliation after a user
// action must still run while its terminal view reconnects.
func (m *Model) pollWorkspace() tea.Cmd {
	if m.workspaceConnecting() {
		return m.schedulePoll()
	}
	return tea.Batch(m.refreshWorkspace(), m.sampleActivity(), m.loadSessions(), m.loadPreview(), m.schedulePoll())
}
