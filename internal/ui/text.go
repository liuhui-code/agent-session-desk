package ui

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// These are defensive bounds for this experimental preview, not measured
// resource or compatibility promises. Only the selected snapshot is retained.
const (
	maxPreviewBytes = 64 * 1024
	maxPreviewLines = 200
	maxLabelBytes   = 4096
)

// plainText turns captured output into inert display text. ANSI from an agent
// must not be replayed into the manager (OSC clipboard/title, cursor movement,
// alternate-screen controls, BEL, etc.). The backend still owns the live VT
// terminal; the preview deliberately does not emulate it.
func plainText(s string) string {
	s = ansi.Strip(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\n' || r == '\t':
			return r
		case unicode.IsControl(r):
			return -1
		default:
			return r
		}
	}, s)
}

func label(s string) string {
	if len(s) > maxLabelBytes {
		s = s[:maxLabelBytes]
		for !utf8.ValidString(s) && len(s) > 0 {
			s = s[:len(s)-1]
		}
	}
	s = plainText(s)
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(s, "\t", " ")
}

func boundedPreview(s string) (string, bool) {
	truncated := len(s) > maxPreviewBytes
	if truncated {
		s = s[len(s)-maxPreviewBytes:]
		for len(s) > 0 && !utf8.RuneStart(s[0]) {
			s = s[1:]
		}
	}
	s = strings.ReplaceAll(plainText(s), "\t", strings.Repeat(" ", previewTabWidth))
	// Tabs and replacement runes can expand after decoding. Bound the retained
	// display text as well as the untrusted source bytes.
	if len(s) > maxPreviewBytes {
		s = s[len(s)-maxPreviewBytes:]
		for len(s) > 0 && !utf8.RuneStart(s[0]) {
			s = s[1:]
		}
		truncated = true
	}
	s = strings.TrimRight(s, "\n")
	lines := strings.Split(s, "\n")
	if len(lines) > maxPreviewLines {
		lines = lines[len(lines)-maxPreviewLines:]
		truncated = true
	}
	return strings.Join(lines, "\n"), truncated
}
