package terminal

import (
	"github.com/charmbracelet/x/ansi"
	"unicode/utf8"
)

// ValidCaptureEncoding admits UTF-8 text and legacy single-byte C1 VT controls
// in an ANSI-rich capture. Strip alone can silently discard an invalid byte
// such as 0xff, so check the raw encoding before validating the stripped text.
// Structured native screen text remains subject to its stricter UTF-8 rules.
func ValidCaptureEncoding(raw string) bool {
	for rest := raw; len(rest) > 0; {
		r, n := utf8.DecodeRuneInString(rest)
		if r == utf8.RuneError && n == 1 && (rest[0] < 0x80 || rest[0] > 0x9f) {
			return false
		}
		rest = rest[n:]
	}
	return utf8.ValidString(ansi.Strip(raw))
}
