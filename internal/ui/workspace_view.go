// Layout portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Sources: internal/ui/home_view_frame.go and split_config.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: product Tabs, one native
// psmux screen, v2 size/cursor handling and token-only geometry.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type workspaceGeometry struct{ x, y, width, height, left, outerWidth, bodyHeight int }

func (m *Model) workspaceGeometry() workspaceGeometry {
	t := m.tokens
	bodyHeight := max(0, m.height-t.chromeRows()-1) // existing header/footer plus Tab row
	left := 0
	gap := lipgloss.Width(t.divider())
	if m.width >= t.layout.stackedBelow {
		percent := t.workspace.sidebarPercent
		if m.workspace.focused {
			percent = t.workspace.compactPercent
		}
		left = max(t.workspace.minPane, m.width*percent/100)
		left = min(left, max(0, m.width-gap-t.workspace.minPane))
	} else {
		gap = 0
	}
	outer := max(0, m.width-left-gap)
	style := m.styles.terminalFrame
	return workspaceGeometry{
		x: left + gap + style.GetBorderLeftSize(), y: 2 + t.layout.frameGap + style.GetBorderTopSize(),
		width: max(0, outer-style.GetHorizontalFrameSize()), height: max(0, bodyHeight-style.GetVerticalFrameSize()),
		left: left, outerWidth: outer, bodyHeight: bodyHeight,
	}
}

func (m *Model) renderWorkspace() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	w, g := m.workspace, m.workspaceGeometry()
	header := m.styles.title.Render("Agent Session Desk")
	count := m.styles.dim.Render(fmt.Sprintf("%d 会话", len(m.sessions)))
	gap := m.width - lipgloss.Width(header) - lipgloss.Width(count)
	if gap > m.tokens.layout.headerMinGap {
		header += strings.Repeat(" ", gap) + count
	}
	var content string
	switch {
	case len(w.tabs) == 0:
		content = m.styles.dim.Render("暂无打开的 Tab · 新建终端或从会话列表重新打开")
	case w.busy:
		content = m.styles.dim.Render("正在连接原终端…")
	case w.err != nil:
		content = m.styles.error.Render("终端不可交互：" + label(w.err.Error()))
	default:
		content = w.screen.Render(g.width, g.height)
	}
	content = fitFrame(content, g.width, g.height)
	frameStyle := m.styles.terminalFrame
	if w.focused {
		frameStyle = m.styles.terminalFocused
	}
	right := frameStyle.Width(g.outerWidth).Height(g.bodyHeight).Render(content)
	body := right
	if g.left > 0 {
		left := m.renderPanelTitle("会话", g.left) + "\n" + m.renderSessionList(g.left, max(0, g.bodyHeight-m.tokens.layout.headingRows))
		separator := m.styles.separator.Render(m.tokens.divider())
		body = lipgloss.JoinHorizontal(lipgloss.Top, fitFrame(left, g.left, g.bodyHeight), strings.TrimSuffix(strings.Repeat(separator+"\n", g.bodyHeight), "\n"), right)
	}
	footer := m.styles.dim.Render("终端模式 · F6 管理 · Ctrl+PgUp/PgDn 切换 Tab")
	if !w.focused {
		footer = m.styles.dim.Render("管理模式 · Enter 终端 · n 新建 · x 关 Tab · h 历史 · q 退出 · F6 返回终端")
	}
	if m.status != "" {
		footer = m.styles.dim.Render(label(m.status))
	}
	return fitFrame(header+"\n"+m.workspaceTabs()+m.tokens.sectionBreak()+fitFrame(body, m.width, g.bodyHeight)+m.tokens.sectionBreak()+footer, m.width, m.height)
}

type workspaceTabCell struct {
	index, x, width int
	text            string
}

func (m *Model) workspaceTabCells() []workspaceTabCell {
	w := m.workspace
	labels := make([]string, len(w.tabs))
	for i, tab := range w.tabs {
		style := m.styles.tab
		if i == w.active {
			style = m.styles.tabActive
		}
		title := label(tab.session.Title)
		if title == "" {
			title = label(tab.session.ID)
		}
		labels[i] = style.Render(title)
	}
	start := w.active
	for start > 0 && lipgloss.Width(strings.Join(labels[start-1:w.active+1], m.tokens.itemGap())) <= m.width {
		start--
	}
	var cells []workspaceTabCell
	x := 0
	for i := start; i < len(labels) && x < m.width; i++ {
		width := min(lipgloss.Width(labels[i]), m.width-x)
		cells = append(cells, workspaceTabCell{index: i, x: x, width: width, text: cellTruncate(labels[i], width, m.tokens.marks.ellipsis)})
		x += width + m.tokens.layout.itemGap
	}
	return cells
}
func (m *Model) workspaceTabs() string {
	if len(m.workspace.tabs) == 0 {
		return m.styles.dim.Render("Tab")
	}
	var labels []string
	for _, cell := range m.workspaceTabCells() {
		labels = append(labels, cell.text)
	}
	return strings.Join(labels, m.tokens.itemGap())
}

func (m *Model) workspaceCursor() *tea.Cursor {
	w := m.workspace
	if !w.focused || w.conn == nil || w.err != nil || !w.screen.Cursor.Visible {
		return nil
	}
	g := m.workspaceGeometry()
	native := w.screen.Cursor
	if native.X < 0 || native.X >= g.width || native.Y < 0 || native.Y >= g.height {
		return nil
	}
	cursor := tea.NewCursor(g.x+native.X, g.y+native.Y)
	switch native.Shape {
	case 3, 4:
		cursor.Shape = tea.CursorUnderline
	case 5, 6:
		cursor.Shape = tea.CursorBar
	}
	cursor.Blink = native.Shape == 0 || native.Shape%2 == 1
	return cursor
}
