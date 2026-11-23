package session

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const MaxTitleRunes = 128
const MaxTitleBytes = 512
const MaxTitleInputBytes = 4096 // also admits existing launch-snapshot names

// NormalizeRenameTitle does not parse shell syntax or normalize Unicode forms.
// Legacy launch names remain immutable snapshots; this applies to new edits.
func NormalizeRenameTitle(value string) (string, error) {
	if len(value) > MaxTitleInputBytes || !utf8.ValidString(value) {
		return "", fmt.Errorf("name input exceeds bounds or is not UTF-8")
	}
	for _, r := range value {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return "", fmt.Errorf("name must be one line without control characters")
		}
	}
	value = strings.TrimSpace(value)
	if value == "" || len(value) > MaxTitleBytes || utf8.RuneCountInString(value) > MaxTitleRunes {
		return "", fmt.Errorf("name must contain 1–128 Unicode code points and at most 512 UTF-8 bytes")
	}
	return value, nil
}
