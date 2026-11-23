// Adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: group_dialog.go and newdialog.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0; token styles and bounded overlay.
package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) pathPopupAvailable() bool {
	return m.width >= m.tokens.paths.minWidth && m.height >= m.tokens.paths.minHeight
}
func (m *Model) overlayPathPicker(view string, p *pathPicker, x, inputY int) string {
	if p == nil || p.hidden || !m.pathPopupAvailable() || m.focusedPathPicker() != p || (!p.launch && len(p.matches) == 0) {
		return view
	}
	t := m.tokens.paths
	s := m.styles
	boxStyle := s.pathBox
	if p.active {
		boxStyle = s.pathBrowseBox
	}
	width := min(t.width, m.width)
	content := max(1, width-boxStyle.GetHorizontalFrameSize())
	rows := t.rows
	if m.height <= t.shortAt {
		rows = t.shortRows
	}
	available := max(inputY, m.height-inputY-1) - boxStyle.GetVerticalFrameSize()
	fixed := 1
	if p.launch {
		fixed++
	}
	rows = max(1, min(rows, available-fixed))
	showCounts := available-fixed-rows >= 2
	realCursor := p.selected
	if p.launch {
		realCursor--
	}
	start := max(0, min(realCursor-rows/2, len(p.matches)-rows))
	end := min(len(p.matches), start+rows)
	var lines []string
	render := func(text string, style lipgloss.Style) string {
		text = cellTruncate(label(text), content, m.tokens.marks.ellipsis)
		return style.Render(text + strings.Repeat(" ", max(0, content-lipgloss.Width(text))))
	}
	if p.launch {
		prefix, style := t.idle, s.pathCustom
		if p.selected == 0 {
			prefix, style = t.selected, s.pathCustomSelected
		}
		lines = append(lines, render(prefix+t.custom, style))
	}
	if start > 0 && showCounts {
		lines = append(lines, render(fmt.Sprintf("%s%s%d", t.idle, t.above, start), s.pathText))
	}
	for i := start; i < end; i++ {
		prefix, style := t.idle, s.pathText
		if i == realCursor {
			prefix, style = t.selected, s.pathSelected
		}
		lines = append(lines, render(prefix+p.matches[i], style))
	}
	if end < len(p.matches) && showCounts {
		lines = append(lines, render(fmt.Sprintf("%s%s%d", t.idle, t.below, len(p.matches)-end), s.pathText))
	}
	hint := "↑↓ 选择 │ Tab/Enter 采纳 │ Esc 隐藏"
	if !p.launch && len(p.matches) < len(p.corpus) {
		hint = fmt.Sprintf("%d/%d 匹配 │ %s", len(p.matches), len(p.corpus), hint)
	}
	if p.launch {
		hint = "→/Space 浏览"
		if p.active {
			hint = "↑↓ 浏览 │ Space 采纳 │ Enter 下一项"
		} else if len(p.cycle) > 0 {
			hint = "Tab 轮换目录 │ Enter 下一项"
		}
	}
	lines = append(lines, render(hint, s.pathHint))
	box := boxStyle.Width(width).Render(strings.Join(lines, "\n"))
	bw, bh := lipgloss.Size(box)
	x = max(0, min(x, m.width-bw))
	y := inputY + 1
	if y+bh > m.height {
		y = max(0, inputY-bh)
	}
	// Above/below placement may cover other form rows, never the focused input.
	if y <= inputY && y+bh > inputY {
		return view
	}
	return overlayLaunchDialog(view, box, x, y, m.width, m.height)
}
