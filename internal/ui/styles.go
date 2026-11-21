// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/styles.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified for Agent Session Desk: retain the Tokyo Night palette and small
// style/menu helpers, use Lip Gloss v2, remove global themes and agent branding.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.

package ui

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
)

// palette is immutable for the lifetime of a UI. There is no theme watcher or
// shared mutable renderer state in this first migration slice.
type palette struct {
	bg, surface, border, text, dim, accent, cyan, red color.Color
}

var darkColors = palette{
	bg:      lipgloss.Color("#1a1b26"),
	surface: lipgloss.Color("#24283b"),
	border:  lipgloss.Color("#414868"),
	text:    lipgloss.Color("#c0caf5"),
	dim:     lipgloss.Color("#787fa0"),
	accent:  lipgloss.Color("#7aa2f7"),
	cyan:    lipgloss.Color("#7dcfff"),
	red:     lipgloss.Color("#f7768e"),
}

type styles struct {
	title, highlight, dim, error, text, key, separator lipgloss.Style
}

func newStyles() styles {
	return styles{
		title: lipgloss.NewStyle().Bold(true).
			Foreground(darkColors.accent),
		highlight: lipgloss.NewStyle().Foreground(darkColors.bg).
			Background(darkColors.accent).Bold(true),
		dim:       lipgloss.NewStyle().Foreground(darkColors.dim),
		error:     lipgloss.NewStyle().Foreground(darkColors.red).Bold(true),
		text:      lipgloss.NewStyle().Foreground(darkColors.text),
		key:       lipgloss.NewStyle().Foreground(darkColors.accent).Bold(true),
		separator: lipgloss.NewStyle().Foreground(darkColors.border),
	}
}

// menuKey is adapted from Agent Deck's MenuKey. Keep the small, consistent
// key/description treatment without importing its application hotkey registry.
func (s styles) menuKey(key, description string) string {
	return fmt.Sprintf("%s %s %s", s.key.Render(key),
		s.separator.Render("·"), s.text.Render(description))
}
