package ui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) messageDeliveryBox() (string, bool) {
	d, t := m.workspace.delivery, m.tokens.confirm
	outer, width := m.groupRemovalWidths()
	title := "Review queued message"
	details := "Checking pending messages…"
	preview := ""
	hint := "Esc Cancel"
	if !d.loading && d.err == nil {
		details = fmt.Sprintf("To: %s\nSession: %s\nRun: %s\nMessage: %s", label(d.target.Title), d.target.ID, d.target.RunID, d.message.ID)
		if s := d.message.Source; s != nil {
			details += fmt.Sprintf("\nFrom Session: %s\nSource Run: %s\nSnapshot: %s · lines %d-%d", s.SessionID, s.RunID, s.RecordID, s.StartLine, s.EndLine)
		}
		rows := m.messageDeliveryLines()
		page := m.tokens.workspace.deliveryPreviewRows
		offset := min(d.offset, max(0, len(rows)-page))
		preview = fmt.Sprintf("Escaped preview %d-%d / %d\n", offset+1, min(len(rows), offset+page), len(rows)) + strings.Join(rows[offset:min(len(rows), offset+page)], "\n")
		hint = "Tab Select · Enter · ↑/↓ Scroll · Esc Cancel"
	}
	warning := "Paste only; no Enter is appended.\nMultiline paste may execute in some shells. Review the target program; agent readiness is unknown."
	if d.busy {
		hint = "Writing once through the original input owner; wait…"
	}
	if d.submitted {
		title = "Message input result"
		details = "State: " + label(d.message.State) + "\nMessage: " + label(d.message.ID)
		hint = "Esc Back"
		warning = "Input written is not agent acceptance or completion."
	}
	errLine := ""
	if d.err != nil {
		errLine = label(d.err.Error())
		hint = "Esc Back"
	}
	button := func(text string, focus bool) string {
		style, prefix := m.styles.confirmCancel, t.idlePrefix
		if focus {
			style, prefix = m.styles.confirmCancelFocused, t.focusedPrefix
		}
		return style.Render(prefix + text)
	}
	buttons := ""
	if !d.loading && !d.busy && !d.submitted && d.err == nil {
		buttons = lipgloss.JoinHorizontal(lipgloss.Top, button("Paste", d.focus == 0), strings.Repeat(" ", t.buttonGap), button("Cancel", d.focus == 1))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, m.styles.confirmTitle.Width(width).Render(title), m.styles.confirmDetails.Width(width).Render(details), m.styles.confirmWarning.Width(width).Render(warning), m.styles.confirmDetails.Width(width).Render(preview), m.styles.confirmError.Width(width).Render(errLine), buttons, m.styles.confirmHint.Width(width).Render(hint))
	box := m.styles.confirmBox.Width(outer).Render(content)
	bw, bh := lipgloss.Size(box)
	return box, width > 0 && bw <= m.width && bh <= m.height
}
func (m *Model) renderMessageDelivery() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	box, fits := m.messageDeliveryBox()
	if !fits {
		outer, width := m.groupRemovalWidths()
		hint := "Enlarge the window to review the complete target · Esc Cancel"
		if m.workspace.delivery.busy {
			hint = "Writing; wait for result"
		}
		box = m.styles.confirmBox.Width(outer).Render(fitFrame(m.styles.confirmDetails.Render(hint), width, max(0, m.height-m.styles.confirmBox.GetVerticalFrameSize())))
	}
	bw, bh := lipgloss.Size(box)
	return overlayLaunchDialog(fitFrame("", m.width, m.height), box, max(0, (m.width-bw)/2), max(0, (m.height-bh)/2), m.width, m.height)
}

func (m *Model) messageDeliveryLines() []string {
	_, width := m.groupRemovalWidths()
	lines := strings.Split(m.workspace.delivery.message.Body, "\n")
	for i, line := range lines {
		quoted := strconv.QuoteToGraphic(line)
		lines[i] = quoted[1 : len(quoted)-1]
	}
	return strings.Split(ansi.Wrap(strings.Join(lines, "\n"), max(1, width), ""), "\n")
}
