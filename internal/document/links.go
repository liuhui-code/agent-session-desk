// Package document recognizes explicit Markdown references, never terminal commands.
package document

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

var (
	markdownReference = regexp.MustCompile(`(?i)^(.+\.(?:md|markdown))(?:#[^\r\n]*|:[0-9]+(?::[0-9]+)?)?$`)
	inlineLink        = regexp.MustCompile(`\[[^\]\r\n]+\]\((?:<([^>\r\n]+)>|([^()\r\n]+))\)`)
	quotedLink        = regexp.MustCompile("`[^`\\r\\n]+`|\"[^\"\\r\\n]+\"|'[^'\\r\\n]+'")
)

// At maps the displayed terminal cell to a single fully visible reference.
// It supports inline Markdown, quoted paths and plain paths; no VT is executed.
func At(text string, column int) (string, bool) {
	if column < 0 || len(text) > 64*1024 || !utf8.ValidString(text) {
		return "", false
	}
	hit := func(start, end int) bool {
		return column >= ansi.StringWidth(text[:start]) && column < ansi.StringWidth(text[:end])
	}
	for _, m := range inlineLink.FindAllStringSubmatchIndex(text, -1) {
		if !hit(m[0], m[1]) {
			continue
		}
		start, end := m[2], m[3]
		if start < 0 {
			start, end = m[4], m[5]
		}
		ref := strings.TrimSpace(text[start:end])
		return ref, markdownReference.MatchString(ref)
	}
	for _, m := range quotedLink.FindAllStringIndex(text, -1) {
		if hit(m[0], m[1]) {
			ref := text[m[0]+1 : m[1]-1]
			return ref, markdownReference.MatchString(ref)
		}
	}
	// Complex Markdown or an unquoted path containing spaces is ambiguous.
	// Refuse to reinterpret a suffix as a different relative filename.
	if strings.Contains(text, "](") {
		return "", false
	}
	uncertainPrefix := false
	boundary := func(r rune) bool { return unicode.IsSpace(r) || strings.ContainsRune("`\"'()[]{}<>", r) }
	for offset := 0; offset < len(text); {
		r, size := utf8.DecodeRuneInString(text[offset:])
		if boundary(r) {
			offset += size
			continue
		}
		start := offset
		for offset < len(text) {
			r, size = utf8.DecodeRuneInString(text[offset:])
			if boundary(r) {
				break
			}
			offset += size
		}
		end := offset
		ref := strings.TrimRight(text[start:end], ",;.!?")
		end = start + len(ref)
		isLink := markdownReference.MatchString(ref)
		if hit(start, end) && isLink {
			if uncertainPrefix {
				return "", false
			}
			return ref, true
		}
		if !isLink && strings.ContainsAny(ref, `/\`) {
			uncertainPrefix = true
		}
	}
	return "", false
}
