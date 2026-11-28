package ui

import tea "charm.land/bubbletea/v2"

func (m *Model) workspaceModelEvent(msg tea.Msg) (tea.Cmd, bool) {
	if cmd, handled := m.attentionEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.attentionCenterEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.sessionStopEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.helpEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.groupRemovalEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.sessionOrderEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.groupOrderEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.groupDialogEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.uiPreferencesEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.sessionRenameEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.sessionSearchEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.densitySettingsEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.quickLaunchEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.launchEvent(msg); handled {
		return cmd, true
	}
	if cmd, handled := m.workspaceEvent(msg); handled {
		return cmd, true
	}
	return nil, false
}
