// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/home_view_frame.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified for Agent Session Desk: retain responsive dual/stacked/single
// layout, panel titles and exact-size helpers; route product Tab terminals, remove
// filter/account/system panels and v1 renderer controls; use Bubble Tea v2.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.

package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) View() tea.View {
	if m.workspace != nil && m.help {
		view := tea.NewView(m.renderHelpOverlay())
		view.AltScreen = true
		view.BackgroundColor, view.ForegroundColor = m.tokens.colors.bg, m.tokens.colors.text
		return view
	}
	if m.workspace != nil && m.sessionStop != nil {
		view := tea.NewView(m.renderSessionStop())
		view.AltScreen = true
		view.BackgroundColor, view.ForegroundColor = m.tokens.colors.bg, m.tokens.colors.text
		return view
	}
	if m.workspace != nil && m.groupRemoval != nil {
		view := tea.NewView(m.renderGroupRemoval())
		view.AltScreen = true
		view.BackgroundColor, view.ForegroundColor = m.tokens.colors.bg, m.tokens.colors.text
		return view
	}
	if m.workspace != nil && m.groupDialog != nil {
		content, cursor := m.renderGroupDialog()
		view := tea.NewView(content)
		view.AltScreen = true
		view.BackgroundColor, view.ForegroundColor = m.tokens.colors.bg, m.tokens.colors.text
		view.Cursor = cursor
		return view
	}
	if m.workspace != nil && m.rename != nil {
		content, cursor := m.renderSessionRename()
		view := tea.NewView(content)
		view.AltScreen = true
		view.BackgroundColor, view.ForegroundColor = m.tokens.colors.bg, m.tokens.colors.text
		view.Cursor = cursor
		return view
	}
	if m.workspace != nil && m.search != nil {
		content, cursor := m.renderSessionSearch()
		view := tea.NewView(content)
		view.AltScreen = true
		view.BackgroundColor, view.ForegroundColor = m.tokens.colors.bg, m.tokens.colors.text
		view.Cursor = cursor
		return view
	}
	if m.workspace != nil && m.densitySettings != nil {
		view := tea.NewView(m.renderDensitySettings())
		view.AltScreen = true
		view.BackgroundColor, view.ForegroundColor = m.tokens.colors.bg, m.tokens.colors.text
		return view
	}
	if m.workspace != nil && m.launch != nil {
		content, cursor := m.renderLaunchDialog()
		view := tea.NewView(content)
		view.AltScreen = true
		view.BackgroundColor, view.ForegroundColor = m.tokens.colors.bg, m.tokens.colors.text
		view.Cursor = cursor
		return view
	}
	if m.workspace != nil && !m.history.open && !m.help {
		view := tea.NewView(m.renderWorkspace())
		view.AltScreen = true
		view.BackgroundColor, view.ForegroundColor = m.tokens.colors.bg, m.tokens.colors.text
		view.MouseMode = tea.MouseModeAllMotion
		view.Cursor = m.workspaceCursor()
		return view
	}
	view := tea.NewView(m.renderFrame())
	view.AltScreen = true
	return view
}

func (m *Model) renderFrame() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	header := m.styles.title.Render("Agent Session Desk")
	count := m.styles.dim.Render(fmt.Sprintf("%d 会话", len(m.sessions)))
	if m.history.open {
		count = m.styles.dim.Render(fmt.Sprintf("%d 条历史", len(m.history.rows)))
	}
	gap := m.width - lipgloss.Width(header) - lipgloss.Width(count)
	if gap > m.tokens.layout.headerMinGap {
		header += strings.Repeat(" ", gap) + count
	}
	if m.height < m.tokens.layout.minFrameHeight {
		return fitFrame(header+"\n"+m.styles.dim.Render("请增大窗口 · q 退出"), m.width, m.height)
	}
	contentHeight := m.height - m.tokens.chromeRows() // header, breathing row, breathing row, footer
	var content string
	switch {
	case m.history.open:
		content = m.renderHistory(contentHeight)
	case m.help:
		content = m.renderHelp(contentHeight)
	case !m.previewVisible():
		content = m.renderSingleColumnLayout(contentHeight)
	case m.width < m.tokens.layout.stackedBelow:
		content = m.renderStackedLayout(contentHeight)
	default:
		content = m.renderDualColumnLayout(contentHeight)
	}
	footer := m.renderHelpBar()
	if m.history.open {
		footer = m.renderHistoryHelpBar()
	}
	if !m.history.open && m.status != "" {
		footer = m.styles.dim.Render(cellTruncate(label(m.status), max(0, m.width-m.tokens.layout.statusHelpWidth), m.tokens.marks.ellipsis)) + m.tokens.itemGap() + "? 帮助" + m.tokens.itemGap() + "q 退出"
	}
	if !m.history.open && m.listErr != nil {
		footer = m.styles.error.Render("后台状态未知："+label(m.listErr.Error())) + m.tokens.itemGap() + "r 重试"
	}
	return fitFrame(header+m.tokens.sectionBreak()+ensureExactHeight(content, contentHeight)+m.tokens.sectionBreak()+footer, m.width, m.height)
}

func (m *Model) renderPanelTitle(title string, width int) string {
	if width <= 0 {
		return ""
	}
	title = cellTruncate(title, width, m.tokens.marks.ellipsis)
	titleStyle := m.styles.panelTitle.Width(width)
	underlineStyle := m.styles.separator.Width(width)
	return titleStyle.Render(title) + "\n" + underlineStyle.Render(strings.Repeat(m.tokens.marks.rule, width))
}

// ensureExactHeight is retained from Agent Deck. Having exactly the requested
// number of rows prevents one panel from moving its neighbour when joined.
func ensureExactHeight(content string, n int) string {
	if n <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) > n {
		lines = lines[:n]
	} else {
		for len(lines) < n {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

// ensureExactWidth adapts Agent Deck's per-row width normalization. Truncate
// by display cells rather than bytes/runes so CJK and emoji stay in their pane.
func ensureExactWidth(content string, width int) string {
	if width <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) > width {
			line = cellTruncate(line, width, "")
		}
		padding := max(0, width-lipgloss.Width(line))
		lines[i] = line + strings.Repeat(" ", padding)
	}
	return strings.Join(lines, "\n")
}

func fitFrame(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	return ensureExactWidth(ensureExactHeight(content, height), width)
}

func (m *Model) renderDualColumnLayout(contentHeight int) string {
	dividerWidth := lipgloss.Width(m.tokens.divider())
	leftWidth := min(m.tokens.layout.listMax, max(m.tokens.layout.listMin, (m.width-dividerWidth)/m.tokens.layout.listFraction))
	rightWidth := m.width - leftWidth - dividerWidth
	panelContentHeight := max(0, contentHeight-m.tokens.layout.headingRows)
	leftPanel := m.renderPanelTitle("会话", leftWidth) + "\n" +
		ensureExactHeight(m.renderSessionList(leftWidth, panelContentHeight), panelContentHeight)
	rightPanel := m.renderPanelTitle("屏幕预览", rightWidth) + "\n" +
		ensureExactHeight(m.renderPreviewPane(rightWidth, panelContentHeight), panelContentHeight)

	// Style the constant separator once, then repeat it for each exact row.
	separatorCell := m.styles.separator.Render(m.tokens.divider())
	separatorLines := make([]string, contentHeight)
	for i := range separatorLines {
		separatorLines[i] = separatorCell
	}
	leftPanel = fitFrame(leftPanel, leftWidth, contentHeight)
	rightPanel = fitFrame(rightPanel, rightWidth, contentHeight)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel,
		strings.Join(separatorLines, "\n"), rightPanel)
}

func (m *Model) renderStackedLayout(totalHeight int) string {
	// Unlike the upstream minimum floors, clamp both regions to the actual
	// available height; a short terminal must never overflow into the footer.
	listHeight := max(m.tokens.layout.stackedListMin, totalHeight/m.tokens.layout.stackedFraction)
	previewHeight := max(0, totalHeight-listHeight-m.tokens.layout.stackedGap)
	list := m.renderPanelTitle("会话", m.width) + "\n" +
		ensureExactHeight(m.renderSessionList(m.width, listHeight-m.tokens.layout.headingRows), listHeight-m.tokens.layout.headingRows)
	preview := m.renderPanelTitle("屏幕预览", m.width) + "\n" +
		ensureExactHeight(m.renderPreviewPane(m.width, previewHeight-m.tokens.layout.headingRows), previewHeight-m.tokens.layout.headingRows)
	return ensureExactHeight(list, listHeight) + "\n" +
		ensureExactHeight(m.styles.separator.Render(strings.Repeat(m.tokens.marks.rule, m.width)), m.tokens.layout.stackedGap) + "\n" +
		ensureExactHeight(preview, previewHeight)
}

func (m *Model) renderSingleColumnLayout(totalHeight int) string {
	return m.renderPanelTitle("会话", m.width) + "\n" +
		ensureExactHeight(m.renderSessionList(m.width, totalHeight-m.tokens.layout.headingRows), totalHeight-m.tokens.layout.headingRows)
}

func (m *Model) renderPreviewPane(width, height int) string {
	s, ok := m.selected()
	if !ok {
		return m.styles.dim.Render("选择会话后查看当前屏幕" + m.tokens.sectionBreak() + "? 查看创建方法")
	}
	if m.listErr != nil {
		return m.styles.dim.Render("后台连接尚未确认；按 r 重试")
	}
	if m.previewErr != nil {
		return m.styles.error.Render("预览不可用：\n" + label(m.previewErr.Error()))
	}
	if s.TerminalState != session.TerminalAvailable && s.TerminalState != session.TerminalExited {
		return m.styles.dim.Render("终端" + terminalLabel(s.TerminalState) + "；CLI 工作状态未知")
	}
	if m.preview == "" {
		if m.previewBusy {
			return m.styles.dim.Render("正在读取当前屏幕…")
		}
		return m.styles.dim.Render("当前屏幕暂无文字")
	}
	// The capture is a snapshot, not durable history or a reconstructed
	// conversation. Keep its newest visible rows without retaining more copies.
	lines := strings.Split(m.preview, "\n")
	budget := max(0, height-1-m.tokens.layout.captionGap)
	clipped := len(lines) > budget
	if clipped {
		lines = lines[len(lines)-budget:]
	}
	caption := "当前屏幕 · CLI 工作状态未知"
	if clipped || m.previewTruncated {
		caption = "当前屏幕（节选）· CLI 状态未知"
	}
	if s.TerminalState == session.TerminalExited {
		caption = "终端已退出 · 保留屏幕，不可输入"
		if clipped || m.previewTruncated {
			caption = "终端已退出 · 保留屏幕（节选）"
		}
	}
	return m.styles.dim.Render(cellTruncate(caption, width, m.tokens.marks.ellipsis)) + strings.Repeat("\n", m.tokens.layout.captionGap+1) +
		strings.Join(lines, "\n")
}

func (m *Model) renderHelpBar() string {
	if m.width < m.tokens.layout.singleBelow {
		return m.styles.menuKey("↑↓", "选择") + m.tokens.itemGap() + m.styles.menuKey("↵", "进入") + m.tokens.itemGap() + "?" + m.tokens.itemGap() + "q"
	}
	parts := []string{
		m.styles.menuKey("↑↓", "选择"), m.styles.menuKey("Enter", "进入"),
		m.styles.menuKey("h", "历史"), m.styles.menuKey("r", "刷新"), m.styles.menuKey("?", "帮助"), m.styles.menuKey("q", "退出"),
	}
	return strings.Join(parts, m.tokens.itemGap())
}

func (m *Model) renderHelp(height int) string {
	content := m.renderPanelTitle("使用方法", m.width) + m.tokens.sectionBreak() +
		"在另一个终端创建会话：\n" +
		m.styles.text.Render(m.tokens.itemGap()+"asd start --name 开发 -- codex") + "\n" +
		m.styles.text.Render(m.tokens.itemGap()+"asd start --name 评审 --cwd 路径 -- 自定义程序 参数") + m.tokens.sectionBreak() +
		"↑↓ / j k 选择会话；Enter 全屏进入。\n" +
		"分离前取消候选并切至英文输入模式。\n" +
		"Ctrl+b，松开后按 d，返回管理器。\n" +
		"h 只读历史；Enter 详情，Esc 返回。\n" +
		"q 退出管理器，保留后台会话。" + m.tokens.sectionBreak() + "" +
		m.styles.dim.Render("预览仅为当前屏幕；CLI 工作状态保持未知。") + "\n" +
		"Esc / ? 返回"
	return ensureExactHeight(content, height)
}
