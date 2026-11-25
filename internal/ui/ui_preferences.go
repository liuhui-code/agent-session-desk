package ui

import (
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// One writer covers the whole small preference document, including choices
// accepted while a previous write is pending. No terminal effect is involved.
type uiPreferenceState struct {
	saved, wanted, accepted, writing config.UISettings
	forgottenFolds                   map[string]bool
	busy                             bool
	queued                           bool // An explicit preference accepted while the writer was busy.
	generation                       uint64
}
type uiPreferencesSaved struct {
	generation uint64
	settings   config.UISettings
	err        error
}

func (m *Model) saveUIPreferences() tea.Cmd {
	m.preferences.accepted = m.preferences.wanted
	return m.writeUIPreferences()
}

func (m *Model) writeUIPreferences() tea.Cmd {
	p := &m.preferences
	p.accepted.CollapsedGroups = withoutGroupFolds(p.accepted.CollapsedGroups, p.forgottenFolds)
	p.wanted.CollapsedGroups = withoutGroupFolds(p.wanted.CollapsedGroups, p.forgottenFolds)
	p.retainForgottenFolds()
	if p.busy {
		p.queued = true
		return nil
	}
	if p.saved == p.accepted {
		if p.wanted == p.accepted {
			m.applySidebarNavigation(p.saved)
		}
		return nil
	}
	p.busy = true
	p.generation++
	p.writing = p.accepted
	snapshot, generation := p.writing, p.generation
	save, densitySave, ctx, previousDensity, previousTheme := m.saveUISettings, m.saveSidebarDensity, m.ctx, p.saved.SidebarDensity, p.saved.Theme
	return func() tea.Msg {
		var err error
		if save != nil {
			err = save(ctx, snapshot)
		} else if snapshot.Theme != previousTheme {
			err = errors.New("Theme persistence is unsupported; not saved")
		} else if densitySave != nil && snapshot.SidebarDensity != previousDensity {
			err = densitySave(ctx, snapshot.SidebarDensity)
		}
		return uiPreferencesSaved{generation, snapshot, err}
	}
}
func (m *Model) uiPreferencesEvent(msg tea.Msg) (tea.Cmd, bool) {
	r, ok := msg.(uiPreferencesSaved)
	if !ok {
		return nil, false
	}
	p := &m.preferences
	if !p.busy || r.generation != p.generation {
		return nil, true
	}
	p.busy = false
	p.writing = config.UISettings{}
	queued := p.queued
	p.queued = false
	if r.err != nil {
		if m.sessionOrder != nil {
			m.sessionOrder.quitting = false
		}
		if m.groupOrder != nil {
			m.groupOrder.quitting = false
		}
		p.wanted = p.saved
		p.wanted.CollapsedGroups = withoutGroupFolds(p.wanted.CollapsedGroups, p.forgottenFolds)
		p.accepted = p.wanted
		p.retainForgottenFolds()
		m.sidebarDensity = p.saved.SidebarDensity
		m.applyUITheme(p.saved.Theme)
		m.applySidebarNavigation(p.wanted)
		resize := m.restoreWorkspaceSplit(p.saved.SidebarPercent)
		m.sidebarPreferenceError = r.err
		m.status = "UI preferences not saved: " + label(r.err.Error())
		if m.workspace.transition == workspaceQuit {
			m.workspace.transition = workspaceIdle
		}
		if d := m.densitySettings; d != nil {
			d.busy, d.closing, d.err = false, false, r.err
		}
		return tea.Batch(resize, m.syncSystemTheme()), true
	}
	p.saved = r.settings
	p.retainForgottenFolds()
	m.sidebarDensity = r.settings.SidebarDensity
	m.applyUITheme(r.settings.Theme)
	themeCmd := m.syncSystemTheme()
	if r.settings.SidebarMode == p.wanted.SidebarMode && r.settings.DefaultGroupCollapsed == p.wanted.DefaultGroupCollapsed && r.settings.CollapsedGroups == p.wanted.CollapsedGroups {
		m.applySidebarNavigation(r.settings)
	}
	m.sidebarPreferenceError = nil
	if strings.HasPrefix(m.status, "UI preferences not saved: ") || strings.HasPrefix(m.status, "UI preferences not loaded") {
		m.status = ""
	}
	if d := m.densitySettings; d != nil {
		d.busy, d.err = false, nil
		if sidebarDensities[d.choice] != m.sidebarDensity || uiThemes[d.themeChoice] != effectiveUITheme(r.settings.Theme) {
			p.wanted = m.settingsChoice()
			p.accepted = p.wanted
			d.busy = true
			queued = true
		}
	}
	var cmd tea.Cmd
	if queued {
		cmd = m.writeUIPreferences()
	}
	if !p.busy && m.workspace.transition == workspaceQuit {
		m.workspace.transition = workspaceIdle
		return tea.Batch(themeCmd, cmd, m.workspaceKey(tea.KeyPressMsg{Code: 'q', Text: "q"})), true
	}
	if d := m.densitySettings; d != nil && d.closing && !p.busy {
		return tea.Batch(themeCmd, cmd, m.closeDensitySettings()), true
	}
	return tea.Batch(themeCmd, cmd), true
}
func (m *Model) applySidebarNavigation(settings config.UISettings) {
	changed := m.sidebarMode != settings.SidebarMode || m.sidebarCollapsed != settings.DefaultGroupCollapsed || strings.Join(sortedCollapsedIDs(m.collapsedGroups), ",") != settings.CollapsedGroups
	m.sidebarMode, m.sidebarCollapsed = settings.SidebarMode, settings.DefaultGroupCollapsed
	m.collapsedGroups = map[string]bool{}
	for _, id := range strings.Split(settings.CollapsedGroups, ",") {
		if id != "" {
			m.collapsedGroups[id] = true
		}
	}
	if changed {
		m.viewOffset = 0
		m.workspace.listClick = workspaceListClick{}
	}
	if m.groupedSidebar() {
		if v, ok := m.selected(); ok {
			id := session.GroupOf(v)
			for depth := 0; id != "" && depth < session.MaxGroupDepth; depth++ {
				if m.groupCollapsed(id) {
					m.cursor, m.selectedGroupID = -1, id
				}
				id = m.groupByID(id).ParentID
			}
		} else if len(m.sessions) == 0 {
			m.cursor = -1
		}
	} else if m.cursor < 0 && !m.groupedSidebar() {
		m.cursor = 0
	}
}
