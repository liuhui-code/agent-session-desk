package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"strings"
)

// One writer covers the whole small preference document, including choices
// accepted while a previous write is pending. No terminal effect is involved.
type uiPreferenceState struct {
	saved, wanted config.UISettings
	busy          bool
	generation    uint64
}
type uiPreferencesSaved struct {
	generation uint64
	settings   config.UISettings
	err        error
}

func (m *Model) saveUIPreferences() tea.Cmd {
	p := &m.preferences
	if p.busy {
		return nil
	}
	if p.saved == p.wanted {
		m.applySidebarNavigation(p.saved)
		return nil
	}
	p.busy = true
	p.generation++
	snapshot, generation := p.wanted, p.generation
	save, densitySave, ctx, previousDensity := m.saveUISettings, m.saveSidebarDensity, m.ctx, p.saved.SidebarDensity
	return func() tea.Msg {
		var err error
		if save != nil {
			err = save(ctx, snapshot)
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
	if r.err != nil {
		p.wanted = p.saved
		m.sidebarDensity = p.saved.SidebarDensity
		m.applySidebarNavigation(p.saved)
		m.sidebarPreferenceError = r.err
		m.status = "管理模式 · 侧栏偏好未保存：" + label(r.err.Error())
		if m.workspace.transition == workspaceQuit {
			m.workspace.transition = workspaceIdle
		}
		if d := m.densitySettings; d != nil {
			d.busy, d.closing, d.err = false, false, r.err
		}
		return nil, true
	}
	p.saved = r.settings
	m.sidebarDensity = r.settings.SidebarDensity
	if r.settings.SidebarMode == p.wanted.SidebarMode && r.settings.DefaultGroupCollapsed == p.wanted.DefaultGroupCollapsed {
		m.applySidebarNavigation(r.settings)
	}
	m.sidebarPreferenceError = nil
	if strings.HasPrefix(m.status, "管理模式 · 侧栏偏好未保存：") || strings.HasPrefix(m.status, "界面设置未载入") {
		m.status = ""
	}
	if d := m.densitySettings; d != nil {
		d.busy, d.err = false, nil
		if sidebarDensities[d.choice] != m.sidebarDensity {
			p.wanted.SidebarDensity = sidebarDensities[d.choice]
			d.busy = true
		}
	}
	cmd := m.saveUIPreferences()
	if !p.busy && m.workspace.transition == workspaceQuit {
		m.workspace.transition = workspaceIdle
		return m.workspaceKey(tea.KeyPressMsg{Code: 'q', Text: "q"}), true
	}
	if d := m.densitySettings; d != nil && d.closing && !p.busy {
		return m.closeDensitySettings(), true
	}
	return cmd, true
}
func (m *Model) applySidebarNavigation(settings config.UISettings) {
	changed := m.sidebarMode != settings.SidebarMode || m.sidebarCollapsed != settings.DefaultGroupCollapsed
	m.sidebarMode, m.sidebarCollapsed = settings.SidebarMode, settings.DefaultGroupCollapsed
	if changed {
		m.viewOffset = 0
		m.workspace.listClick = workspaceListClick{}
	}
	if m.groupedSidebar() && (m.sidebarCollapsed || len(m.sessions) == 0) {
		m.cursor = -1
	} else if m.cursor < 0 && !m.groupedSidebar() {
		m.cursor = 0
	}
}
