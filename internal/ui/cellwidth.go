// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/cellwidth.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0, cellTruncate.
// Modified for Agent Session Desk: keep only the bounded keycap-aware
// truncation helper; replace keycapCount == 0 with a conservative rune check
// to avoid importing a second grapheme iterator just to test for absence.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.

package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// cellTruncate retains Agent Deck's workaround for x/ansi's inconsistent
// keycap measurement/truncation. ANSI StringWidth reports two cells for 1️⃣,
// while Truncate's internal accounting can count only one. Shrink its budget
// until the measured output fits; preserve ANSI boundaries and reset codes.
func cellTruncate(s string, width int, tail string) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= width {
		return s
	}
	if !strings.ContainsRune(s, 0x20E3) {
		return ansi.Truncate(s, width, tail)
	}
	for adjusted := width; adjusted > 0; adjusted-- {
		t := tail
		if adjusted <= ansi.StringWidth(tail) {
			t = ""
		}
		if out := ansi.Truncate(s, adjusted, t); ansi.StringWidth(out) <= width {
			return out
		}
	}
	return ""
}
