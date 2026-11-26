// Settings row/radio treatment adapted from Agent Deck, Copyright
// (c) 2025 Ashesh Goplani. SPDX-License-Identifier: MIT
// Source: settings_panel.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: bounded theme/density subset, v2 tokens and short-window focus row.
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) settingsRow(row, width int, compact bool) string {
	d, t := m.densitySettings, m.tokens
	names, selected, labelText := []string{"Dark", "Light", "System"}, d.themeChoice, ""
	if compact {
		labelText = "Theme: "
	}
	if row == 1 {
		names, selected, labelText = []string{"Full", "Compact", "Minimal", "Auto"}, d.choice, "Sidebar density: "
		if compact {
			labelText = "Density: "
		}
	}
	base, radio, choice, prefix := m.styles.settingsText, m.styles.settingsRadioIdle, m.styles.settingsChoiceIdle, t.sidebar.settingsIdlePrefix
	marker := base
	if d.row == row {
		base, radio, choice, prefix = m.styles.settingsFocused, m.styles.settingsRadio, m.styles.settingsChoice, t.sidebar.settingsFocusedPrefix
		marker = m.styles.settingsFocusMarker
	}
	var parts []string
	for i, name := range names {
		style, mark := radio, t.sidebar.radioUnselected
		if i == selected {
			style, mark = choice, t.sidebar.radioMarker
		}
		parts = append(parts, style.Render(mark+name))
	}
	radioText := strings.Join(parts, base.Render(strings.Repeat(" ", t.sidebar.radioGap)))
	if lipgloss.Width(prefix+labelText)+lipgloss.Width(radioText) > width {
		radioText = choice.Render(t.sidebar.radioMarker + names[selected])
		if lipgloss.Width(prefix+labelText)+lipgloss.Width(radioText) > width {
			labelText = ""
		}
	}
	line := marker.Render(prefix) + base.Render(labelText) + radioText
	return line + base.Render(strings.Repeat(" ", max(0, width-lipgloss.Width(line))))
}

func (m *Model) renderDensitySettings() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	d, t := m.densitySettings, m.tokens
	width := max(1, min(t.sidebar.settingsWidth, m.width))
	contentWidth := max(1, width-m.styles.settingsBox.GetHorizontalFrameSize())
	lines := []string{m.styles.settingsTitle.Render("Settings")}
	for range t.dialog.titleGap {
		lines = append(lines, "")
	}
	lines = append(lines, m.styles.settingsSection.Render("Theme"), m.settingsRow(0, contentWidth, false))
	if caption := m.systemThemeCaption(); caption != "" {
		lines = append(lines, m.styles.settingsHint.Render(caption))
	}
	for range t.dialog.fieldGap {
		lines = append(lines, "")
	}
	lines = append(lines, m.styles.settingsSection.Render("Layout"), m.settingsRow(1, contentWidth, false), m.styles.settingsHint.Render("3 / 2 / 1 content rows per session; Auto fits available space"), m.styles.settingsHint.Render("↑↓ / j/k / Tab Select row · ←→ / h/l Choose"), m.styles.settingsHint.Render("Space Cycle · Apply after saving · Esc/S Back"))
	if d.busy {
		lines = append(lines, m.styles.settingsHint.Render("Saving…"))
	}
	if d.err != nil {
		lines = append(lines, m.styles.settingsError.Render("Not saved: "+label(d.err.Error())))
	}
	maxRows := max(1, m.height-m.styles.settingsBox.GetVerticalFrameSize())
	if len(lines) > maxRows {
		lines = []string{m.settingsRow(d.row, contentWidth, true)}
		if d.err != nil {
			lines = append(lines, m.styles.settingsError.Render("Not saved: "+label(d.err.Error())))
		}
		if caption := m.systemThemeCaption(); caption != "" {
			lines = append(lines, m.styles.settingsHint.Render(caption))
		}
		if d.busy {
			lines = append(lines, m.styles.settingsHint.Render("Saving…"))
		}
		lines = append(lines, m.styles.settingsHint.Render("↑↓ Select row · ←→ Choose · Esc Back"))
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
