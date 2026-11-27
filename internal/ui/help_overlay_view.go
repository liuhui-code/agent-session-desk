// Help layout adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/help.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: v2 total widths, token styles, localized real-cell rows and bounded
// short windows. No upstream agent/account/remote shortcuts are advertised.
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) helpWidth() int {
	t := m.tokens.help
	width := t.shortWidth
	if m.helpPage.all {
		width = t.fullWidth
		if m.width >= t.wideAt {
			width = t.wideWidth
		}
	}
	return max(1, min(m.width, max(t.minWidth, min(width, m.width-t.gutter))))
}

func (m *Model) helpRows(key, desc string, width int, short bool) []string {
	t := m.tokens.help
	indent, keyWidth := t.rowIndent, t.keyWidth
	if m.helpWidth() < t.compactBelow {
		keyWidth = t.compactKeyWidth
	}
	if short {
		indent, keyWidth = t.shortIndent, t.shortKeyWidth
	}
	keyWidth = max(1, min(keyWidth, (width-indent-t.columnGap)/2))
	descWidth := max(1, width-indent-keyWidth-t.columnGap)
	keys := strings.Split(ansi.Wrap(key, keyWidth, ""), "\n")
	descriptions := strings.Split(ansi.Wrap(desc, descWidth, ""), "\n")
	rows := make([]string, max(len(keys), len(descriptions)))
	for i := range rows {
		k, d := "", ""
		if i < len(keys) {
			k = keys[i]
		}
		if i < len(descriptions) {
			d = descriptions[i]
		}
		rows[i] = m.styles.helpText.Render(strings.Repeat(" ", indent)) + m.styles.helpKey.Width(keyWidth).Render(k) + m.styles.helpText.Render(strings.Repeat(" ", t.columnGap)) + m.styles.helpText.Render(d)
	}
	return rows
}

func (m *Model) allHelpRows(width int) []string {
	rows := []string{m.styles.helpTitle.Render("All shortcuts"), ""}
	for i, section := range m.configuredHelp() {
		if i > 0 {
			rows = append(rows, make([]string, m.tokens.help.sectionGap)...)
		}
		rows = append(rows, strings.Split(m.styles.helpSection.Width(width).Render(section.title), "\n")...)
		for _, item := range section.items {
			rows = append(rows, m.helpRows(item[0], item[1], width, false)...)
		}
	}
	rows = append(rows, "", m.styles.helpHint.Render("Agent Session Desk"), m.styles.helpNote.Render("Closing a tab does not stop its run"))
	return rows
}

func (m *Model) helpBodyBudget(footerRows int) int {
	t := m.tokens.help
	return min(max(0, m.height-m.styles.helpBox.GetVerticalFrameSize()-footerRows-t.footerGap), max(t.minContentRows, m.height-t.screenReserve))
}

// Row counting and navigation share the same wrapped content and footer budget.
// Both indicator slots stay reserved in the scroll budget. Like the source,
// the final page omits the unused bottom indicator row from the actual box.
func (m *Model) fullHelpPage() (lines, footer []string, capacity, limit int) {
	width := max(1, m.helpWidth()-m.styles.helpBox.GetHorizontalFrameSize())
	lines = m.allHelpRows(width)
	footer = strings.Split(m.styles.helpHint.Width(width).Render("j/k Scroll · Tab Quick start · Any other key Close"), "\n")
	capacity = m.helpBodyBudget(len(footer))
	if len(lines) > capacity {
		capacity = max(1, capacity-2) // Fixed top/bottom indicator rows.
	}
	limit = max(0, len(lines)-capacity)
	return
}

func (m *Model) renderHelpOverlay() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	outer, t := m.helpWidth(), m.tokens.help
	width := max(1, outer-m.styles.helpBox.GetHorizontalFrameSize())
	var rows []string
	if !m.helpPage.all {
		rows = []string{m.styles.helpTitle.Render("Quick start"), ""}
		for _, entry := range [][2]string{{"Enter", "Open selected terminal"}, {"F6", "Return to navigation/terminal focus"}, {"n", "New terminal"}, {"↑ ↓", "Select session"}, {"?", "Open help"}} {
			rows = append(rows, m.helpRows(entry[0], entry[1], width, true)...)
		}
		rows = append(rows, "", m.styles.helpNote.Width(width).Render("Type directly in the terminal; the program owns its draft."), "", m.styles.helpTab.Render("Tab All shortcuts")+m.styles.helpNote.Render(t.separator+"Esc Close"))
	} else {
		lines, footer, capacity, limit := m.fullHelpPage()
		offset := max(0, min(m.helpPage.offset, limit))
		rows = lines[offset:min(len(lines), offset+capacity)]
		if limit > 0 {
			top, bottom := "", ""
			if offset > 0 {
				top = m.styles.helpScroll.Render(t.above)
			}
			if offset < limit {
				bottom = m.styles.helpScroll.Render(t.below)
			}
			rows = append([]string{top}, rows...)
			if offset < limit {
				rows = append(rows, bottom)
			}
		}
		rows = append(rows, make([]string, t.footerGap)...)
		rows = append(rows, footer...)
	}
	maxRows := max(0, m.height-m.styles.helpBox.GetVerticalFrameSize())
	lines := strings.Split(strings.Join(rows, "\n"), "\n")
	// Inner styles reset SGR. Paint every gap and trailing cell explicitly,
	// preserving the source surface rather than the outer terminal background.
	for i, line := range lines {
		line = cellTruncate(line, width, "")
		lines[i] = line + m.styles.helpText.Render(strings.Repeat(" ", max(0, width-lipgloss.Width(line))))
	}
	box := m.styles.helpBox.Width(outer).Render(fitFrame(strings.Join(lines, "\n"), width, min(maxRows, len(lines))))
	bw, bh := lipgloss.Size(box)
	return overlayLaunchDialog(fitFrame("", m.width, m.height), box, max(0, (m.width-bw)/2), max(0, (m.height-bh)/2), m.width, m.height)
}
