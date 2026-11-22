package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) workspaceCopyPrompt() (string, *tea.Cursor) {
	w := m.workspace
	prompt := w.screen.CopyPrompt
	if !w.focused || w.conn == nil || w.err != nil || !w.screen.CopyMode || prompt == nil || m.width < 1 || m.height < m.tokens.workspaceChromeRows() {
		return "", nil
	}
	if prompt.LabelBytes < 0 || prompt.CursorBytes < prompt.LabelBytes || prompt.CursorBytes > len(prompt.Text) {
		return "", nil
	}
	prefix := cellTruncate(prompt.Text[:prompt.LabelBytes], m.width-1, "")
	input := prompt.Text[prompt.LabelBytes:]
	prefixWidth := lipgloss.Width(prefix)
	available := m.width - prefixWidth
	at := ansi.StringWidth(input[:prompt.CursorBytes-prompt.LabelBytes])
	start := max(0, at-available+1)
	text := prefix + ansi.Cut(input, start, start+available)
	cursor := tea.NewCursor(prefixWidth+at-start, m.height-1-m.tokens.workspace.footerBottomGapRows)
	cursor.Shape, cursor.Blink = m.tokens.dialog.cursorShape, m.tokens.dialog.cursorBlink
	return text, cursor
}
