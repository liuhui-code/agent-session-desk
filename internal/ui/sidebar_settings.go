// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/settings_panel.go and hotkeys.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: theme/density subset,
// v2 messages, token styling and asynchronous serialized preference saves.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package ui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/appearance"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type Options struct {
	LoadQuickLaunch func(context.Context) (session.Launch, error)
	RememberLaunch  func(context.Context, session.Launch) error
	LoadRecentDirs  func(context.Context) ([]string, error)
	RememberDirs    func(context.Context, []string) error
	// An injected provider must bind all resources to ctx, even if its returned
	// connection is never received because the UI has already exited.
	OpenSystemTheme        func(context.Context) (appearance.Connection, error)
	Theme                  string
	Footer                 string
	SidebarPercent         int
	SidebarDensity         string
	SidebarMode            string
	DefaultGroupCollapsed  bool
	CollapsedGroups        string
	SaveUISettings         func(context.Context, config.UISettings) error
	SaveSidebarDensity     func(context.Context, string) error
	SidebarPreferenceError error
}

func NewWithOptions(ctx context.Context, service Service, options Options) *Model {
	m := New(ctx, service)
	m.loadQuickLaunch, m.rememberLaunch = options.LoadQuickLaunch, options.RememberLaunch
	m.systemTheme.open = options.OpenSystemTheme
	if m.workspace != nil {
		m.workspace.directories.load = options.LoadRecentDirs
		m.workspace.directories.save = options.RememberDirs
	}
	if !config.ValidUITheme(options.Theme) {
		options.Theme = ""
	}
	m.applyUITheme(options.Theme)
	if !config.ValidSidebarPercent(options.SidebarPercent) {
		options.SidebarPercent = 0
	}
	if m.workspace != nil {
		m.workspace.split.override = options.SidebarPercent != 0
		m.workspace.split.percent = options.SidebarPercent
	}
	m.sidebarDensity = config.NormalizeSidebarDensity(options.SidebarDensity)
	m.sidebarMode = config.NormalizeSidebarMode(options.SidebarMode)
	m.sidebarCollapsed = options.DefaultGroupCollapsed
	m.saveUISettings = options.SaveUISettings
	if options.Footer != "" {
		options.Footer = config.NormalizeFooter(options.Footer)
	}
	m.preferences.saved = config.UISettings{Footer: options.Footer, Theme: options.Theme, SidebarPercent: options.SidebarPercent, SidebarDensity: m.sidebarDensity, SidebarMode: m.sidebarMode, DefaultGroupCollapsed: m.sidebarCollapsed, CollapsedGroups: options.CollapsedGroups}
	m.preferences.wanted = m.preferences.saved
	m.preferences.accepted = m.preferences.saved
	m.collapsedGroups = map[string]bool{}
	for _, id := range strings.Split(options.CollapsedGroups, ",") {
		if id != "" {
			m.collapsedGroups[id] = true
		}
	}
	if m.groupedSidebar() && m.sidebarCollapsed {
		m.cursor = -1
	}
	m.saveSidebarDensity = options.SaveSidebarDensity
	m.sidebarPreferenceError = options.SidebarPreferenceError
	if options.SidebarPreferenceError != nil {
		m.status = "UI preferences unavailable; using defaults: " + label(options.SidebarPreferenceError.Error())
	}
	return m
}

var sidebarDensities = [...]string{"full", "compact", "minimal", "auto"}
var uiThemes = [...]string{"dark", "light", "system"}

type densitySettings struct {
	choice           int
	themeChoice, row int
	busy, closing    bool
	err              error
}

func (m *Model) openDensitySettings() {
	d := &densitySettings{err: m.sidebarPreferenceError}
	for i, theme := range uiThemes {
		if effectiveUITheme(m.preferences.saved.Theme) == theme {
			d.themeChoice = i
		}
	}
	for i, value := range sidebarDensities {
		if value == m.sidebarDensity {
			d.choice = i
			break
		}
	}
	m.densitySettings = d
	m.workspace.focused = false
}
func (m *Model) closeDensitySettings() tea.Cmd {
	m.densitySettings = nil
	return m.connectWorkspace()
}
func (m *Model) saveDensityChoice() tea.Cmd {
	d := m.densitySettings
	if d.busy {
		return nil
	}
	if sidebarDensities[d.choice] == m.sidebarDensity && uiThemes[d.themeChoice] == effectiveUITheme(m.preferences.saved.Theme) {
		return nil
	}
	d.busy, d.err = true, nil
	m.preferences.wanted = m.settingsChoice()
	return m.saveUIPreferences()
}

func (m *Model) settingsChoice() config.UISettings {
	wanted, d := m.preferences.wanted, m.densitySettings
	wanted.SidebarDensity = sidebarDensities[d.choice]
	if effectiveUITheme(wanted.Theme) != uiThemes[d.themeChoice] {
		wanted.Theme = uiThemes[d.themeChoice]
	}
	return wanted
}
func (m *Model) densitySettingsEvent(msg tea.Msg) (tea.Cmd, bool) {
	d := m.densitySettings
	if d == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		return nil, true
	case tea.KeyPressMsg:
		if d.closing {
			return nil, true
		}
		switch msg.String() {
		case "r":
			return m.retrySystemTheme(), true
		case "esc", "S":
			if d.busy {
				d.closing = true
				return nil, true
			}
			return m.closeDensitySettings(), true
		case "left", "h":
			if d.row == 0 {
				d.themeChoice = max(0, d.themeChoice-1)
			} else {
				d.choice = max(0, d.choice-1)
			}
		case "right", "l":
			if d.row == 0 {
				d.themeChoice = min(len(uiThemes)-1, d.themeChoice+1)
			} else {
				d.choice = min(len(sidebarDensities)-1, d.choice+1)
			}
		case "space":
			if d.row == 0 {
				d.themeChoice = (d.themeChoice + 1) % len(uiThemes)
			} else {
				d.choice = (d.choice + 1) % len(sidebarDensities)
			}
		case "up", "k":
			d.row = max(0, d.row-1)
			return nil, true
		case "down", "j":
			d.row = min(1, d.row+1)
			return nil, true
		case "tab":
			d.row = (d.row + 1) % 2
			return nil, true
		default:
			return nil, true
		}
		return m.saveDensityChoice(), true
	case tea.KeyReleaseMsg, tea.PasteMsg, tea.MouseClickMsg, tea.MouseWheelMsg, tea.MouseMotionMsg, tea.MouseReleaseMsg:
		return nil, true
	}
	return nil, false
}
