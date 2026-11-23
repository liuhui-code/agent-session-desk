package ui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
)

// Pinned Bubbles 2.2.2 indexes its placeholder rune buffer using cell width.
// A wide-character placeholder can therefore emit zero runes. Remove only
// those generated display characters; preserve the model, styles and Cursor.
func inputView(input textinput.Model) string {
	view := input.View()
	if input.Value() == "" {
		view = strings.ReplaceAll(view, "\x00", "")
	}
	return view
}
