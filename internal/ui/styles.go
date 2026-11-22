// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/styles.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified for Agent Session Desk: retain the Tokyo Night palette and small
// style/menu helpers, use Lip Gloss v2, remove global themes and agent branding.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.

package ui

import (
	"charm.land/bubbles/v2/textinput"
	"strings"

	"charm.land/lipgloss/v2"
)

type styles struct {
	sidebarSelected, sidebarMetadataSelected                       lipgloss.Style
	separatorDrag                                                  lipgloss.Style
	formTitle, formText, formHint, formError                       lipgloss.Style
	dialog, backdrop                                               lipgloss.Style
	formInput                                                      textinput.Styles
	tab, tabActive, terminalFrame, terminalFocused                 lipgloss.Style
	title, highlight, dim, error, text, key, separator, panelTitle lipgloss.Style
	menuDot, menuGap                                               string
}

func newStyles(t designTokens) styles {
	return styles{
		sidebarSelected:         lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.surface),
		sidebarMetadataSelected: lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.surface),
		separatorDrag:           lipgloss.NewStyle().Foreground(t.colors.accent),
		formTitle:               lipgloss.NewStyle().Foreground(t.colors.cyan).Background(t.colors.surface).Bold(t.bold),
		formText:                lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.surface),
		formHint:                lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.surface),
		formError:               lipgloss.NewStyle().Foreground(t.colors.red).Background(t.colors.surface).Bold(t.bold),
		dialog:                  lipgloss.NewStyle().Border(t.workspace.border).BorderForeground(t.colors.cyan).Background(t.colors.surface).Padding(t.dialog.paddingY, t.dialog.paddingX),
		backdrop:                lipgloss.NewStyle().Faint(t.dialog.dimBackdrop),
		formInput:               formInputStyles(t),
		tab:                     lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.surface).PaddingLeft(t.workspace.tabPadding).PaddingRight(t.workspace.tabPadding),
		tabActive:               lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.accent).Bold(t.bold).PaddingLeft(t.workspace.tabPadding).PaddingRight(t.workspace.tabPadding),
		terminalFrame:           lipgloss.NewStyle().Border(t.workspace.border).BorderForeground(t.colors.border),
		terminalFocused:         lipgloss.NewStyle().Border(t.workspace.border).BorderForeground(t.colors.accent),
		title: lipgloss.NewStyle().Bold(t.bold).
			Foreground(t.colors.accent),
		highlight: lipgloss.NewStyle().Foreground(t.colors.bg).
			Background(t.colors.accent).Bold(t.bold),
		dim:        lipgloss.NewStyle().Foreground(t.colors.dim),
		error:      lipgloss.NewStyle().Foreground(t.colors.red).Bold(t.bold),
		text:       lipgloss.NewStyle().Foreground(t.colors.text),
		key:        lipgloss.NewStyle().Foreground(t.colors.accent).Bold(t.bold),
		separator:  lipgloss.NewStyle().Foreground(t.colors.border),
		panelTitle: lipgloss.NewStyle().Foreground(t.colors.cyan).Bold(t.bold),
		menuDot:    t.marks.menuDot,
		menuGap:    strings.Repeat(" ", t.layout.menuGap),
	}
}

// menuKey is adapted from Agent Deck's MenuKey. Keep the small, consistent
// key/description treatment without importing its application hotkey registry.
func (s styles) menuKey(key, description string) string {
	return s.key.Render(key) + s.menuGap + s.separator.Render(s.menuDot) + s.menuGap + s.text.Render(description)
}

func formInputStyles(t designTokens) textinput.Styles {
	text := lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.surface)
	dim := lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.surface)
	prompt := lipgloss.NewStyle().Foreground(t.colors.cyan).Background(t.colors.surface)
	state := textinput.StyleState{Text: text, Placeholder: dim, Suggestion: dim, Prompt: prompt}
	return textinput.Styles{Focused: state, Blurred: state, Cursor: textinput.CursorStyle{Color: t.colors.cyan, Shape: t.dialog.cursorShape, Blink: t.dialog.cursorBlink}}
}
