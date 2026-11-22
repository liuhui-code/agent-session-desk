// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/embedded_sidebar.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: flat Session/Run cards, default compact density, terminal-only
// observations and display command; no inferred agent status or branding.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) renderWorkspaceSidebar(width, height int) string {
	var lines []string
	rowLines := m.workspaceSidebarRows(width, height)
	window := m.sessionListWindowForRows(height, rowLines)
	if window.leading > 0 {
		lines = append(lines, m.styles.dim.Render(fmt.Sprintf("%s%s 上方 %d", m.tokens.itemGap(), m.tokens.marks.above, window.start)))
	}
	for i := window.start; i < window.end; i++ {
		lines = append(lines, m.workspaceSidebarCard(i, width, rowLines)...)
	}
	if remaining := len(m.sessions) - window.end; remaining > 0 && height > len(lines) {
		lines = append(lines, m.styles.dim.Render(fmt.Sprintf("%s%s 下方 %d", m.tokens.itemGap(), m.tokens.marks.below, remaining)))
	}
	return strings.Join(lines[:min(height, len(lines))], "\n")
}

func (m *Model) workspaceSidebarRows(width, height int) int {
	if width < m.tokens.sidebar.cardMinWidth {
		return 1
	}
	return max(1, min(height, m.tokens.sidebar.cardRows))
}

func (m *Model) workspaceSidebarCard(index, width, rowLines int) []string {
	s := m.sessions[index]
	title := label(s.Title)
	if title == "" {
		title = label(s.ID)
	}
	prefix := strings.Repeat(" ", m.tokens.layout.rowPrefixWidth)
	selected := index == m.cursor
	if selected {
		prefix = m.tokens.sidebar.selected
	}
	w := m.workspace
	active := w.active >= 0 && w.active < len(w.tabs) && sameRun(w.tabs[w.active].session, s)
	if active {
		prefix = m.tokens.sidebar.active
	}
	fit := func(text string) string {
		text = cellTruncate(text, width, m.tokens.marks.ellipsis)
		return text + strings.Repeat(" ", max(0, width-lipgloss.Width(text)))
	}
	titleStyle, metadataStyle := m.styles.text, m.styles.dim
	if selected || active {
		titleStyle, metadataStyle = m.styles.sidebarSelected, m.styles.sidebarMetadataSelected
	}
	lines := []string{titleStyle.Render(fit(prefix + title))}
	if rowLines > 1 {
		state := s.TerminalState
		if m.listErr != nil {
			state = session.TerminalUnknown
		}
		metadata := terminalLabel(state)
		if command := label(s.Command); command != "" {
			// Display the executable name, not a machine-specific launch path.
			// Both separators are recognized for cross-platform metadata views.
			command = command[strings.LastIndexAny(command, `/\`)+1:]
			metadata += m.tokens.sidebar.separator + command
		}
		lines = append(lines, metadataStyle.Render(fit(m.tokens.sidebar.metadata+metadata)))
	}
	return lines
}
