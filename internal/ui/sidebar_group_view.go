// Group header adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/home_view_list.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: stable nested groups, discovered terminal counts and token styles;
// no work-state summaries, group concurrency or remote model.
package ui

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) renderSidebarGroup(width int) string {
	return m.renderSidebarGroupAt("default", 0, width)
}
func (m *Model) renderSidebarGroupAt(id string, level, width int) string {
	t := m.tokens.sidebar
	g := m.groupByID(id)
	selected := m.sidebarGroupSelected() && m.selectedGroup().ID == id
	nameStyle, partStyle := m.styles.groupName, m.styles.groupPart
	icon := t.groupExpanded
	if m.groupCollapsed(id) {
		icon = t.groupCollapsed
	}
	gutter := strings.Repeat(" ", t.groupGutter)
	if selected && !m.workspace.focused {
		gutter = m.styles.key.Render(t.selected)
		nameStyle, partStyle = m.styles.groupSelectedName, m.styles.groupSelectedPart
	}
	n := 0
	for _, v := range m.sessions {
		if session.GroupOf(v) == id {
			n++
		}
	}
	name := g.Name
	// Translate only the legacy built-in label, never custom group names.
	if g.ID == session.DefaultGroupID && name == "\u6211\u7684\u4f1a\u8bdd" {
		name = "My Sessions"
	}
	count := t.groupCountPrefix + strconv.Itoa(n) + t.groupCountSuffix
	row := gutter + strings.Repeat(" ", level*t.indentWidth) + partStyle.Render(icon) + strings.Repeat(" ", m.tokens.layout.menuGap) + nameStyle.Render(label(name)) + partStyle.Render(count)
	return cellTruncate(row, width, m.tokens.marks.ellipsis) + strings.Repeat(" ", max(0, width-lipgloss.Width(row)))
}
