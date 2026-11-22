// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/home_view_list.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0, renderSessionList.
// Modified for Agent Session Desk: flat stable-identity session rows, bounded
// viewport and above/below indicators; remove agent/group/provider state,
// jump overlays, background status guesses and embedded terminal cards.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.

package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type sessionListWindow struct {
	start, end, leading int
}

func (m *Model) sessionListWindow(height int) sessionListWindow {
	if height <= 0 || len(m.sessions) == 0 {
		return sessionListWindow{}
	}
	viewOffset := m.viewOffset
	lineBudget := height - m.tokens.layout.indicatorRows // reserve the below indicator as upstream does
	if lineBudget < 1 {
		lineBudget = 1
	}
	if m.cursor < viewOffset {
		viewOffset = m.cursor
	}
	if m.cursor >= viewOffset+lineBudget {
		viewOffset = max(0, m.cursor-lineBudget+1)
	}
	// The above indicator also occupies a row. Ensure the selected item is
	// visible after reserving it, including at two-row and one-row boundaries.
	if viewOffset > 0 && height >= m.tokens.layout.aboveMinHeight {
		lineBudget--
		viewOffset = max(viewOffset, m.cursor-lineBudget+1)
		return sessionListWindow{start: viewOffset, end: min(len(m.sessions), viewOffset+lineBudget), leading: 1}
	}
	return sessionListWindow{start: viewOffset, end: min(len(m.sessions), viewOffset+lineBudget)}
}

func (m *Model) renderSessionList(width, height int) string {
	if height <= 0 || width <= 0 {
		return ""
	}
	if len(m.sessions) == 0 {
		if m.listBusy {
			return m.styles.dim.Render("正在读取会话…")
		}
		return m.styles.dim.Render("暂无会话" + m.tokens.sectionBreak() + "? 查看创建方法")
	}
	var b strings.Builder
	window := m.sessionListWindow(height)
	if window.leading > 0 {
		viewOffset := window.start
		b.WriteString(m.styles.dim.Render(fmt.Sprintf("%s%s 上方 %d", m.tokens.itemGap(), m.tokens.marks.above, viewOffset)))
		b.WriteString("\n")
	}
	nextItem := window.start
	for i := window.start; i < window.end; i++ {
		s := m.sessions[i]
		title := label(s.Title)
		if title == "" {
			title = label(s.ID)
		}
		prefix := strings.Repeat(" ", m.tokens.layout.rowPrefixWidth)
		if i == m.cursor {
			prefix = m.tokens.marks.selected + strings.Repeat(" ", max(0, m.tokens.layout.rowPrefixWidth-lipgloss.Width(m.tokens.marks.selected)))
		}
		state := s.TerminalState
		if m.listErr != nil {
			state = session.TerminalUnknown
		}
		suffix := m.tokens.itemGap() + terminalLabel(state)
		if width < m.tokens.layout.suffixBelow {
			suffix = ""
		}
		nameWidth := max(0, width-lipgloss.Width(prefix)-lipgloss.Width(suffix))
		row := prefix + cellTruncate(title, nameWidth, m.tokens.marks.ellipsis)
		row += strings.Repeat(" ", max(0, width-lipgloss.Width(row)-lipgloss.Width(suffix))) + suffix
		if i == m.cursor {
			row = m.styles.highlight.Render(row)
		} else {
			row = m.styles.text.Render(row)
		}
		b.WriteString(row)
		b.WriteString("\n")
		nextItem = i + 1
	}
	if remaining := len(m.sessions) - nextItem; remaining > 0 && height > 1 {
		b.WriteString(m.styles.dim.Render(fmt.Sprintf("%s%s 下方 %d", m.tokens.itemGap(), m.tokens.marks.below, remaining)))
	}
	return strings.TrimRight(b.String(), "\n")
}

func terminalLabel(state session.TerminalState) string {
	switch state {
	case session.TerminalAvailable:
		return "可连接"
	case session.TerminalExited:
		return "已退出"
	default:
		return "未知"
	}
}
