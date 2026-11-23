// Group header adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/home_view_list.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: implicit default group, discovered terminal count and token styles;
// no work-state summaries, group concurrency or remote model.
package ui

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) renderSidebarGroup(width int) string {
	t := m.tokens.sidebar
	selected := m.sidebarGroupSelected()
	nameStyle, partStyle := m.styles.groupName, m.styles.groupPart
	icon := t.groupExpanded
	if m.sidebarCollapsed {
		icon = t.groupCollapsed
	}
	gutter := m.styles.groupHotkey.Render(t.rootShortcut)
	if selected {
		gutter = m.styles.key.Render(t.selected)
		nameStyle, partStyle = m.styles.groupSelectedName, m.styles.groupSelectedPart
	}
	if m.workspace.focused && !selected {
		gutter = ""
	}
	count := t.groupCountPrefix + strconv.Itoa(len(m.sessions)) + t.groupCountSuffix
	row := gutter + partStyle.Render(icon) + strings.Repeat(" ", m.tokens.layout.menuGap) + nameStyle.Render("我的会话") + partStyle.Render(count)
	return cellTruncate(row, width, m.tokens.marks.ellipsis) + strings.Repeat(" ", max(0, width-lipgloss.Width(row)))
}
