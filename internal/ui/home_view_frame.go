// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/home_view_frame.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified for Agent Session Desk: retain responsive dual/stacked/single
// layout, panel titles and exact-size helpers; remove embedded terminal,
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

const (
	layoutBreakpointSingle  = 50
	layoutBreakpointStacked = 80
)

func (m *Model) View() tea.View {
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
	gap := m.width - lipgloss.Width(header) - lipgloss.Width(count)
	if gap > 1 {
		header += strings.Repeat(" ", gap) + count
	}
	if m.height < 5 {
		return fitFrame(header+"\n"+m.styles.dim.Render("请增大窗口 · q 退出"), m.width, m.height)
	}
	contentHeight := m.height - 4 // header, breathing row, breathing row, footer
	var content string
	switch {
	case m.help:
		content = m.renderHelp(contentHeight)
	case !m.previewVisible():
		content = m.renderSingleColumnLayout(contentHeight)
	case m.width < layoutBreakpointStacked:
		content = m.renderStackedLayout(contentHeight)
	default:
		content = m.renderDualColumnLayout(contentHeight)
	}
	footer := m.renderHelpBar()
	if m.status != "" {
		footer = m.styles.dim.Render(cellTruncate(label(m.status), max(0, m.width-18), "…")) + "  ? 帮助  q 退出"
	}
	if m.listErr != nil {
		footer = m.styles.error.Render("后台状态未知："+label(m.listErr.Error())) + "  r 重试"
	}
	return fitFrame(header+"\n\n"+ensureExactHeight(content, contentHeight)+"\n\n"+footer, m.width, m.height)
}

func (m *Model) renderPanelTitle(title string, width int) string {
	if width <= 0 {
		return ""
	}
	title = cellTruncate(title, width, "…")
	titleStyle := lipgloss.NewStyle().Foreground(darkColors.cyan).Bold(true).Width(width)
	underlineStyle := lipgloss.NewStyle().Foreground(darkColors.border).Width(width)
	return titleStyle.Render(title) + "\n" + underlineStyle.Render(strings.Repeat("─", width))
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
	leftWidth := min(42, max(24, (m.width-3)/3))
	rightWidth := m.width - leftWidth - 3
	panelContentHeight := max(0, contentHeight-2)
	leftPanel := m.renderPanelTitle("会话", leftWidth) + "\n" +
		ensureExactHeight(m.renderSessionList(leftWidth, panelContentHeight), panelContentHeight)
	rightPanel := m.renderPanelTitle("屏幕预览", rightWidth) + "\n" +
		ensureExactHeight(m.renderPreviewPane(rightWidth, panelContentHeight), panelContentHeight)

	// Style the constant separator once, then repeat it for each exact row.
	separatorCell := m.styles.separator.Render(" │ ")
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
	listHeight := max(3, totalHeight/2)
	previewHeight := max(0, totalHeight-listHeight-1)
	list := m.renderPanelTitle("会话", m.width) + "\n" +
		ensureExactHeight(m.renderSessionList(m.width, listHeight-2), listHeight-2)
	preview := m.renderPanelTitle("屏幕预览", m.width) + "\n" +
		ensureExactHeight(m.renderPreviewPane(m.width, previewHeight-2), previewHeight-2)
	return ensureExactHeight(list, listHeight) + "\n" +
		m.styles.separator.Render(strings.Repeat("─", m.width)) + "\n" +
		ensureExactHeight(preview, previewHeight)
}

func (m *Model) renderSingleColumnLayout(totalHeight int) string {
	return m.renderPanelTitle("会话", m.width) + "\n" +
		ensureExactHeight(m.renderSessionList(m.width, totalHeight-2), totalHeight-2)
}

func (m *Model) renderPreviewPane(width, height int) string {
	s, ok := m.selected()
	if !ok {
		return m.styles.dim.Render("选择会话后查看当前屏幕\n\n? 查看创建方法")
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
	budget := max(0, height-2)
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
	return m.styles.dim.Render(cellTruncate(caption, width, "…")) + "\n\n" +
		strings.Join(lines, "\n")
}

func (m *Model) renderHelpBar() string {
	if m.width < 50 {
		return m.styles.menuKey("↑↓", "选择") + "  " + m.styles.menuKey("↵", "进入") + "  ?  q"
	}
	parts := []string{
		m.styles.menuKey("↑↓", "选择"), m.styles.menuKey("Enter", "进入"),
		m.styles.menuKey("r", "刷新"), m.styles.menuKey("?", "帮助"), m.styles.menuKey("q", "退出"),
	}
	return strings.Join(parts, "  ")
}

func (m *Model) renderHelp(height int) string {
	content := m.renderPanelTitle("使用方法", m.width) + "\n\n" +
		"在另一个终端创建会话：\n" +
		m.styles.text.Render("  asd start --name 开发 -- codex") + "\n" +
		m.styles.text.Render("  asd start --name 评审 --cwd 路径 -- 自定义程序 参数") + "\n\n" +
		"↑↓ / j k 选择会话；Enter 全屏进入。\n" +
		"分离前取消候选并切至英文输入模式。\n" +
		"Ctrl+b，松开后按 d，返回管理器。\n" +
		"q 退出管理器，保留后台会话。\n\n" +
		m.styles.dim.Render("预览仅为当前屏幕；CLI 工作状态保持未知。") + "\n" +
		"Esc / ? 返回"
	return ensureExactHeight(content, height)
}
