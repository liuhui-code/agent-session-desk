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
	"github.com/charmbracelet/x/ansi"
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
	rows := m.tokens.sidebar.cardRows
	switch m.sidebarDensity {
	case "full":
		rows = m.tokens.sidebar.fullRows
	case "minimal":
		rows = 1
	case "auto":
		rows = 1
		for candidate := m.tokens.sidebar.fullRows; candidate > 1; candidate-- {
			if len(m.sessions)*candidate <= height {
				rows = candidate
				break
			}
		}
	}
	return max(1, min(height, rows))
}

func (m *Model) workspaceSidebarCard(index, width, rowLines int) []string {
	s := m.sessions[index]
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
	title := m.sessionTitle(s, m.sessions, max(0, width-lipgloss.Width(prefix)))
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
		count := rowLines - 1
		metadataWidth := max(1, width-lipgloss.Width(m.tokens.sidebar.metadata))
		wrapped := strings.Split(ansi.Wrap(metadata, metadataWidth, ""), "\n")
		if len(wrapped) > count {
			wrapped[count-1] = cellTruncate(strings.Join(wrapped[count-1:], " "), metadataWidth, m.tokens.marks.ellipsis)
			wrapped = wrapped[:count]
		}
		for i := 0; i < count; i++ {
			prefix, text := m.tokens.sidebar.metadataMiddle, ""
			if i == count-1 {
				prefix = m.tokens.sidebar.metadata
			}
			if i < len(wrapped) {
				text = wrapped[i]
			}
			lines = append(lines, metadataStyle.Render(fit(prefix+text)))
		}
	}
	return lines
}
