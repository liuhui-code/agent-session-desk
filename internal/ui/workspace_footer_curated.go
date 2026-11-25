// Curated footer formatting and greedy fitting selectively adapted from
// Agent Deck home_view_frame.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0,
// Copyright (c) 2025 Ashesh Goplani. SPDX-License-Identifier: MIT.
// Modified: token styles, localized supported terminal actions and product
// Session/Run availability. See third_party/agent-deck/LICENSE.
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) curatedFooterHints() []footerHint {
	var hints []footerHint
	if m.sidebarGroupSelected() {
		action := "Collapse"
		if m.groupCollapsed(m.selectedGroup().ID) {
			action = "Expand"
		}
		hints = append(hints, footerHint{"Tab", action}, footerHint{"n", "New"}, footerHint{"r", "Rename"})
	} else if selected, ok := m.selected(); ok {
		if selected.TerminalState == session.TerminalAvailable {
			hints = append(hints, footerHint{"Enter", "Open"})
		}
		name := "Rename"
		if m.workspace.err != nil {
			name = "Retry"
		}
		hints = append(hints, footerHint{"n", "New"}, footerHint{"r", name})
	} else {
		hints = append(hints, footerHint{"n", "New"})
		if _, ok := m.service.(groupService); ok {
			hints = append(hints, footerHint{"g", "Groups"})
		}
	}
	return hints[:min(len(hints), m.tokens.footer.curatedHintLimit)]
}

func (m *Model) renderCuratedFooter() string {
	t := m.tokens.footer
	sep := m.styles.footerSeparator.Render(strings.Repeat(" ", t.curatedGap))
	render := func(h footerHint) string {
		return m.styles.footerCuratedKey.Render(h.key) + strings.Repeat(" ", t.labelGap) + m.styles.dim.Render(h.label)
	}
	globals := []string{render(footerHint{"S", "Settings"}), render(footerHint{"?", "Help"})}
	used := lipgloss.Width(strings.Join(globals, sep))
	var parts []string
	for _, hint := range m.curatedFooterHints() {
		text := render(hint)
		next := used + t.curatedGap + lipgloss.Width(text)
		if next > m.width {
			break
		}
		parts, used = append(parts, text), next
	}
	return strings.Join(append(parts, globals...), sep)
}
