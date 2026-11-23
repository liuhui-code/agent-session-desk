// Layout adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/search.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: stable v2 total widths/real cursor, token styling, scoped terminal
// metadata and bounded short windows. See third_party/agent-deck/LICENSE.
package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) sessionSearchWidths() (outer, content, input int) {
	t := m.tokens.search
	outer = t.preferredWidth
	if m.width < t.responsiveBelow {
		outer = max(t.minWidth, m.width-t.gutter)
	}
	outer = max(1, min(m.width, outer))
	content = max(0, outer-m.styles.searchBox.GetHorizontalFrameSize())
	input = max(1, content-t.inputRightGap)
	return
}

func (m *Model) sizeSessionSearch() {
	_, _, input := m.sessionSearchWidths()
	m.search.input.SetWidth(max(1, input-m.styles.searchInput.GetHorizontalFrameSize()-lipgloss.Width(m.search.input.Prompt)))
}

func (m *Model) renderSessionSearch() (string, *tea.Cursor) {
	if m.width <= 0 || m.height <= 0 {
		return "", nil
	}
	d, t := m.search, m.tokens.search
	width, contentWidth, inputWidth := m.sessionSearchWidths()
	input := strings.Split(m.styles.searchInput.Width(inputWidth).Render(d.input.View()), "\n")
	head := []string{m.styles.searchTitle.Render(t.titlePrefix + "本地会话搜索"), ""}
	inputStart := len(head)
	head = append(head, input...)
	if d.notice != "" {
		head = append(head, m.styles.error.Render(label(d.notice)))
	} else if d.input.Value() == "" {
		head = append(head, m.styles.searchTip.Render("按名称或程序搜索"))
	}
	head = append(head, "")
	count := fmt.Sprintf("%d 个结果", len(d.results))
	if len(d.results) == 0 {
		count = "无匹配会话"
	}
	help := "Enter 进入 · ↑↓ 导航 · Esc 取消"
	if lipgloss.Width(help) > contentWidth {
		help = "↵ · ↑↓ · Esc"
	}
	foot := []string{m.styles.searchCount.Render(count), m.styles.searchHint.Render(help)}
	maxRows := max(0, m.height-m.styles.searchBox.GetVerticalFrameSize())
	if maxRows < len(head)+len(foot)+1 {
		head, inputStart = input, 0
		if maxRows < len(head)+len(foot)+1 {
			foot = nil
		}
	}
	visible := min(t.maxResults, max(0, maxRows-len(head)-len(foot)))
	start := max(0, d.cursor-visible+1)
	end := min(len(d.results), start+visible)
	lines := append([]string(nil), head...)
	for i := start; i < end; i++ {
		s := d.results[i]
		programInfo := ""
		if program := label(s.Command); program != "" {
			programInfo = t.programPrefix + program[strings.LastIndexAny(program, `/\`)+1:] + t.programSuffix
		}
		prefix := strings.Repeat(" ", lipgloss.Width(t.selected))
		style := m.styles.searchResult
		if i == d.cursor {
			prefix, style = t.selected, m.styles.searchSelected
		}
		rowWidth := max(0, contentWidth-style.GetHorizontalFrameSize())
		nameWidth := max(0, rowWidth-lipgloss.Width(prefix))
		if lipgloss.Width(programInfo) >= nameWidth {
			programInfo = ""
		}
		name := m.sessionTitle(s, m.sessions, max(0, nameWidth-lipgloss.Width(programInfo))) + programInfo
		row := prefix + name
		lines = append(lines, style.Render(fitFrame(row, rowWidth, 1)))
	}
	lines = append(lines, foot...)
	box := m.styles.searchBox.Width(width).Render(fitFrame(strings.Join(lines, "\n"), contentWidth, min(maxRows, len(lines))))
	boxWidth, boxHeight := lipgloss.Size(box)
	x, y := max(0, (m.width-boxWidth)/2), max(0, (m.height-boxHeight)/2)
	view := overlayLaunchDialog(fitFrame("", m.width, m.height), box, x, y, m.width, m.height)
	cursor := d.input.Cursor()
	if cursor != nil {
		cursor.X += x + m.styles.searchBox.GetBorderLeftSize() + t.paddingX + m.styles.searchInput.GetBorderLeftSize() + t.inputPaddingX
		cursor.Y += y + m.styles.searchBox.GetBorderTopSize() + t.paddingY + inputStart + m.styles.searchInput.GetBorderTopSize() + t.inputPaddingY
		if cursor.X >= m.width || cursor.Y >= m.height || cursor.Y >= y+boxHeight-m.styles.searchBox.GetBorderBottomSize() {
			cursor = nil
		}
	}
	return view, cursor
}
