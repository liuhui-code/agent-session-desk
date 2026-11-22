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

type workspaceGeometry struct{ x, y, width, height, left, outerWidth, bodyHeight, above, terminalHeight, listWidth, listHeight int }

func (m *Model) workspaceGeometry() workspaceGeometry {
	t := m.tokens
	bodyHeight := max(0, m.height-t.chromeRows()-1) // existing header/footer plus Tab row
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
	minList := t.layout.headingRows + t.layout.stackedListMin
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
		x: left + gap + style.GetBorderLeftSize(), y: 2 + t.layout.frameGap + above + style.GetBorderTopSize(),
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
		if w.screen.Width > 0 && w.screen.Height > 0 {
			content = w.screen.Render(g.width, g.height)
		} else {
			content = m.styles.error.Render("终端不可交互：" + label(w.err.Error()))
		}
	default:
		content = w.screen.Render(g.width, g.height)
	}
	content = fitFrame(content, g.width, g.height)
	frameStyle := m.styles.terminalFrame
	if w.focused {
		frameStyle = m.styles.terminalFocused
	}
	right := frameStyle.Width(g.outerWidth).Height(g.terminalHeight).Render(content)
	body := right
	if g.left > 0 {
		left := m.renderPanelTitle("会话", g.left) + "\n" + m.renderSessionList(g.left, max(0, g.bodyHeight-m.tokens.layout.headingRows))
		separatorStyle := m.styles.separator
		if w.split.dragging {
			separatorStyle = m.styles.separatorDrag
		}
		separator := separatorStyle.Render(m.tokens.divider())
		body = lipgloss.JoinHorizontal(lipgloss.Top, fitFrame(left, g.left, g.bodyHeight), strings.TrimSuffix(strings.Repeat(separator+"\n", g.bodyHeight), "\n"), right)
	}
	if g.above > 0 {
		listHeight := g.listHeight
		list := m.renderPanelTitle("会话", m.width) + "\n" + m.renderSessionList(m.width, max(0, listHeight-m.tokens.layout.headingRows))
		body = fitFrame(list, m.width, listHeight)
		if g.terminalHeight > 0 {
			body += strings.Repeat("\n", m.tokens.layout.stackedGap+1) + right
		}
	}
	footer := m.styles.dim.Render("终端模式 · F6 管理 · Ctrl+PgUp/PgDn 切换 Tab")
	if !w.focused {
		footer = m.styles.dim.Render("管理模式 · Enter 终端 · n 新建 · x 关Tab · <> 比例 · = 默认 · ? 帮助 · q 退出 · F6 终端")
	} else if w.screen.CopyMode {
		text := fmt.Sprintf("滚动历史 · %d 行 · Esc/q 返回 · ↑↓/PgUp/PgDn 浏览", w.screen.ScrollbackOffset)
		if w.screen.Status != "" {
			text = "滚动历史 · " + label(w.screen.Status)
		}
		footer = m.styles.dim.Render(text)
	} else if w.screen.ScrollbackOffset > 0 {
		footer = m.styles.dim.Render(fmt.Sprintf("滚动位置 · %d 行 · 向下滚动返回实时输出", w.screen.ScrollbackOffset))
	}
	if !w.split.noticeUntil.IsZero() {
		pct := m.workspaceSidebarPercent()
		footer = m.styles.dim.Render(fmt.Sprintf("会话 %d%% / 终端 %d%%", pct, 100-pct))
	}
	if m.status != "" {
		footer = m.styles.dim.Render(label(m.status))
	}
	if w.err != nil && w.screen.Width > 0 && w.screen.Height > 0 {
		footer = m.styles.error.Render("只读：上次观察屏幕；连接不可用，r 重查。" + label(w.err.Error()))
	}
	return fitFrame(header+"\n"+m.workspaceTabs()+m.tokens.sectionBreak()+fitFrame(body, m.width, g.bodyHeight)+m.tokens.sectionBreak()+footer, m.width, m.height)
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
