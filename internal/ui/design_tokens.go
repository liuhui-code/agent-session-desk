// Palette adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/styles.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package ui

import (
	tea "charm.land/bubbletea/v2"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// This is the single style definition site. Tokens are copied into each model;
// there is no mutable global theme, file watcher or runtime theme format yet.
// Zero/one in viewport algorithms are counts, not additional style choices.
type designTokens struct {
	dialog    dialogTokens
	colors    palette
	workspace workspaceTokens
	layout    layoutTokens
	marks     markTokens
	bold      bool
}

type dialogTokens struct {
	preferredWidth, minWidth, gutter, paddingX, paddingY, titleGap, fieldGap int
	dimBackdrop                                                              bool
	cursorShape                                                              tea.CursorShape
	cursorBlink                                                              bool
}

type workspaceTokens struct {
	sidebarPercent, compactPercent, minPane, tabPadding int
	border                                              lipgloss.Border
}

type palette struct{ bg, surface, border, text, dim, accent, cyan, red color.Color }
type layoutTokens struct {
	singleBelow, stackedBelow                                      int
	minFrameHeight, minPreviewHeight                               int
	frameGap, headingRows, captionGap                              int
	listMin, listMax, listFraction                                 int
	stackedFraction, stackedListMin, stackedGap, previewMin        int
	statusHelpWidth, suffixBelow, indicatorRows, aboveMinHeight    int
	headerMinGap, itemGap, menuGap, dividerPadding, rowPrefixWidth int
}
type markTokens struct{ rule, divider, selected, ellipsis, menuDot, above, below string }

func defaultDesignTokens() designTokens {
	return designTokens{
		dialog:    dialogTokens{preferredWidth: 84, minWidth: 44, gutter: 10, paddingX: 4, paddingY: 1, titleGap: 1, fieldGap: 1, dimBackdrop: true, cursorShape: tea.CursorBar, cursorBlink: true},
		workspace: workspaceTokens{sidebarPercent: 35, compactPercent: 18, minPane: 8, tabPadding: 1, border: lipgloss.RoundedBorder()},
		colors: palette{
			bg: lipgloss.Color("#1a1b26"), surface: lipgloss.Color("#24283b"),
			border: lipgloss.Color("#414868"), text: lipgloss.Color("#c0caf5"),
			dim: lipgloss.Color("#787fa0"), accent: lipgloss.Color("#7aa2f7"),
			cyan: lipgloss.Color("#7dcfff"), red: lipgloss.Color("#f7768e"),
		},
		layout: layoutTokens{
			singleBelow: 50, stackedBelow: 80, minFrameHeight: 5, minPreviewHeight: 10,
			frameGap: 1, headingRows: 2, captionGap: 1,
			listMin: 24, listMax: 42, listFraction: 3,
			stackedFraction: 2, stackedListMin: 3, stackedGap: 1, previewMin: 5,
			statusHelpWidth: 18, suffixBelow: 22, indicatorRows: 1, aboveMinHeight: 3,
			headerMinGap: 1, itemGap: 2, menuGap: 1, dividerPadding: 1, rowPrefixWidth: 2,
		},
		marks: markTokens{rule: "─", divider: "│", selected: "›", ellipsis: "…", menuDot: "·", above: "⋮", below: "⋮"},
		bold:  true,
	}
}

func (t designTokens) itemGap() string      { return strings.Repeat(" ", t.layout.itemGap) }
func (t designTokens) sectionBreak() string { return strings.Repeat("\n", t.layout.frameGap+1) }
func (t designTokens) divider() string {
	pad := strings.Repeat(" ", t.layout.dividerPadding)
	return pad + t.marks.divider + pad
}

const previewTabWidth = 4

func (t designTokens) chromeRows() int { return 2 + 2*t.layout.frameGap }
