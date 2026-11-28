package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) diagnosticPageSize() int {
	t := m.tokens.diagnostic
	return max(1, min(t.rows, max(0, m.height-m.styles.searchBox.GetVerticalFrameSize()))-3)
}
func (m *Model) diagnosticLines() []string {
	r := m.diagnosticReport()
	revision := r.Runtime.Revision
	if revision == "" {
		revision = "unknown"
	}
	lines := []string{"Version: " + r.Runtime.ProductVersion + " · Build: " + revision, fmt.Sprintf("Modified: %t · %s · %s/%s", r.Runtime.Dirty, r.Runtime.Go, r.Runtime.OS, r.Runtime.Arch), fmt.Sprintf("Backend: %s %s", r.Runtime.Backend, r.Runtime.BackendVersion)}
	if m.diagnostic.target.ID != "" {
		lines = append(lines, "Selected: "+label(m.diagnostic.target.Title), "Reference: "+r.Target)
	}
	lines = append(lines, m.diagnosticSummary())
	for _, category := range []string{"discovery", "connect", "failure"} {
		found := false
		for i := len(r.Events) - 1; i >= 0; i-- {
			e := r.Events[i]
			match := e.Operation == category
			if category == "connect" {
				match = match && e.Target == r.Target
			}
			if category == "failure" {
				match = e.Cause != "ok" && e.Cause != "cancelled" && (e.Target == r.Target || e.Target == "")
			}
			if match {
				lines = append(lines, fmt.Sprintf("Last %s: %s · %s · %dms", category, e.Operation, e.Cause, e.ElapsedMS))
				found = true
				break
			}
		}
		if !found {
			lines = append(lines, "Last "+category+": none in retained observations")
		}
	}
	lines = append(lines, "Recent operation receipts (UTC; elapsed includes scheduling):")
	if len(r.Events) == 0 {
		lines = append(lines, "No observations in this workbench yet")
	}
	for i := len(r.Events) - 1; i >= 0; i-- {
		e := r.Events[i]
		lines = append(lines, fmt.Sprintf("%s %s · %s · %dms · %s", e.At.UTC().Format(time.TimeOnly), e.Operation, e.Cause, e.ElapsedMS, e.Target))
	}
	if m.diagnostic.exportPath != "" {
		lines = append([]string{"Local report path: " + label(m.diagnostic.exportPath)}, lines...)
	}
	width := max(1, m.diagnosticContentWidth())
	return strings.Split(ansi.Hardwrap(strings.Join(lines, "\n"), width, true), "\n")
}
func (m *Model) diagnosticContentWidth() int {
	t := m.tokens.diagnostic
	width := min(m.width, t.width)
	if m.width < t.width {
		width = max(0, m.width-t.gutter)
	}
	return max(0, width-m.styles.searchBox.GetHorizontalFrameSize())
}
func (m *Model) renderDiagnostics() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	t := m.tokens.diagnostic
	outer := min(m.width, t.width)
	if m.width < t.width {
		outer = max(0, m.width-t.gutter)
	}
	width := m.diagnosticContentWidth()
	height := min(t.rows, max(0, m.height-m.styles.searchBox.GetVerticalFrameSize()))
	lines := m.diagnosticLines()
	page := max(0, height-3)
	offset := min(m.diagnostic.offset, max(0, len(lines)-max(1, page)))
	content := []string{m.styles.searchTitle.Render(cellTruncate("Diagnostics", width, m.tokens.marks.ellipsis))}
	if page > 0 {
		content = append(content, m.styles.text.Render(strings.Join(lines[offset:min(len(lines), offset+page)], "\n")))
	}
	status := m.diagnostic.message
	if status == "" {
		status = "Reports exclude terminal content, raw errors, commands, paths and credentials"
	}
	content = append(content, m.styles.searchTip.Render(cellTruncate(status, width, m.tokens.marks.ellipsis)), m.styles.searchHint.Render(cellTruncate("↑↓/PgUp/PgDown Scroll · E Export · Esc Back", width, m.tokens.marks.ellipsis)))
	box := m.styles.searchBox.Width(outer).Render(fitFrame(strings.Join(content, "\n"), width, height))
	if lipgloss.Width(box) > m.width || lipgloss.Height(box) > m.height {
		box = fitFrame(box, m.width, m.height)
	}
	return overlayLaunchDialog(fitFrame("", m.width, m.height), box, max(0, (m.width-lipgloss.Width(box))/2), max(0, (m.height-lipgloss.Height(box))/2), m.width, m.height)
}
