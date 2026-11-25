// Confirmation layout adapted from Agent Deck, Copyright (c) 2025 Ashesh
// Goplani. SPDX-License-Identifier: MIT
// Source: confirm_dialog.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: complete target review, localized stop-only text and existing tokens.
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) sessionStopBox() (string, bool) {
	d, t := m.sessionStop, m.tokens.confirm
	outer, width := m.groupRemovalWidths()
	title := t.titlePrefix + "Stop session?"
	warning := "Stop this terminal and its program:\n\n" + strings.Repeat(" ", t.nameIndent) + t.nameOpen + label(d.target.Title) + t.nameClose
	details := "Session：" + label(d.target.ID) + "\nRun：" + label(d.target.RunID) + "\nProgram: " + label(d.target.Command) + "\n" + t.bullet + "Unsaved work may be lost\n" + t.bullet + "Tab, name, group and history are preserved"
	button := func(text string, focused, destructive bool) string {
		style, prefix := m.styles.confirmCancel, t.idlePrefix
		if destructive {
			style = m.styles.confirmDelete
		}
		if focused {
			prefix = t.focusedPrefix
			if destructive {
				style = m.styles.confirmDeleteFocused
			} else {
				style = m.styles.confirmCancelFocused
			}
		}
		return style.Render(prefix + text)
	}
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, button("Stop", d.focus == 0, true), strings.Repeat(" ", t.buttonGap), button("Cancel", d.focus == 1, false))
	if lipgloss.Width(buttons) > width {
		buttons = lipgloss.JoinVertical(lipgloss.Left, button("Stop", d.focus == 0, true), button("Cancel", d.focus == 1, false))
	}
	hint := "y Stop · n Cancel · ←/→ Select · Enter · Esc"
	if d.busy {
		hint = "Verifying and stopping the original run…"
	} else if d.submitted {
		hint = "Esc Back and refresh to verify"
	}
	errLine := ""
	if d.err != nil {
		errLine = m.styles.confirmError.Width(width).Render(label(d.err.Error()))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, m.styles.confirmTitle.Width(width).Render(title), m.styles.confirmWarning.Width(width).Render(warning), m.styles.confirmDetails.Width(width).Render(details), errLine, buttons, m.styles.confirmHint.Width(width).Render(hint))
	box := m.styles.confirmBox.Width(outer).Render(content)
	bw, bh := lipgloss.Size(box)
	return box, width > 0 && bw <= m.width && bh <= m.height
}

func (m *Model) renderSessionStop() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	box, fits := m.sessionStopBox()
	if !fits {
		outer, width := m.groupRemovalWidths()
		hint := "Esc Cancel"
		if m.sessionStop.busy {
			hint = "Stopping; wait for the result"
		} else if m.sessionStop.submitted {
			hint = "Outcome unknown · Esc Back and refresh to verify"
		}
		compact := m.styles.confirmDetails.Render("Enlarge the window to review the complete message") + "\n" + m.styles.confirmHint.Render(hint)
		box = m.styles.confirmBox.Width(outer).Render(fitFrame(compact, width, max(0, m.height-m.styles.confirmBox.GetVerticalFrameSize())))
	}
	bw, bh := lipgloss.Size(box)
	return overlayLaunchDialog(fitFrame("", m.width, m.height), box, max(0, (m.width-bw)/2), max(0, (m.height-bh)/2), m.width, m.height)
}
