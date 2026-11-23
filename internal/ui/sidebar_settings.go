// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/settings_panel.go and hotkeys.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: only sidebar density,
// v2 messages, token styling and asynchronous serialized preference saves.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package ui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/liuhui-code/agent-session-desk/internal/config"
)

type Options struct {
	SidebarDensity         string
	SidebarMode            string
	DefaultGroupCollapsed  bool
	SaveUISettings         func(context.Context, config.UISettings) error
	SaveSidebarDensity     func(context.Context, string) error
	SidebarPreferenceError error
}

func NewWithOptions(ctx context.Context, service Service, options Options) *Model {
	m := New(ctx, service)
	m.sidebarDensity = config.NormalizeSidebarDensity(options.SidebarDensity)
	m.sidebarMode = config.NormalizeSidebarMode(options.SidebarMode)
	m.sidebarCollapsed = options.DefaultGroupCollapsed
	m.saveUISettings = options.SaveUISettings
	m.preferences.saved = config.UISettings{SidebarDensity: m.sidebarDensity, SidebarMode: m.sidebarMode, DefaultGroupCollapsed: m.sidebarCollapsed}
	m.preferences.wanted = m.preferences.saved
	if m.groupedSidebar() && m.sidebarCollapsed {
		m.cursor = -1
	}
	m.saveSidebarDensity = options.SaveSidebarDensity
	m.sidebarPreferenceError = options.SidebarPreferenceError
	if options.SidebarPreferenceError != nil {
		m.status = "界面设置未载入，使用默认值：" + label(options.SidebarPreferenceError.Error())
	}
	return m
}

var sidebarDensities = [...]string{"full", "compact", "minimal", "auto"}

type densitySettings struct {
	choice        int
	busy, closing bool
	err           error
}

func (m *Model) openDensitySettings() {
	d := &densitySettings{err: m.sidebarPreferenceError}
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
	value := sidebarDensities[d.choice]
	if value == m.sidebarDensity {
		return nil
	}
	d.busy, d.err = true, nil
	m.preferences.wanted.SidebarDensity = value
	return m.saveUIPreferences()
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
		case "esc", "S":
			if d.busy {
				d.closing = true
				return nil, true
			}
			return m.closeDensitySettings(), true
		case "left", "h":
			d.choice = max(0, d.choice-1)
		case "right", "l":
			d.choice = min(len(sidebarDensities)-1, d.choice+1)
		case "space":
			d.choice = (d.choice + 1) % len(sidebarDensities)
		default:
			return nil, true
		}
		return m.saveDensityChoice(), true
	case tea.KeyReleaseMsg, tea.PasteMsg, tea.MouseClickMsg, tea.MouseWheelMsg, tea.MouseMotionMsg, tea.MouseReleaseMsg:
		return nil, true
	}
	return nil, false
}

func (m *Model) renderDensitySettings() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	d, t := m.densitySettings, m.tokens
	width := max(1, min(t.sidebar.settingsWidth, m.width))
	contentWidth := max(1, width-m.styles.settingsBox.GetHorizontalFrameSize())
	var choices []string
	names := []string{"Full", "Compact", "Minimal", "Auto"}
	for i, name := range names {
		style, prefix := m.styles.settingsRadio, t.sidebar.radioUnselected
		if i == d.choice {
			style, prefix = m.styles.settingsChoice, t.sidebar.radioMarker
		}
		choices = append(choices, style.Render(prefix+name))
	}
	radio := strings.Join(choices, m.styles.settingsFocused.Render(strings.Repeat(" ", t.sidebar.radioGap)))
	labelText := m.styles.settingsFocused.Render("侧栏密度: ")
	if lipgloss.Width(labelText)+lipgloss.Width(radio) > contentWidth {
		radio = m.styles.settingsChoice.Render(t.sidebar.radioMarker + names[d.choice])
	}
	row := labelText + radio
	row += m.styles.settingsFocused.Render(strings.Repeat(" ", max(0, contentWidth-lipgloss.Width(row))))
	lines := []string{m.styles.settingsTitle.Render("设置")}
	for range t.dialog.titleGap {
		lines = append(lines, "")
	}
	lines = append(lines, row, m.styles.settingsHint.Render("每会话 3 / 2 / 1 行；Auto 按空间选择"), m.styles.settingsHint.Render("←→ / h/l 选择 · 空格轮换 · Esc/S 返回"))
	if d.busy {
		lines = append(lines, m.styles.settingsHint.Render("正在保存…"))
	}
	if d.err != nil {
		lines = append(lines, m.styles.settingsError.Render("未保存："+label(d.err.Error())))
	}
	maxRows := max(1, m.height-m.styles.settingsBox.GetVerticalFrameSize())
	if len(lines) > maxRows {
		lines = []string{row}
		if d.err != nil {
			lines = append(lines, m.styles.settingsError.Render("未保存："+label(d.err.Error())))
		}
		if d.busy {
			lines = append(lines, m.styles.settingsHint.Render("正在保存…"))
		}
		lines = append(lines, m.styles.settingsHint.Render("←→ 选择 · Esc 返回"))
	}
	for i, line := range lines {
		lines[i] = line + m.styles.settingsText.Render(strings.Repeat(" ", max(0, contentWidth-lipgloss.Width(line))))
	}
	content := fitFrame(strings.Join(lines, "\n"), contentWidth, min(len(lines), maxRows))
	box := m.styles.settingsBox.Width(width).Render(content)
	bw, bh := lipgloss.Size(box)
	x, y := max(0, (m.width-bw)/2), max(0, (m.height-bh)/2)
	return overlayLaunchDialog(fitFrame("", m.width, m.height), box, x, y, m.width, m.height)
}
