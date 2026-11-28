package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) recoveryDetailLines(d *recoveryCenter) []string {
	r := d.review
	lines := []string{"Session: " + r.Source.ID, "Previous Run: " + r.Source.RunID, "Terminal liveness is not inferred from saved metadata"}
	if r.Profile == nil {
		lines = append(lines, "No Agent recovery declaration")
	} else {
		p := r.Profile
		lines = append(lines, fmt.Sprintf("Agent: %s · Profile revision: %d", label(p.Declaration.AgentID), p.Revision), "Conversation: "+label(p.Declaration.ConversationID))
	}
	if r.Attempt != nil {
		lines = append(lines, "Latest recovery attempt: "+label(r.Attempt.State))
	}
	if r.Launch == nil {
		lines = append(lines, "No complete original launch snapshot")
	} else {
		lines = append(lines, "Original launch available for a new Session; does not restore a conversation")
	}
	if d.confirm && r.Profile != nil {
		command, _ := json.Marshal(r.Profile.Declaration.Resume)
		lines = append(lines, "Private recovery command review: "+label(string(command)), "A new Run and fresh credential will be created; no queued input is replayed", "I confirm the previous Run ended")
	}
	return strings.Split(ansi.Hardwrap(strings.Join(lines, "\n"), max(1, m.diagnosticContentWidth()), true), "\n")
}
func (m *Model) renderRecoveryCenter() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	d := m.recoveryUI.center
	t := m.tokens.diagnostic
	outer := min(m.width, t.width)
	if m.width < t.width {
		outer = max(0, m.width-t.gutter)
	}
	width := m.diagnosticContentWidth()
	height := min(t.rows, max(0, m.height-m.styles.searchBox.GetVerticalFrameSize()))
	fixed := 3
	if d.confirm {
		fixed = 5
	}
	page := max(0, height-fixed)
	lines := []string{m.styles.searchTitle.Render(cellTruncate("Recover sessions", width, m.tokens.marks.ellipsis))}
	if d.detail {
		body := m.recoveryDetailLines(d)
		offset := min(d.offset, max(0, len(body)-max(1, page)))
		if page > 0 {
			lines = append(lines, m.styles.text.Render(strings.Join(body[offset:min(len(body), offset+page)], "\n")))
		}
	} else {
		start := max(0, d.cursor-max(1, page)+1)
		for i := start; i < min(len(d.rows), start+page); i++ {
			row := d.rows[i]
			text := fmt.Sprintf("%s · Run %s", label(row.Title), row.RunID)
			style := m.styles.searchResult
			if i == d.cursor {
				style = m.styles.searchSelected
			}
			lines = append(lines, style.Render(cellTruncate(text, max(0, width-style.GetHorizontalFrameSize()), m.tokens.marks.ellipsis)))
		}
		if len(d.rows) == 0 && page > 0 {
			lines = append(lines, m.styles.dim.Render("No recorded sessions"))
		}
	}
	status := d.message
	if d.busy {
		status = "Loading…"
		if d.submitted {
			status = "Starting recovery… no automatic retry"
		}
	}
	lines = append(lines, m.styles.searchTip.Render(cellTruncate(status, width, m.tokens.marks.ellipsis)))
	hint := "Enter Inspect · a Reattach · n New Session · r Refresh/Restore · Esc Back"
	if d.confirm {
		checked := t.unchecked
		if d.ended {
			checked = t.checked
		}
		choice := t.actionOpen + "Cancel" + t.actionClose + t.actionSeparator + "Start recovery"
		if d.startSelected {
			choice = "Cancel" + t.actionSeparator + t.actionOpen + "Start recovery" + t.actionClose
		}
		lines = append(lines, m.styles.text.Render(cellTruncate(checked+" Previous Run ended", width, m.tokens.marks.ellipsis)), m.styles.searchSelected.Render(cellTruncate(choice, max(0, width-m.styles.searchSelected.GetHorizontalFrameSize()), m.tokens.marks.ellipsis)))
		hint = "Space Confirm ended · Tab Select · Enter · Esc"
	}
	lines = append(lines, m.styles.searchHint.Render(cellTruncate(hint, width, m.tokens.marks.ellipsis)))
	box := m.styles.searchBox.Width(outer).Render(fitFrame(strings.Join(lines, "\n"), width, height))
	if lipgloss.Width(box) > m.width || lipgloss.Height(box) > m.height {
		box = fitFrame(box, m.width, m.height)
	}
	return overlayLaunchDialog(fitFrame("", m.width, m.height), box, max(0, (m.width-lipgloss.Width(box))/2), max(0, (m.height-lipgloss.Height(box))/2), m.width, m.height)
}
