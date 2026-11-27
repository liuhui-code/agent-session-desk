// Palette adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/styles.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// This is the single style definition site. Tokens are copied into each model;
// there is no mutable global palette, file watcher or arbitrary theme format.
// Zero/one in viewport algorithms are counts, not additional style choices.
type designTokens struct {
	activityIdle   string
	activityFrames []string
	pet            PetTokens
	attention      attentionTokens
	paths          pathTokens
	footer         footerTokens
	help           helpTokens
	confirm        confirmTokens
	group          groupTokens
	rename         renameTokens
	search         searchTokens
	copy           copyTokens
	sidebar        sidebarTokens
	dialog         dialogTokens
	colors         palette
	workspace      workspaceTokens
	layout         layoutTokens
	marks          markTokens
	bold           bool
}

// PetTokens keeps the native demo's visual values in the same design entry.
type PetTokens struct {
	MaxDPI                                                 int
	Width, Height, SpriteSize, SpriteTop, Margin, DragSlop int
	BadgeX, BadgeY, BadgeRadius, BadgeScale                int
	CaptionY, CaptionWidth, CaptionHeight, CaptionScale    int
	OpenMS, BlinkMS, AlertMS, BreathOffset                 int
	SpriteAnchorAlpha                                      uint8
	Ink, Cream, Alert, Unknown                             color.RGBA
}

func PetDesign() PetTokens { return defaultDesignTokens().pet }

func PetDesignAtDPI(dpi int) PetTokens {
	t := PetDesign()
	dpi = max(96, min(dpi, t.MaxDPI))
	for _, v := range []*int{&t.Width, &t.Height, &t.SpriteSize, &t.SpriteTop, &t.Margin, &t.DragSlop, &t.BadgeX, &t.BadgeY, &t.BadgeRadius, &t.BadgeScale, &t.CaptionY, &t.CaptionWidth, &t.CaptionHeight, &t.CaptionScale, &t.BreathOffset} {
		*v = max(1, (*v*dpi+48)/96)
	}
	return t
}

type attentionTokens struct {
	snoozedLabel                                                                string
	marker, idle, label, uncertain                                              string
	centerWidth, centerGutter, centerRows, actionWidth, actionHeight, helpBelow int
}

type pathTokens struct {
	background, purple                                                       color.Color
	paddingX, paddingY, width, minWidth, minHeight, rows, shortRows, shortAt int
	selected, idle, custom, above, below                                     string
	italic                                                                   bool
}

type footerTokens struct {
	curatedHintLimit, curatedGap                                                      int
	fullRightGutter, compactRightGutter                                               int
	compactBelow, fullAbove, fullKeyPadding, labelGap, hintGap, blockGap              int
	separator, inlineSeparator, targetPrefix, contextSuffix, badgePrefix, badgeSuffix string
	badgePadding                                                                      int
	contextColor                                                                      color.Color
}

type helpTokens struct {
	hintItalic                                                                                        bool
	shortWidth, fullWidth, wideWidth, wideAt, minWidth, gutter                                        int
	paddingX, paddingY, shortKeyWidth, keyWidth, compactKeyWidth, compactBelow                        int
	shortIndent, rowIndent, columnGap, sectionGap, footerGap, screenReserve, minContentRows, pageStep int
	keyColor                                                                                          color.Color
	above, below, separator                                                                           string
}

type confirmTokens struct {
	preferredWidth, responsiveBelow, narrowGutter, paddingX, paddingY   int
	sectionGap, buttonPadding, buttonGap, nameIndent                    int
	titlePrefix, bullet, focusedPrefix, idlePrefix, nameOpen, nameClose string
}

type groupTokens struct {
	placeholderColor                      color.Color
	labelWidth, inputWidth, choicePadding int
	modeSeparator, modeOpen, modeClose    string
}

type renameTokens struct {
	preferredWidth, minWidth, gutter, inputWidth, paddingX, paddingY int
	prompt, warningPrefix                                            string
	borderColor                                                      color.Color
}

type searchTokens struct {
	placeholderColor                                                                           color.Color
	titlePrefix, programPrefix, programSuffix                                                  string
	preferredWidth, minWidth, responsiveBelow, gutter                                          int
	paddingX, paddingY, inputPaddingX, inputPaddingY, inputRightGap, resultPadding, maxResults int
	prompt, selected                                                                           string
	countColor                                                                                 color.Color
	hintItalic                                                                                 bool
}

type copyTokens struct {
	foreground, selection, match, current, mark string
	cursorShape                                 tea.CursorShape
	cursorBlink                                 bool
}

type sidebarTokens struct {
	groupExpanded, groupCollapsed, rootShortcut, groupCountPrefix, groupCountSuffix string
	groupGutter, indentWidth                                                        int
	rootShortcutSuffix                                                              string
	radioUnselected                                                                 string
	cardGap, groupRows                                                              int
	groupBold                                                                       bool
	cardMinWidth, cardRows                                                          int
	fullRows, settingsWidth, settingsPaddingX, settingsPaddingY, radioGap           int
	selected, active, metadata, metadataMiddle, separator, radioMarker              string
	underlineChoice                                                                 bool
	settingsFocusedPrefix, settingsIdlePrefix                                       string
}

type dialogTokens struct {
	firstPaddingY                                                            int
	preferredWidth, minWidth, gutter, paddingX, paddingY, titleGap, fieldGap int
	dimBackdrop                                                              bool
	cursorShape                                                              tea.CursorShape
	cursorBlink                                                              bool
}

type workspaceTokens struct {
	terminalHint, managementHint, copyHint                            string
	sideInset, topInsetRows                                           int
	footerRevealRows, dividerInsetRows                                int
	headerGapRows, footerGapRows, footerRuleRows, footerBottomGapRows int
	minTerminalRows                                                   int
	sidebarPercent, compactPercent, minPane, tabPadding, maxTabWidth  int
	splitStep, minSplitPercent, maxSplitPercent, dividerGrabSlop      int
	splitNoticeDuration                                               time.Duration
	doubleClickInterval                                               time.Duration
	border                                                            lipgloss.Border
}

type palette struct{ bg, surface, border, text, dim, accent, activityBlue, cyan, red, yellow, activeBackground, activeText, groupText, focusBackground, focusText color.Color }
type layoutTokens struct {
	singleBelow, stackedBelow                                      int
	minFrameHeight, minPreviewHeight                               int
	frameGap, headingRows, captionGap                              int
	listMin, listMax, listFraction                                 int
	stackedFraction, stackedListMin, stackedGap, previewMin        int
	statusHelpWidth, suffixBelow, indicatorRows, aboveMinHeight    int
	headerMinGap, itemGap, menuGap, dividerPadding, rowPrefixWidth int
}
type markTokens struct {
	rule, divider, selected, ellipsis, menuDot, above, below, add, close string
	dividerHandle                                                        string
	identityOpen, identityClose                                          string
	identityPrefixLength                                                 int
}

func defaultDesignTokens() designTokens {
	colors, purple := themePalette("dark")
	return designTokens{
		pet: PetTokens{MaxDPI: 384, Width: 160, Height: 176, SpriteSize: 128, SpriteTop: 18, Margin: 16, DragSlop: 4,
			BadgeX: 126, BadgeY: 24, BadgeRadius: 17, BadgeScale: 2,
			CaptionY: 152, CaptionWidth: 62, CaptionHeight: 20, CaptionScale: 2,
			OpenMS: 3200, BlinkMS: 160, AlertMS: 600, BreathOffset: 1,
			SpriteAnchorAlpha: 200,
			Ink:               color.RGBA{R: 83, G: 48, B: 32, A: 255}, Cream: color.RGBA{R: 255, G: 245, B: 222, A: 255},
			Alert: color.RGBA{R: 255, G: 198, B: 81, A: 255}, Unknown: color.RGBA{R: 195, G: 197, B: 202, A: 255}},
		attention: attentionTokens{marker: "! ", idle: "  ", label: "! %d need attention · Alt+N · Alt+A All", uncertain: " · last observed", snoozedLabel: " · %d snoozed", centerWidth: 84, centerGutter: 4, centerRows: 14, actionWidth: 27, actionHeight: 10, helpBelow: 58},
		paths:     pathTokens{background: lipgloss.Color("#222222"), purple: purple, paddingX: 1, width: 60, minWidth: 32, minHeight: 12, rows: 5, shortRows: 3, shortAt: 30, selected: "▶ ", idle: "  ", custom: "✎ Enter a custom path…", above: "↑ More above ", below: "↓ More below ", italic: true},
		footer:    footerTokens{curatedHintLimit: 3, curatedGap: 2, fullRightGutter: 2, compactRightGutter: 1, compactBelow: 70, fullAbove: 100, fullKeyPadding: 1, labelGap: 1, hintGap: 1, blockGap: 2, separator: " │ ", inlineSeparator: " · ", targetPrefix: "→ ", contextSuffix: ":", badgePrefix: " -- ", badgeSuffix: " -- ", badgePadding: 1, contextColor: purple},
		help:      helpTokens{hintItalic: true, shortWidth: 58, fullWidth: 72, wideWidth: 82, wideAt: 100, minWidth: 37, gutter: 8, paddingX: 2, paddingY: 1, shortKeyWidth: 9, keyWidth: 14, compactKeyWidth: 10, compactBelow: 47, shortIndent: 1, rowIndent: 2, columnGap: 1, sectionGap: 1, footerGap: 1, screenReserve: 8, minContentRows: 10, pageStep: 10, keyColor: purple, above: "▲ More above", below: "▼ More below", separator: " │ "},
		confirm:   confirmTokens{preferredWidth: 52, responsiveBelow: 60, narrowGutter: 8, paddingX: 2, paddingY: 1, sectionGap: 1, buttonPadding: 2, buttonGap: 2, nameIndent: 2, titlePrefix: "⚠  ", bullet: "• ", focusedPrefix: "▸ ", idlePrefix: "  ", nameOpen: "\"", nameClose: "\""},
		group:     groupTokens{placeholderColor: lipgloss.Color("240"), labelWidth: 14, inputWidth: 30, choicePadding: 1, modeSeparator: " ─── ", modeOpen: "[", modeClose: "]"},
		// Pinned group_dialog.go: v1 Width 44/30 excludes the two border cells.
		rename: renameTokens{preferredWidth: 46, minWidth: 32, gutter: 8, inputWidth: 32, paddingX: 2, paddingY: 1, prompt: "> ", warningPrefix: "⚠ ", borderColor: purple},
		// Pinned search.go: v1 Width 60/30 includes padding, excludes border.
		search: searchTokens{preferredWidth: 62, minWidth: 32, responsiveBelow: 70, gutter: 8, paddingX: 2, paddingY: 1, inputPaddingX: 1, inputRightGap: 4, resultPadding: 2, maxResults: 10, prompt: "> ", selected: "› ", titlePrefix: "🔍 ", programPrefix: " (", programSuffix: ")", countColor: lipgloss.Color("241"), placeholderColor: lipgloss.Color("240"), hintItalic: true},
		// Pinned native copy-mode defaults; interpreted as character attributes.
		copy:         copyTokens{foreground: "idx:0", selection: "idx:3", match: "idx:6", current: "idx:5", mark: "idx:1", cursorShape: tea.CursorBlock},
		activityIdle: " ●", activityFrames: []string{" ⠋", " ⠙", " ⠹", " ⠸", " ⠼", " ⠴", " ⠦", " ⠧", " ⠇", " ⠏"},
		sidebar:   sidebarTokens{cardGap: 1, groupRows: 2, groupBold: true, cardMinWidth: 28, cardRows: 2, fullRows: 3, selected: "  ", active: "┃ ", metadata: "  ╰ ", metadataMiddle: "  │ ", separator: " · ", settingsWidth: 66, settingsPaddingX: 2, settingsPaddingY: 1, radioGap: 2, radioMarker: ">", radioUnselected: " ", underlineChoice: true, settingsFocusedPrefix: "▶ ", settingsIdlePrefix: "  ", groupExpanded: "▾", groupCollapsed: "▸", rootShortcut: "1·", groupCountPrefix: " (", groupCountSuffix: ")", groupGutter: 0, indentWidth: 2, rootShortcutSuffix: "·"},
		dialog:    dialogTokens{preferredWidth: 86, minWidth: 44, gutter: 10, paddingX: 4, paddingY: 2, firstPaddingY: 1, titleGap: 1, fieldGap: 1, dimBackdrop: true, cursorShape: tea.CursorBar, cursorBlink: true},
		workspace: workspaceTokens{terminalHint: "F6 Navigate · Alt+J/K Switch sessions", managementHint: "F6 Terminal · n New · / Search · S Settings · ? Help", copyHint: "Space Select · Enter Copy · / Search · Esc/q Back", sideInset: 2, topInsetRows: 1, footerRevealRows: 1, dividerInsetRows: 2, sidebarPercent: 35, compactPercent: 18, minPane: 8, tabPadding: 1, maxTabWidth: 32, splitStep: 5, minSplitPercent: 10, maxSplitPercent: 90, dividerGrabSlop: 2, minTerminalRows: 3, splitNoticeDuration: 1500 * time.Millisecond, doubleClickInterval: 500 * time.Millisecond, border: lipgloss.RoundedBorder()},
		colors:    colors,
		layout: layoutTokens{
			singleBelow: 50, stackedBelow: 80, minFrameHeight: 5, minPreviewHeight: 10,
			frameGap: 1, headingRows: 2, captionGap: 1,
			listMin: 24, listMax: 42, listFraction: 3,
			stackedFraction: 2, stackedListMin: 3, stackedGap: 1, previewMin: 5,
			statusHelpWidth: 18, suffixBelow: 22, indicatorRows: 1, aboveMinHeight: 3,
			headerMinGap: 1, itemGap: 2, menuGap: 1, dividerPadding: 1, rowPrefixWidth: 2,
		},
		marks: markTokens{rule: "─", divider: "│", dividerHandle: "┃", selected: "›", ellipsis: "…", menuDot: "·", add: "+", close: "×", above: "⋮", below: "⋮", identityOpen: " [", identityClose: "]", identityPrefixLength: 6},
		bold:  false,
	}
}

// Light retains the pinned source. Dark follows the user's iTerm Default
// black/#bbbbbb palette; neutral surfaces and dim roles keep UI chrome quiet.
func themePalette(theme string) (palette, color.Color) {
	if theme == "light" {
		return palette{activityBlue: lipgloss.Color("#1677ff"), focusBackground: lipgloss.Color("#b3c7d4"), focusText: lipgloss.Color("#1a3140"), groupText: lipgloss.Color("#34548a"), activeBackground: lipgloss.Color("#cbdce8"), activeText: lipgloss.Color("#23465a"), bg: lipgloss.Color("#d5d6db"), surface: lipgloss.Color("#e9e9ec"), border: lipgloss.Color("#9699a3"), text: lipgloss.Color("#343b58"), dim: lipgloss.Color("#6a6d7c"), accent: lipgloss.Color("#34548a"), cyan: lipgloss.Color("#166775"), red: lipgloss.Color("#8c4351"), yellow: lipgloss.Color("#8f5e15")}, lipgloss.Color("#7847bd")
	}
	return palette{activityBlue: lipgloss.Color("#1677ff"), focusBackground: lipgloss.Color("#38576b"), focusText: lipgloss.Color("#ffffff"), groupText: lipgloss.Color("#8ca5b8"), activeBackground: lipgloss.Color("#20303b"), activeText: lipgloss.Color("#bcd4e2"), bg: lipgloss.Color("#000000"), surface: lipgloss.Color("#222222"), border: lipgloss.Color("#333333"), text: lipgloss.Color("#bbbbbb"), dim: lipgloss.Color("#777777"), accent: lipgloss.Color("#bbbbbb"), cyan: lipgloss.Color("#bbbbbb"), red: lipgloss.Color("#bb0000"), yellow: lipgloss.Color("#bbbb00")}, lipgloss.Color("#bbbbbb")
}

func (t designTokens) withTheme(theme string) designTokens {
	colors, purple := themePalette(theme)
	t.colors = colors
	t.paths.purple = purple
	t.paths.background = lipgloss.Color("#222222")
	if theme == "light" {
		t.paths.background = lipgloss.Color("#dcdde2")
	}
	t.footer.contextColor, t.help.keyColor, t.rename.borderColor = purple, purple, purple
	return t
}

// Pinned TrueColor observations, not replacement authored palette values:
// Agent Deck 80494e5 via Lip Gloss 1.1.0/termenv 0.16.0 truncates these two
// channels. Keep the measured display adaptation local to style construction.
// See ADR-0035/0036; only UI tokens pass through, never native screen attributes.
func (t designTokens) forDisplay(profile colorprofile.Profile) designTokens {
	if profile == colorprofile.ANSI || profile == colorprofile.ANSI256 || profile == colorprofile.ASCII || profile == colorprofile.NoTTY {
		return t.forLimitedDisplay(profile)
	}
	// ADR-0037: pinned floating-menu dark RGB is emitted as 40,46,65.
	if t.paths.background == lipgloss.Color("#292e42") {
		t.paths.background = lipgloss.Color("#282e41")
	}
	if t.colors.accent == lipgloss.Color("#7aa2f7") {
		t.colors.accent = lipgloss.Color("#79a2f7")
	}
	if t.colors.dim == lipgloss.Color("#6a6d7c") {
		t.colors.dim = lipgloss.Color("#696d7c")
	}
	return t
}

func (t designTokens) itemGap() string      { return strings.Repeat(" ", t.layout.itemGap) }
func (t designTokens) sectionBreak() string { return strings.Repeat("\n", t.layout.frameGap+1) }
func (t designTokens) divider() string {
	return t.dividerWith(t.marks.divider)
}
func (t designTokens) dividerWith(glyph string) string {
	pad := strings.Repeat(" ", t.layout.dividerPadding)
	return pad + glyph + pad
}

const previewTabWidth = 4

func (t designTokens) chromeRows() int { return 2 + 2*t.layout.frameGap }

func (t designTokens) workspaceChromeRows() int {
	return 0 // No permanent workspace header, duplicate Tab strip or footer.
}

// Data measured from the fixed source renderer, not a general RGB converter.
func sourceIndexedColor(value color.Color, profile colorprofile.Profile) color.Color {
	if profile == colorprofile.ASCII || profile == colorprofile.NoTTY {
		return nil
	}
	for _, rule := range [...]struct{ authored, ansi, ansi256 string }{
		{"#38576b", "4", "24"},
		{"#ffffff", "15", "15"},
		{"#b3c7d4", "7", "152"},
		{"#1a3140", "0", "236"},
		{"#8ca5b8", "7", "109"},
		{"#20303b", "4", "236"},
		{"#bcd4e2", "7", "152"},
		{"#cbdce8", "7", "189"},
		{"#23465a", "4", "24"},
		{"#000000", "0", "0"},
		{"#bbbbbb", "7", "7"},
		{"#777777", "8", "243"},
		{"#222222", "0", "235"},
		{"#333333", "8", "236"},
		{"#bb0000", "1", "1"},
		{"#bbbb00", "3", "3"},
		{"#292e42", "4", "17"},
		{"#dcdde2", "7", "188"},
		{"#166775", "12", "24"},
		{"#1a1b26", "0", "232"},
		{"#24283b", "4", "17"},
		{"#343b58", "8", "59"},
		{"#34548a", "12", "60"},
		{"#414868", "8", "59"},
		{"#6a6d7c", "12", "60"},
		{"#7847bd", "12", "97"},
		{"#787fa0", "12", "103"},
		{"#7aa2f7", "12", "111"},
		{"#7dcfff", "14", "117"},
		{"#8c4351", "8", "95"},
		{"#8f5e15", "3", "94"},
		{"#9699a3", "12", "103"},
		{"#bb9af7", "13", "141"},
		{"#c0caf5", "14", "153"},
		{"#d5d6db", "7", "188"},
		{"#e0af68", "3", "179"},
		{"#e9e9ec", "13", "189"},
		{"#f7768e", "9", "210"},
		{"240", "8", "240"},
		{"241", "8", "241"},
	} {
		if value == lipgloss.Color(rule.authored) {
			if profile == colorprofile.ANSI {
				return lipgloss.Color(rule.ansi)
			}
			return lipgloss.Color(rule.ansi256)
		}
	}
	return value
}

func (t designTokens) forLimitedDisplay(profile colorprofile.Profile) designTokens {
	for _, value := range []*color.Color{&t.paths.background, &t.paths.purple, &t.colors.bg, &t.colors.surface, &t.colors.activeBackground, &t.colors.activeText, &t.colors.groupText, &t.colors.focusBackground, &t.colors.focusText, &t.colors.border, &t.colors.text, &t.colors.dim, &t.colors.accent, &t.colors.activityBlue, &t.colors.cyan, &t.colors.red, &t.colors.yellow, &t.footer.contextColor, &t.help.keyColor, &t.rename.borderColor, &t.group.placeholderColor, &t.search.placeholderColor, &t.search.countColor} {
		*value = sourceIndexedColor(*value, profile)
	}
	if profile == colorprofile.ASCII || profile == colorprofile.NoTTY {
		t.paths.italic = false
		t.bold, t.help.hintItalic, t.search.hintItalic, t.sidebar.underlineChoice = false, false, false, false
	}
	return t
}
