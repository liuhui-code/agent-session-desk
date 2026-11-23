// Footer roles and fitting adapted from Agent Deck, Copyright (c) 2025
// Ashesh Goplani. SPDX-License-Identifier: MIT
// Source: home_view_frame.go and insert_mode.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: actual supported product
// actions, Tab/F6 ownership, stable v2 tokens and cell-bounded localized hints.
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type footerHint struct{ key, label string }

func (m *Model) managementFooterHints() (primary, secondary []footerHint) {
	group := m.sidebarGroupSelected()
	if group {
		label := "折叠"
		if m.groupCollapsed(m.selectedGroup().ID) {
			label = "展开"
		}
		primary = append(primary, footerHint{"Enter", label})
	} else if selected, ok := m.selected(); ok && selected.TerminalState == session.TerminalAvailable {
		primary = append(primary, footerHint{"Enter", "打开"})
	}
	primary = append(primary, footerHint{"n", "新建"})
	name := "改名"
	if m.workspace.err != nil {
		name = "重查"
	}
	if group {
		if _, ok := m.service.(groupService); ok {
			primary = append(primary, footerHint{"g", "分组"})
		}
		secondary = append(secondary, footerHint{"r", name})
		if m.selectedGroup().ID != session.DefaultGroupID {
			secondary = append(secondary, footerHint{"d", "删组"})
		}
	} else if _, ok := m.selected(); ok {
		secondary = append(secondary, footerHint{"r", name})
		if _, ok := m.service.(visiblePaneService); ok {
			secondary = append(secondary, footerHint{"V", "复制页"})
		}
		if _, ok := m.service.(groupService); ok {
			secondary = append(secondary, footerHint{"M", "移动"})
		}
	}
	return
}

func (m *Model) renderManagementFooter() string {
	t := m.tokens.footer
	space := strings.Repeat(" ", t.labelGap)
	if m.width < m.tokens.layout.singleBelow {
		choices := []string{"管理模式" + strings.Repeat(" ", t.blockGap) + strings.Join([]string{"F6", "?", "q"}, t.inlineSeparator), "F6 ? q", "? q", "?"}
		for _, text := range choices {
			if lipgloss.Width(text) <= m.width {
				return strings.Repeat(" ", max(0, (m.width-lipgloss.Width(text))/2)) + m.styles.dim.Render(text)
			}
		}
		return ""
	}
	minimal := m.width < t.compactBelow
	keyStyle := m.styles.footerKey
	rightGutter := t.compactRightGutter
	if m.width > t.fullAbove {
		keyStyle = m.styles.footerFullKey
		rightGutter = t.fullRightGutter
	}
	render := func(h footerHint) string {
		key := keyStyle.Render(h.key)
		if minimal {
			return key
		}
		return key + space + m.styles.footerDescription.Render(h.label)
	}
	primary, secondary := m.managementFooterHints()
	hints := append(append([]footerHint(nil), primary...), secondary...)
	context := m.styles.footerContext.Render("管理模式" + t.contextSuffix)
	sep := m.styles.footerSeparator.Render(t.separator)
	joinHints := func(n int) string {
		parts := make([]string, 0, n)
		for _, h := range hints[:n] {
			parts = append(parts, render(h))
		}
		left := context
		if n > 0 {
			left += space + strings.Join(parts, strings.Repeat(" ", t.hintGap))
		}
		if n < len(hints) {
			left += space + m.styles.dim.Render(m.tokens.marks.ellipsis)
		}
		return left
	}
	pinned := []footerHint{{"↑↓", "选择"}, {"F6", "终端"}, {"?", "帮助"}, {"q", "退出"}}
	optional := []footerHint{{"/", "搜索"}, {"S", "设置"}, {"Alt+V", "分组/平铺"}}
	right := func(n int) string {
		parts := make([]string, 0, len(pinned)+n)
		ordered := append(append([]footerHint(nil), pinned[:len(pinned)-1]...), optional[:n]...)
		ordered = append(ordered, pinned[len(pinned)-1])
		for _, h := range ordered {
			text := h.key
			if !minimal {
				text += space + h.label
			}
			parts = append(parts, m.styles.dim.Render(text))
		}
		return strings.Join(parts, sep)
	}
	assemble := func(left, right string) (string, bool) {
		if lipgloss.Width(left)+t.blockGap+lipgloss.Width(right) > m.width {
			return "", false
		}
		gap := max(t.blockGap, m.width-lipgloss.Width(left)-lipgloss.Width(right)-rightGutter)
		return left + strings.Repeat(" ", gap) + right, true
	}
	for n := len(optional); n >= 0; n-- {
		if line, ok := assemble(joinHints(len(hints)), right(n)); ok {
			return line
		}
	}
	for n := len(hints) - 1; n >= 0; n-- {
		if line, ok := assemble(joinHints(n), right(0)); ok {
			return line
		}
	}
	return fitFrame(right(0), m.width, 1)
}

func (m *Model) renderTerminalFooter() string {
	t := m.tokens.footer
	space := strings.Repeat(" ", t.labelGap)
	gap := strings.Repeat(" ", t.blockGap)
	badge := m.styles.footerBadge.Render(t.badgePrefix + "终端模式" + t.badgeSuffix)
	choices := []string{"F6 管理" + t.inlineSeparator + "Ctrl+PgUp/PgDn 切换 Tab", "F6 管理" + t.inlineSeparator + "Ctrl+PgUp/PgDn", "F6 管理"}
	hint := choices[len(choices)-1]
	for _, choice := range choices {
		if lipgloss.Width(badge)+t.blockGap+lipgloss.Width(choice) <= m.width {
			hint = choice
			break
		}
	}
	if lipgloss.Width(badge)+t.blockGap+lipgloss.Width(hint) > m.width {
		badge = m.styles.footerCompactBadge.Render("终端模式")
	}
	if lipgloss.Width(badge)+t.blockGap+lipgloss.Width(hint) > m.width {
		return cellTruncate(m.styles.dim.Render("F6"), m.width, "")
	}
	line := badge
	budget := m.width - lipgloss.Width(badge) - t.labelGap - lipgloss.Width(t.targetPrefix) - t.blockGap - lipgloss.Width(hint)
	w := m.workspace
	if budget > 0 && w.active >= 0 && w.active < len(w.tabs) {
		// Measure at a normal Tab width first: a tiny optional footer budget
		// must not turn a colliding name into only the end of its ID suffix.
		title := m.sessionTitle(w.tabs[w.active].session, m.titlePeers(), max(budget, m.tokens.workspace.maxTabWidth))
		if title != "" && lipgloss.Width(title) <= budget {
			line += space + m.styles.footerDescription.Render(t.targetPrefix+title)
		}
	}
	return line + gap + m.styles.dim.Render(hint)
}
