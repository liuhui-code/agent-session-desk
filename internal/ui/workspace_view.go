// Layout portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Sources: internal/ui/home_view_frame.go and split_config.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: product Tabs, one native
// psmux screen, v2 size/cursor handling and token-only geometry.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type workspaceGeometry struct{ x, y, width, height, left, outerWidth, bodyHeight, above, terminalHeight, listWidth, listHeight int }

func (m *Model) workspaceGeometry() workspaceGeometry {
	t := m.tokens
	bodyHeight := max(0, m.height-t.workspaceChromeRows()) // Full-height terminal; hints overlay without resizing the child.
	left := 0
	gap := lipgloss.Width(t.divider())
	if m.width >= t.layout.stackedBelow {
		percent := m.workspaceSidebarPercent()
		left = max(t.workspace.minPane, m.width*percent/100)
		left = min(left, max(0, m.width-gap-t.workspace.minPane))
	} else {
		gap = 0
	}
	outer := max(0, m.width-left-gap)
	style := m.styles.terminalFrame
	above, terminalHeight := 0, bodyHeight
	minList := t.layout.stackedListMin
	minFrame := style.GetVerticalFrameSize() + t.workspace.minTerminalRows
	if m.width < t.layout.stackedBelow && !m.workspace.focused {
		above, terminalHeight = bodyHeight, 0
		if m.width >= t.layout.singleBelow && bodyHeight >= minList+t.layout.stackedGap+minFrame {
			listHeight := max(minList, bodyHeight*m.workspaceSidebarPercent()/100)
			listHeight = min(listHeight, bodyHeight-t.layout.stackedGap-minFrame)
			above = listHeight + t.layout.stackedGap
			terminalHeight = bodyHeight - above
		}
	}
	listWidth, listHeight := left, bodyHeight
	if above > 0 {
		listWidth, listHeight = m.width, above
		if terminalHeight > 0 {
			listHeight -= t.layout.stackedGap
		}
	}

	return workspaceGeometry{
		x: left + gap + style.GetBorderLeftSize(), y: above + style.GetBorderTopSize(),
		width: max(0, outer-style.GetHorizontalFrameSize()), height: max(0, terminalHeight-style.GetVerticalFrameSize()),
		left: left, outerWidth: outer, bodyHeight: bodyHeight, above: above, terminalHeight: terminalHeight,
		listWidth: listWidth, listHeight: listHeight,
	}
}

func (m *Model) renderWorkspace() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	w, g := m.workspace, m.workspaceGeometry()
	var content string
	switch {
	case len(w.tabs) == 0:
		content = m.styles.dim.Render("暂无打开的 Tab · 新建终端或从会话列表重新打开")
	case w.busy:
		content = m.styles.dim.Render("正在连接原终端…")
	case w.err != nil:
		if w.screen.Width > 0 && w.screen.Height > 0 {
			content = w.screen.RenderWithCopyStyles(g.width, g.height, m.styles.copy)
		} else {
			content = m.styles.error.Render("终端不可交互：" + label(w.err.Error()))
		}
	default:
		content = w.screen.RenderWithCopyStyles(g.width, g.height, m.styles.copy)
	}
	content = fitFrame(content, g.width, g.height)
	frameStyle := m.styles.terminalFrame
	if w.focused {
		frameStyle = m.styles.terminalFocused
	}
	right := frameStyle.Width(g.outerWidth).Height(g.terminalHeight).Render(content)
	body := right
	if g.left > 0 {
		left := m.renderSessionList(g.left, g.bodyHeight)
		separatorStyle := m.styles.separator
		if w.split.hovered || w.split.dragging {
			separatorStyle = m.styles.separatorDrag
		}
		divider := m.tokens.divider()
		if w.split.hovered || w.split.dragging {
			divider = m.tokens.dividerWith(m.tokens.marks.dividerHandle)
		}
		separator := separatorStyle.Render(divider)
		separatorLines := make([]string, g.bodyHeight)
		for y := range separatorLines {
			separatorLines[y] = strings.Repeat(" ", lipgloss.Width(divider))
			if y >= m.tokens.workspace.dividerInsetRows && y < g.bodyHeight-m.tokens.workspace.dividerInsetRows {
				separatorLines[y] = separator
			}
		}
		body = lipgloss.JoinHorizontal(lipgloss.Top, fitFrame(left, g.left, g.bodyHeight), strings.Join(separatorLines, "\n"), right)
	}
	if g.above > 0 {
		listHeight := g.listHeight
		list := m.renderSessionList(m.width, listHeight)
		body = fitFrame(list, m.width, listHeight)
		if g.terminalHeight > 0 {
			body += strings.Repeat("\n", m.tokens.layout.stackedGap+1) + right
		}
	}
	return m.renderWorkspaceOverlay(fitFrame(body, m.width, m.height))
}

func (m *Model) workspaceCursor() *tea.Cursor {
	w := m.workspace
	if _, cursor := m.workspaceCopyPrompt(); cursor != nil {
		return cursor
	}
	if !w.focused || w.conn == nil || w.err != nil {
		return nil
	}
	g := m.workspaceGeometry()
	native := w.screen.Cursor
	if w.screen.CopyMode && w.screen.CopyCursor != nil {
		native = *w.screen.CopyCursor
	}
	if !native.Visible {
		return nil
	}
	if native.X < 0 || native.X >= g.width || native.Y < 0 || native.Y >= g.height {
		return nil
	}
	if m.workspaceFooterVisible() && g.y+native.Y == m.height-1 {
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
	if w.screen.CopyMode {
		cursor.Shape, cursor.Blink = m.tokens.copy.cursorShape, m.tokens.copy.cursorBlink
	}
	return cursor
}
