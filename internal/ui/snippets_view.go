package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) sizeSnippets() {
	_, _, width := m.sessionSearchWidths()
	sizeInput(&m.snippets.dialog.input, width-m.styles.searchInput.GetHorizontalFrameSize())
}
func (m *Model) renderSnippets() (string, *tea.Cursor) {
	if m.width <= 0 || m.height <= 0 {
		return "", nil
	}
	if m.width < m.tokens.snippets.minWidth || m.height < m.tokens.snippets.minHeight {
		return fitFrame(m.styles.searchHint.Render("Enlarge window · Esc Cancel"), m.width, m.height), nil
	}
	d, t := m.snippets.dialog, m.tokens.search
	width, content, inputWidth := m.sessionSearchWidths()
	maxRows := max(0, m.height-m.styles.searchBox.GetVerticalFrameSize())
	lines := []string{m.styles.searchTitle.Render("Instruction snippets"), m.styles.dim.Render("Target: " + label(d.target.Title))}
	if d.preview {
		item := d.library.Items[d.rows[d.cursor]]
		lines = append(lines, m.styles.searchTitle.Render(item.Source+" · "+item.Title))
		if maxRows < len(lines)+4 {
			lines = lines[1:]
		}
		body := strings.ReplaceAll(item.Body, "\r\n", "\n")
		body = strings.ReplaceAll(body, "\r", "\n")
		body = strings.ReplaceAll(body, "\t", m.tokens.snippets.tabSpaces)
		wrapped := strings.Split(ansi.Hardwrap(body, max(1, content), true), "\n")
		page := max(0, min(m.tokens.snippets.previewRows, maxRows-len(lines)-3))
		offset := min(d.offset, max(0, len(wrapped)-max(1, page)))
		if page > 0 {
			lines = append(lines, m.styles.text.Render(strings.Join(wrapped[offset:min(len(wrapped), offset+page)], "\n")))
		}
		warning, hint := "Multiline paste may execute in shells", "Enter Paste · ↑↓ Scroll · Esc Back"
		if !strings.ContainsAny(item.Body, "\r\n") {
			warning = "No Enter added"
		}
		if lipgloss.Width(warning) > content {
			warning = "Newlines may execute"
		}
		if lipgloss.Width(hint) > content {
			hint = "Enter Paste · Esc Back"
		}
		lines = append(lines, m.styles.searchTip.Render(warning), m.styles.searchHint.Render(hint))
	} else {
		lines = append(lines, m.styles.searchInput.Width(inputWidth).Render(inputView(d.input)))
		headRows := lipgloss.Height(strings.Join(lines, "\n"))
		if maxRows < headRows+3 {
			lines = lines[2:]
			headRows = lipgloss.Height(strings.Join(lines, "\n"))
		}
		page := max(0, min(t.maxResults, maxRows-headRows-2))
		start := max(0, d.cursor-page+1)
		for i := start; i < min(len(d.rows), start+page); i++ {
			item := d.library.Items[d.rows[i]]
			style := m.styles.searchResult
			prefix := strings.Repeat(" ", lipgloss.Width(t.selected))
			if i == d.cursor {
				style = m.styles.searchSelected
				prefix = t.selected
			}
			text := prefix + item.Source + " · " + item.Title
			lines = append(lines, style.Render(fitFrame(cellTruncate(text, max(0, content-style.GetHorizontalFrameSize()), m.tokens.marks.ellipsis), max(0, content-style.GetHorizontalFrameSize()), 1)))
		}
		count := fmt.Sprintf("%d snippets · Project then Global", len(d.rows))
		if d.library.ProjectDir == "" {
			count = "Global snippets only; project directory unavailable"
		}
		if d.loading {
			count = "Loading snippets…"
		}
		if !d.loading && len(d.rows) == 0 && d.notice == "" {
			count = "No snippets. Configure snippets.toml; Esc closes"
		}
		hint := "Enter Preview · ↑↓ Select · Esc Cancel"
		if lipgloss.Width(hint) > content {
			hint = "Enter Preview · Esc Cancel"
		}
		lines = append(lines, m.styles.searchCount.Render(count), m.styles.searchHint.Render(hint))
	}
	if d.notice != "" {
		lines = append(lines, m.styles.error.Render(label(d.notice)))
	}
	// Each authored structural line is clipped; the complete frame is bounded.
	flat := strings.Split(strings.Join(lines, "\n"), "\n")
	for i, s := range flat {
		flat[i] = cellTruncate(s, content, m.tokens.marks.ellipsis)
	}
	box := m.styles.searchBox.Width(width).Render(fitFrame(strings.Join(flat, "\n"), content, min(maxRows, len(flat))))
	bw, bh := lipgloss.Size(box)
	x, y := max(0, (m.width-bw)/2), max(0, (m.height-bh)/2)
	out := overlayLaunchDialog(fitFrame("", m.width, m.height), box, x, y, m.width, m.height)
	cursor := inputCursor(d.input)
	if d.preview || d.loading {
		cursor = nil
	}
	if cursor != nil {
		cursor.X += x + m.styles.searchBox.GetBorderLeftSize() + t.paddingX + m.styles.searchInput.GetBorderLeftSize() + t.inputPaddingX
		inputStart := 2
		if !d.preview && maxRows < 2+lipgloss.Height(m.styles.searchInput.Width(inputWidth).Render(inputView(d.input)))+3 {
			inputStart = 0
		}
		cursor.Y += y + m.styles.searchBox.GetBorderTopSize() + t.paddingY + inputStart + m.styles.searchInput.GetBorderTopSize() + t.inputPaddingY
		if cursor.X >= m.width || cursor.Y >= y+bh-m.styles.searchBox.GetBorderBottomSize() {
			cursor = nil
		}
	}
	return out, cursor
}
