package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

type attentionCenterGeometry struct{ x, y, width, content, height, start, visible, rowY int }

func (m *Model) attentionCenterGeometry(count, selected int) attentionCenterGeometry {
	t := m.tokens.attention
	w := min(m.width, t.centerWidth)
	if m.width < t.centerWidth {
		w = max(0, m.width-t.centerGutter)
	}
	w = max(0, min(m.width, w))
	content := max(0, w-m.styles.searchBox.GetHorizontalFrameSize())
	height := max(0, m.height-m.styles.searchBox.GetVerticalFrameSize())
	// Title/filter and status/help are fixed structure rows, not extra caches.
	visible := min(count, t.centerRows, max(0, height-4))
	body := min(height, visible+4)
	boxHeight := body + m.styles.searchBox.GetVerticalFrameSize()
	x, y := max(0, (m.width-w)/2), max(0, (m.height-boxHeight)/2)
	return attentionCenterGeometry{x: x, y: y, width: w, content: content, height: body, visible: visible, start: max(0, selected-visible+1), rowY: y + m.styles.searchBox.GetBorderTopSize() + m.tokens.search.paddingY + 2}
}
func (m *Model) renderAttentionCenter() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	now := time.Now()
	rows := m.attentionCenterRows(now)
	selected := m.attentionCenterSelection(rows)
	d := m.attention.center
	g := m.attentionCenterGeometry(len(rows), selected)
	filters := []string{"All", "Due", "Snoozed"}
	lines := []string{m.styles.searchTitle.Render("Pending items"), m.styles.searchCount.Render(fmt.Sprintf("%s · %d items · Tab Filter", filters[d.filter], len(rows)))}
	for row := g.start; row < min(len(rows), g.start+g.visible); row++ {
		item := m.attention.items[rows[row]]
		status := attentionWait(item, now)
		if !item.TerminalAvailable {
			status += " · unavailable"
		}
		text := fmt.Sprintf("%s · %s · %s · %s", label(item.Source.Title), item.Kind, label(item.Title), status)
		style, prefix := m.styles.searchResult, m.tokens.attention.idle
		if row == selected {
			style, prefix = m.styles.searchSelected, m.tokens.attention.marker
		}
		lines = append(lines, style.Render(fitFrame(prefix+text, max(0, g.content-style.GetHorizontalFrameSize()), 1)))
	}
	status := "Enter opens the original prompt; viewing never approves"
	if len(rows) == 0 {
		status = "No pending items in this filter"
	}
	if d.busy {
		status = "Saving reminder…"
	} else if d.err != "" {
		status = d.err
	} else if m.listErr != nil {
		status = "Last observed; refresh before acting"
	}
	lines = append(lines, m.styles.searchTip.Render(cellTruncate(status, g.content, m.tokens.marks.ellipsis)))
	help := "Enter Open · s 10m · S 30m · u Resume · Tab Filter · Esc Back"
	if g.content < m.tokens.attention.helpBelow {
		help = "↵ Open · s/S Later · u Resume · Esc"
	}
	lines = append(lines, m.styles.searchHint.Render(cellTruncate(help, g.content, m.tokens.marks.ellipsis)))
	box := m.styles.searchBox.Width(g.width).Render(fitFrame(strings.Join(lines, "\n"), g.content, g.height))
	// Keep tiny windows bounded even when the existing border cannot fit.
	if lipgloss.Width(box) > m.width {
		box = fitFrame(box, m.width, min(m.height, lipgloss.Height(box)))
	}
	return overlayLaunchDialog(fitFrame("", m.width, m.height), box, g.x, g.y, m.width, m.height)
}
