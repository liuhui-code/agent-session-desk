// Layout adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/group_dialog.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: v2 total widths/real cursor, localized labels and bounded viewports.
// See third_party/agent-deck/LICENSE.
package ui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) sessionRenameWidths() (outer, content, input int) {
	t := m.tokens.rename
	outer = max(1, min(m.width, min(t.preferredWidth, max(t.minWidth, m.width-t.gutter))))
	content = max(0, outer-m.styles.renameBox.GetHorizontalFrameSize())
	input = min(t.inputWidth, content)
	return
}
func (m *Model) sizeSessionRename() {
	_, _, input := m.sessionRenameWidths()
	// v2 pads one cursor cell beyond Width; preserve the source's total field.
	sizeInput(&m.rename.input, input)
}
func (m *Model) renderSessionRename() (string, *tea.Cursor) {
	d := m.rename
	return m.renderNameDialog(d.input, "Rename session", d.busy, d.submitted, d.err)
}

func (m *Model) renderNameDialog(field textinput.Model, title string, busy, submitted bool, err error) (string, *tea.Cursor) {
	if m.width <= 0 || m.height <= 0 {
		return "", nil
	}
	t := m.tokens.rename
	outer, content, inputWidth := m.sessionRenameWidths()
	center := func(s string) string {
		s = fitFrame(s, min(content, lipgloss.Width(s)), 1)
		padding := max(0, content-lipgloss.Width(s))
		left := padding / 2
		// Child SGR resets otherwise expose the terminal background through
		// the plain trailing spaces produced by PlaceHorizontal.
		return m.styles.formText.Render(strings.Repeat(" ", left)) + s + m.styles.formText.Render(strings.Repeat(" ", padding-left))
	}
	input := center(fitFrame(inputView(field), inputWidth, 1))
	errLine := ""
	if err != nil {
		errLine = m.styles.formError.Render(t.warningPrefix + label(err.Error()))
	}
	hint := "Enter Confirm | Esc Cancel"
	if busy {
		hint = "Saving name…"
	} else if submitted {
		hint = "Esc Back · Ctrl+R Refresh before editing"
	}
	lines := []string{center(m.styles.renameTitle.Render(title)), "", input, center(errLine), "", center(m.styles.renameHint.Render(hint))}
	inputRow := 2 // source structure: title, blank, input
	maxRows := max(0, m.height-m.styles.renameBox.GetVerticalFrameSize())
	if maxRows < len(lines) {
		lines = []string{input, center(errLine), center(m.styles.renameHint.Render(hint))}
		inputRow = 0
	}
	box := m.styles.renameBox.Width(outer).Render(fitFrame(strings.Join(lines, "\n"), content, min(maxRows, len(lines))))
	boxWidth, boxHeight := lipgloss.Size(box)
	x, y := max(0, (m.width-boxWidth)/2), max(0, (m.height-boxHeight)/2)
	view := overlayLaunchDialog(fitFrame("", m.width, m.height), box, x, y, m.width, m.height)
	cursor := inputCursor(field)
	if cursor != nil {
		cursor.X += x + m.styles.renameBox.GetBorderLeftSize() + t.paddingX + max(0, (content-inputWidth)/2)
		cursor.Y += y + m.styles.renameBox.GetBorderTopSize() + t.paddingY + inputRow
		if busy || submitted || cursor.X >= m.width || cursor.Y >= m.height || cursor.Y >= y+boxHeight-m.styles.renameBox.GetBorderBottomSize()-t.paddingY || cursor.X >= x+boxWidth-m.styles.renameBox.GetBorderRightSize()-t.paddingX {
			cursor = nil
		}
	}
	return view, cursor
}
