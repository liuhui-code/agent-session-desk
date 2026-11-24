// Layout adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/newdialog.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: v2 total-size semantics, generic CLI fields, safe cell slicing and
// focus-aware short viewport. See third_party/agent-deck/LICENSE.
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) sizeLaunchDialog() {
	if m.launch == nil {
		return
	}
	width := m.launchWidth() - m.styles.dialog.GetHorizontalFrameSize()
	for i := range m.launch.fields {
		sizeInput(&m.launch.fields[i], width)
	}
}
func (m *Model) launchWidth() int {
	t := m.tokens.dialog
	available := m.width
	if m.width >= t.minWidth+t.gutter {
		available -= t.gutter
	}
	return max(1, min(t.preferredWidth, available))
}

func (m *Model) renderLaunchDialog() (string, *tea.Cursor) {
	d := m.launch
	t := m.tokens.dialog
	paddingY := t.paddingY
	if d.firstSession {
		paddingY = t.firstPaddingY
	}
	dialogStyle := m.styles.dialog.Padding(paddingY, t.paddingX)
	width := m.launchWidth()
	contentWidth := max(1, width-dialogStyle.GetHorizontalFrameSize())
	var lines []string
	lines = append(lines, m.styles.formTitle.Render("新建终端"))
	for range t.titleGap {
		lines = append(lines, "")
	}
	headerRows := len(lines)
	labels := []string{"名称（可空）", "程序 / CLI", "工作目录（可空）", "参数（JSON 字符串数组，可空）", "环境覆盖（JSON KEY=value 数组，可空）"}
	cursorRow := -1
	for i := range d.fields {
		style := m.styles.formText
		if i == d.focus {
			style = m.styles.formTitle
		}
		lines = append(lines, style.Render(labels[i]))
		if i == d.focus {
			cursorRow = len(lines)
		}
		lines = append(lines, inputView(d.fields[i]))
		for range t.fieldGap {
			lines = append(lines, "")
		}
	}
	button := m.styles.tab.Render("创建终端")
	if d.focus == len(d.fields) {
		button = m.styles.tabActive.Render("创建终端")
	}
	if d.busy {
		button = m.styles.formHint.Render("正在启动；请等待明确结果…")
	}
	footer := []string{button, m.styles.formHint.Render("Tab 切换 · Ctrl+S 创建 · Esc 取消")}
	if d.err != nil {
		footer = append(footer, m.styles.formError.Render(cellTruncate(label(d.err.Error()), contentWidth, m.tokens.marks.ellipsis)))
	}
	maxRows := max(1, m.height-dialogStyle.GetVerticalFrameSize())
	if len(lines)+len(footer) > maxRows {
		budget := max(1, maxRows-headerRows-len(footer))
		start := headerRows
		if cursorRow >= start+budget {
			start = cursorRow - budget + 1
		}
		end := min(len(lines), start+budget)
		cropped := append(append([]string(nil), lines[:min(headerRows, len(lines))]...), lines[start:end]...)
		if cursorRow >= start && cursorRow < end {
			cursorRow = headerRows + cursorRow - start
		} else {
			cursorRow = -1
		}
		lines = cropped
	}
	lines = append(lines, footer...)
	// Child styles terminate with an SGR reset. Paint the trailing cells too,
	// so padding cannot inherit the outer terminal background through it.
	for i, line := range lines {
		padding := strings.Repeat(" ", max(0, contentWidth-lipgloss.Width(line)))
		lines[i] = line + m.styles.formText.Render(padding)
	}
	content := fitFrame(strings.Join(lines, "\n"), contentWidth, min(maxRows, len(lines)))
	box := dialogStyle.Width(width).Render(content)
	boxWidth, boxHeight := lipgloss.Size(box)
	x, y := max(0, (m.width-boxWidth)/2), max(0, (m.height-boxHeight)/2)
	background := m.styles.backdrop.Render(m.renderWorkspace())
	view := overlayLaunchDialog(background, box, x, y, m.width, m.height)
	var cursor *tea.Cursor
	if !d.busy && d.focus < len(d.fields) && cursorRow >= 0 {
		cursor = inputCursor(d.fields[d.focus])
		if cursor != nil {
			cursor.X += x + dialogStyle.GetBorderLeftSize() + t.paddingX
			cursor.Y += y + dialogStyle.GetBorderTopSize() + paddingY + cursorRow
			if cursor.X >= m.width || cursor.Y >= m.height {
				cursor = nil
			}
		}
	}
	if d.focus == 2 && cursorRow >= 0 && !d.busy {
		view = m.overlayPathPicker(view, d.paths, x+dialogStyle.GetBorderLeftSize()+t.paddingX, y+dialogStyle.GetBorderTopSize()+paddingY+cursorRow)
	}
	return view, cursor
}

func overlayLaunchDialog(base, box string, x, y, width, height int) string {
	rows := strings.Split(base, "\n")
	for i, line := range strings.Split(box, "\n") {
		target := y + i
		if target < 0 || target >= len(rows) {
			continue
		}
		rows[target] = ensureExactWidth(ansi.Cut(rows[target], 0, x), x) + line + ansi.Cut(rows[target], x+lipgloss.Width(line), width)
	}
	return fitFrame(strings.Join(rows, "\n"), width, height)
}
