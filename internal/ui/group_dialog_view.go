// Layout adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/group_dialog.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: v2 total widths/real cursor, stable tree labels and bounded viewport.
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) groupDialogWidths() (outer, content, field int) {
	outer, content, _ = m.sessionRenameWidths()
	field = max(1, min(m.tokens.group.inputWidth, content-m.tokens.group.labelWidth))
	return
}
func (m *Model) sizeGroupDialog() {
	d := m.groupDialog
	if d == nil || d.move {
		return
	}
	if d.rename {
		_, _, width := m.sessionRenameWidths()
		sizeInput(&d.inputs[0], width)
		return
	}
	_, _, width := m.groupDialogWidths()
	for i := range d.inputs {
		sizeInput(&d.inputs[i], width)
	}
}
func (m *Model) groupPath(id string) string {
	var parts []string
	for depth := 0; id != "" && depth < session.MaxGroupDepth; depth++ {
		g := m.groupByID(id)
		if g.ID == "" {
			break
		}
		parts = append([]string{label(g.Name)}, parts...)
		id = g.ParentID
	}
	return strings.Join(parts, "/")
}
func (m *Model) renderGroupDialog() (string, *tea.Cursor) {
	if m.width <= 0 || m.height <= 0 {
		return "", nil
	}
	d, t := m.groupDialog, m.tokens
	if d.rename {
		return m.renderNameDialog(d.inputs[0], "Rename group", d.busy, d.submitted, d.err)
	}
	outer, content, fieldWidth := m.groupDialogWidths()
	center := func(s string) string {
		s = fitFrame(s, min(content, lipgloss.Width(s)), 1)
		padding := max(0, content-lipgloss.Width(s))
		left := padding / 2
		return m.styles.formText.Render(strings.Repeat(" ", left)) + s + m.styles.formText.Render(strings.Repeat(" ", padding-left))
	}
	title := "Create group"
	if d.move {
		title = "Move to group"
	} else if d.parentID != "" {
		title = "Create subgroup"
	}
	lines := []string{center(m.styles.renameTitle.Render(title)), ""}
	inputRow := -1
	blockWidth := min(content, t.group.labelWidth+fieldWidth)
	if !d.move {
		if d.contextID != "" {
			root, sub := m.styles.groupLabel.Render("Root group"), m.styles.groupLabel.Render("Subgroup")
			if d.parentID == "" {
				root = m.styles.groupMode.Render(t.group.modeOpen + "Root group" + t.group.modeClose)
			} else {
				sub = m.styles.groupMode.Render(t.group.modeOpen + "Subgroup" + t.group.modeClose)
			}
			lines = append(lines, center(root+m.styles.groupLabel.Render(t.group.modeSeparator)+sub), "")
		}
		if d.parentID != "" {
			lines = append(lines, center(m.styles.groupContext.Render("Parent: "+m.groupPath(d.parentID))), "")
		}
		inputRow = len(lines) + d.focus
		for i, name := range []string{"Name:", "Directory:"} {
			labelText := fitFrame(name, t.group.labelWidth, 1)
			lines = append(lines, center(m.styles.groupLabel.Render(labelText)+fitFrame(inputView(d.inputs[i]), fieldWidth, 1)))
		}
	} else {
		budget := max(1, m.height-m.styles.renameBox.GetVerticalFrameSize()-6) // title/blank/error/blank/hint plus one clipping indicator
		start := max(0, min(d.choice-budget+1, len(d.groups)-budget))
		end := min(len(d.groups), start+budget)
		if start > 0 {
			lines = append(lines, center(m.styles.renameHint.Render(t.marks.above)))
		}
		for i := start; i < end; i++ {
			style := m.styles.groupChoice
			if i == d.choice {
				style = m.styles.groupChoiceSelected
			}
			path := cellTruncate(m.groupPath(d.groups[i].ID), max(0, content-style.GetHorizontalFrameSize()), t.marks.ellipsis)
			lines = append(lines, center(style.Render(path)))
		}
	}
	errLine := ""
	if d.err != nil {
		errLine = m.styles.formError.Render(t.rename.warningPrefix + label(d.err.Error()))
	}
	hint := "Tab Next | Enter Confirm | Esc Cancel"
	if d.move {
		hint = "↑↓ Select | Enter Confirm | Esc Cancel"
	}
	if lipgloss.Width(hint) > content {
		hint = "Enter Confirm | Esc Cancel"
	}
	if d.busy {
		hint = "Saving group…"
	} else if d.submitted {
		hint = "Esc Back · Refresh to verify"
	}
	lines = append(lines, center(errLine), "", center(m.styles.renameHint.Render(hint)))
	maxRows := max(0, m.height-m.styles.renameBox.GetVerticalFrameSize())
	if len(lines) > maxRows && !d.move {
		// Keep both fields when possible; the focused one is first at tiny heights.
		fields := append([]string(nil), lines[inputRow-d.focus:inputRow-d.focus+2]...)
		if maxRows < 2 {
			fields = []string{fields[d.focus]}
			inputRow = 0
		} else {
			inputRow = d.focus
		}
		lines = append(fields, center(errLine), center(m.styles.renameHint.Render(hint)))
	}
	box := m.styles.renameBox.Width(outer).Render(fitFrame(strings.Join(lines, "\n"), content, min(maxRows, len(lines))))
	bw, bh := lipgloss.Size(box)
	x, y := max(0, (m.width-bw)/2), max(0, (m.height-bh)/2)
	view := overlayLaunchDialog(fitFrame("", m.width, m.height), box, x, y, m.width, m.height)
	var cursor *tea.Cursor
	if !d.move && !d.busy && !d.submitted && inputRow >= 0 && inputRow < maxRows {
		cursor = inputCursor(d.inputs[d.focus])
		if cursor != nil {
			cursor.X += x + m.styles.renameBox.GetBorderLeftSize() + t.rename.paddingX + max(0, (content-blockWidth)/2) + t.group.labelWidth
			cursor.Y += y + m.styles.renameBox.GetBorderTopSize() + t.rename.paddingY + inputRow
			if cursor.X >= m.width || cursor.Y >= m.height || cursor.X >= x+bw-m.styles.renameBox.GetBorderRightSize()-t.rename.paddingX || cursor.Y >= y+bh-m.styles.renameBox.GetBorderBottomSize()-t.rename.paddingY {
				cursor = nil
			}
		}
	}
	if !d.move && d.focus == 1 && inputRow >= 0 && !d.busy && !d.submitted {
		view = m.overlayPathPicker(view, d.paths, x+m.styles.renameBox.GetBorderLeftSize()+t.rename.paddingX+max(0, (content-blockWidth)/2)+t.group.labelWidth, y+m.styles.renameBox.GetBorderTopSize()+t.rename.paddingY+inputRow)
	}
	return view, cursor
}
