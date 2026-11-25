package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) renderHistory(_ int) string {
	h := &m.history
	title := m.renderPanelTitle(fmt.Sprintf("History · Latest %d records", historyLimit), m.width)
	if h.busy {
		return title + "\n" + m.styles.dim.Render("Loading history…")
	}
	if h.err != nil {
		return title + "\n" + m.styles.error.Render("Failed to load history: "+label(h.err.Error()))
	}
	if len(h.rows) == 0 {
		return title + "\n" + m.styles.dim.Render("No history records")
	}
	if h.detail {
		lines := m.historyDetailLines()
		offset := min(h.offset, max(0, len(lines)-m.historyPageSize()))
		return title + "\n" + m.styles.text.Render(strings.Join(lines[offset:min(len(lines), offset+m.historyPageSize())], "\n"))
	}
	lines := []string{}
	size := m.historyPageSize()
	start := (h.cursor / size) * size
	for i := start; i < min(len(h.rows), start+size); i++ {
		r := h.rows[i]
		prefix := strings.Repeat(" ", m.tokens.layout.rowPrefixWidth)
		if i == h.cursor {
			prefix = m.tokens.marks.selected + strings.Repeat(" ", max(0, m.tokens.layout.rowPrefixWidth-lipgloss.Width(m.tokens.marks.selected)))
		}
		row := cellTruncate(prefix+label(r.Session.Title), m.width, m.tokens.marks.ellipsis)
		if i == h.cursor {
			row = m.styles.highlight.Render(row)
		} else {
			row = m.styles.text.Render(row)
		}
		lines = append(lines, row)
	}
	return title + "\n" + strings.Join(lines, "\n")
}

// Each metadata value becomes inert, bounded text before display-cell wrapping.
func (m *Model) historyDetailLines() []string {
	h := &m.history
	if len(h.rows) == 0 {
		return nil
	}
	r := h.rows[h.cursor]
	fields := []string{"Read-only record · Current online state unknown", "Name: " + label(r.Session.Title), "Session：" + label(r.Session.ID), "Run：" + label(r.Session.RunID), "Backend：" + label(r.Session.BackendID), "Program: " + label(r.Session.Command), "First recorded: " + r.RecordedAt.UTC().Format(time.RFC3339Nano), "Agent result: unknown"}
	stop := "Not recorded (does not imply still running)"
	if r.StopConfirmedAt != nil {
		stop = r.StopConfirmedAt.UTC().Format(time.RFC3339Nano)
	}
	fields = append(fields, "Confirm stop: "+stop)
	if r.PaneExit == nil {
		fields = append(fields, "Pane exit observation: not recorded")
	} else {
		fields = append(fields, "pane："+label(r.PaneExit.PaneID), "Observation source: "+label(r.PaneExit.Source), "Exit observation time: "+r.PaneExit.ObservedAt.UTC().Format(time.RFC3339Nano))
	}
	fields = append(fields, "Exit code: unknown")
	return strings.Split(ansi.Hardwrap(strings.Join(fields, "\n"), max(1, m.width), true), "\n")
}

func (m *Model) renderHistoryHelpBar() string {
	action := "Select"
	if m.history.detail {
		action = "Scroll"
	}
	parts := []string{m.styles.menuKey("↑↓/PgUp/PgDown", action)}
	if !m.history.detail {
		parts = append(parts, m.styles.menuKey("Enter", "Details"))
	}
	parts = append(parts, m.styles.menuKey("r", "Refresh"), m.styles.menuKey("Esc/h", "Back"), m.styles.menuKey("q", "Quit"))
	return strings.Join(parts, m.tokens.itemGap())
}
