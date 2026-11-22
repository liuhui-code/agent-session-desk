// Palette adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/styles.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package ui

import (
	tea "charm.land/bubbletea/v2"
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// This is the single style definition site. Tokens are copied into each model;
// there is no mutable global theme, file watcher or runtime theme format yet.
// Zero/one in viewport algorithms are counts, not additional style choices.
type designTokens struct {
	copy      copyTokens
	sidebar   sidebarTokens
	dialog    dialogTokens
	colors    palette
	workspace workspaceTokens
	layout    layoutTokens
	marks     markTokens
	bold      bool
}

type copyTokens struct {
	foreground, selection, match, current, mark string
	cursorShape                                 tea.CursorShape
	cursorBlink                                 bool
}

type sidebarTokens struct {
	radioUnselected                                                       string
	cardMinWidth, cardRows                                                int
	fullRows, settingsWidth, settingsPaddingX, settingsPaddingY, radioGap int
	selected, active, metadata, metadataMiddle, separator, radioMarker    string
	underlineChoice                                                       bool
}

type dialogTokens struct {
	firstPaddingY                                                            int
	preferredWidth, minWidth, gutter, paddingX, paddingY, titleGap, fieldGap int
	dimBackdrop                                                              bool
	cursorShape                                                              tea.CursorShape
	cursorBlink                                                              bool
}

type workspaceTokens struct {
	headerGapRows, footerGapRows, footerRuleRows, footerBottomGapRows int
	minTerminalRows                                                   int
	sidebarPercent, compactPercent, minPane, tabPadding, maxTabWidth  int
	splitStep, minSplitPercent, maxSplitPercent, dividerGrabSlop      int
	splitNoticeDuration                                               time.Duration
	doubleClickInterval                                               time.Duration
	border                                                            lipgloss.Border
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
type markTokens struct{ rule, divider, selected, ellipsis, menuDot, above, below, add, close string }

func defaultDesignTokens() designTokens {
	return designTokens{
		// Pinned native copy-mode defaults; interpreted as character attributes.
		copy:      copyTokens{foreground: "idx:0", selection: "idx:3", match: "idx:6", current: "idx:5", mark: "idx:1", cursorShape: tea.CursorBlock},
		sidebar:   sidebarTokens{cardMinWidth: 28, cardRows: 2, fullRows: 3, selected: "▶ ", active: "┃ ", metadata: "  ╰ ", metadataMiddle: "  │ ", separator: " · ", settingsWidth: 66, settingsPaddingX: 2, settingsPaddingY: 1, radioGap: 2, radioMarker: ">", radioUnselected: " ", underlineChoice: true},
		dialog:    dialogTokens{preferredWidth: 86, minWidth: 44, gutter: 10, paddingX: 4, paddingY: 2, firstPaddingY: 1, titleGap: 1, fieldGap: 1, dimBackdrop: true, cursorShape: tea.CursorBar, cursorBlink: true},
		workspace: workspaceTokens{footerRuleRows: 1, footerBottomGapRows: 1, sidebarPercent: 35, compactPercent: 18, minPane: 8, tabPadding: 1, maxTabWidth: 32, splitStep: 5, minSplitPercent: 10, maxSplitPercent: 90, dividerGrabSlop: 2, minTerminalRows: 3, splitNoticeDuration: 1500 * time.Millisecond, doubleClickInterval: 500 * time.Millisecond, border: lipgloss.RoundedBorder()},
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
		marks: markTokens{rule: "─", divider: "│", selected: "›", ellipsis: "…", menuDot: "·", add: "+", close: "×", above: "⋮", below: "⋮"},
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

func (t designTokens) workspaceChromeRows() int {
	return 3 + t.workspace.headerGapRows + t.workspace.footerGapRows + t.workspace.footerRuleRows + t.workspace.footerBottomGapRows
}
