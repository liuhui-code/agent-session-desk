package ui

import (
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// SetWidth alone does not reflow the pinned input's scrolling window. Revisit
// both ends of that window while keeping the value and logical caret intact.
// Width excludes the prompt and the one trailing real-cursor cell.
func sizeInput(input *textinput.Model, width int) {
	position := input.Position()
	input.SetWidth(max(1, width-lipgloss.Width(input.Prompt)-1))
	input.CursorEnd()
	input.CursorStart()
	input.SetCursor(position)
}

// Pinned Bubbles 2.2.2 indexes its placeholder rune buffer using cell width.
// A wide-character placeholder can therefore emit zero runes. Remove only
// those generated display characters; preserve the model, styles and Cursor.
func inputView(input textinput.Model) string {
	view, _ := inputFrame(input)
	return view
}

func inputCursor(input textinput.Model) *tea.Cursor {
	_, cursor := inputFrame(input)
	return cursor
}

func inputFrame(input textinput.Model) (string, *tea.Cursor) {
	// Pinned Bubbles measures keycaps as one cell, unlike the two-cell
	// grapheme measurement used by the product renderer. Project only this
	// affected display through whole clusters; never rewrite Value/Position.
	if strings.ContainsRune(input.Value(), 0x20E3) && input.Width() > 0 {
		return keycapInputFrame(input)
	}
	view := input.View()
	if input.Value() == "" {
		view = strings.ReplaceAll(view, "\x00", "")
	}
	return view, input.Cursor()
}

func keycapInputFrame(input textinput.Model) (string, *tea.Cursor) {
	type cluster struct {
		text  string
		width int
	}
	var clusters []cluster
	caret, runes, partial := 0, 0, 0
	for rest := input.Value(); rest != ""; {
		text, _ := ansi.FirstGraphemeCluster(rest, ansi.GraphemeWidth)
		clusters = append(clusters, cluster{text, ansi.StringWidth(text)})
		if remaining := input.Position() - runes; remaining > 0 && remaining < utf8.RuneCountInString(text) {
			partial = ansi.StringWidth(string([]rune(text)[:remaining]))
		}
		runes += utf8.RuneCountInString(text)
		if runes <= input.Position() {
			caret++
		}
		rest = rest[len(text):]
	}
	// The pinned editor uses rune insertion positions. A caret inside a
	// cluster must retain that prefix width instead of snapping to its start.
	start, x := caret, min(input.Width(), partial)
	for start > 0 && x+clusters[start-1].width <= input.Width() {
		start--
		x += clusters[start].width
	}
	var text strings.Builder
	used := 0
	for i := start; i < len(clusters) && used+clusters[i].width <= input.Width()+1; i++ {
		text.WriteString(clusters[i].text)
		used += clusters[i].width
	}
	text.WriteString(strings.Repeat(" ", max(0, input.Width()+1-used)))
	styles := input.Styles()
	active := styles.Blurred
	if input.Focused() {
		active = styles.Focused
	}
	view := active.Prompt.Inline(true).Render(input.Prompt) + active.Text.Inline(true).Render(text.String())
	cursor := input.Cursor()
	if cursor != nil {
		cursor.X = lipgloss.Width(input.Prompt) + x
	}
	return view, cursor
}
