// Layout adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/confirm_dialog.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: v2 total widths, localized reviewed counts, token styles and bounded
// viewports; no terminal or Session/Run deletion action.
package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) groupRemovalWidths() (outer, content int) {
	t := m.tokens.confirm
	outer = t.preferredWidth
	if m.width < t.responsiveBelow {
		outer = m.width - t.narrowGutter
	}
	outer = max(1, min(m.width, outer))
	content = max(0, outer-m.styles.confirmBox.GetHorizontalFrameSize())
	return
}

func (m *Model) groupRemovalBox() (string, bool) {
	d, t := m.groupRemoval, m.tokens.confirm
	outer, width := m.groupRemovalWidths()
	name := d.target.Name
	if d.plan.Root.ID != "" {
		name = d.plan.Root.Name
	}
	title := t.titlePrefix + "Delete group?"
	warning := "Delete this group:\n\n" + strings.Repeat(" ", t.nameIndent) + t.nameOpen + label(name) + t.nameClose
	details := "Checking groups and members…"
	if !d.loading && d.plan.Stamp == "" && d.err != nil {
		details = "Cannot verify impact; go back and refresh."
	}
	if d.plan.Stamp != "" {
		details = fmt.Sprintf("%sRemove %d groups, including subgroups\n%sMove %d sessions to%s%s%s\n%sTerminals, output and history are preserved", t.bullet, len(d.plan.Groups), t.bullet, d.plan.MemberCount, t.nameOpen, label(d.plan.Destination.Name), t.nameClose, t.bullet)
	}
	button := func(text string, focused, destructive bool) string {
		style := m.styles.confirmCancel
		prefix := t.idlePrefix
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
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, button("Delete", d.focus == 0, true), strings.Repeat(" ", t.buttonGap), button("Cancel", d.focus == 1, false))
	if lipgloss.Width(buttons) > width {
		buttons = lipgloss.JoinVertical(lipgloss.Left, button("Delete", d.focus == 0, true), button("Cancel", d.focus == 1, false))
	}
	hint := "y Delete · n Cancel · ←/→ Select · Enter · Esc"
	if d.loading {
		hint = "Reviewing impact · Esc Cancel"
	} else if d.busy {
		hint = "Moving members and removing groups…"
	} else if d.submitted {
		hint = "Esc Back and refresh to verify"
	}
	titleStyle, boxStyle := m.styles.confirmTitle, m.styles.confirmBox
	if d.target.ID == session.DefaultGroupID {
		title = t.titlePrefix + "Cannot delete the default group"
		warning = t.nameOpen + label(name) + t.nameClose + " is the default group and cannot be deleted.\n\nSessions must belong to a group."
		details = ""
		buttons = button("Confirm", true, false)
		hint = "Enter / Esc / o Back"
		titleStyle, boxStyle = m.styles.confirmNoticeTitle, m.styles.confirmNoticeBox
	}
	errLine := ""
	if d.err != nil {
		errLine = m.styles.confirmError.Width(width).Render(label(d.err.Error()))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, titleStyle.Width(width).Render(title), m.styles.confirmWarning.Width(width).Render(warning), m.styles.confirmDetails.Width(width).Render(details), errLine, buttons, m.styles.confirmHint.Width(width).Render(hint))
	box := boxStyle.Width(outer).Render(content)
	bw, bh := lipgloss.Size(box)
	return box, bw <= m.width && bh <= m.height && width > 0
}

func (m *Model) renderGroupRemoval() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	box, fits := m.groupRemovalBox()
	if !fits {
		outer, width := m.groupRemovalWidths()
		compact := m.styles.confirmDetails.Render("Enlarge the window to review the complete message") + "\n" + m.styles.confirmHint.Render("Esc Cancel")
		box = m.styles.confirmBox.Width(outer).Render(fitFrame(compact, width, max(0, m.height-m.styles.confirmBox.GetVerticalFrameSize())))
	}
	bw, bh := lipgloss.Size(box)
	return overlayLaunchDialog(fitFrame("", m.width, m.height), box, max(0, (m.width-bw)/2), max(0, (m.height-bh)/2), m.width, m.height)
}
