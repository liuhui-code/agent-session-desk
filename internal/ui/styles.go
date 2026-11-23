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
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type styles struct {
	footerCuratedKey                                                                                                          lipgloss.Style
	settingsSection, settingsFocusMarker, settingsChoiceIdle, settingsRadioIdle                                               lipgloss.Style
	footerContext, footerKey, footerFullKey, footerDescription, footerSeparator, footerBadge, footerCompactBadge              lipgloss.Style
	helpBox, helpTitle, helpSection, helpKey, helpText, helpHint, helpNote, helpTab, helpScroll                               lipgloss.Style
	confirmBox, confirmNoticeBox, confirmTitle, confirmNoticeTitle, confirmWarning, confirmDetails, confirmHint, confirmError lipgloss.Style
	confirmDelete, confirmDeleteFocused, confirmCancel, confirmCancelFocused                                                  lipgloss.Style
	groupInput                                                                                                                textinput.Styles
	groupLabel, groupContext, groupMode, groupChoice, groupChoiceSelected                                                     lipgloss.Style
	groupName, groupPart, groupHotkey, groupSelectedName, groupSelectedPart                                                   lipgloss.Style
	renameBox, renameTitle, renameHint                                                                                        lipgloss.Style
	renameInput                                                                                                               textinput.Styles
	searchTip                                                                                                                 lipgloss.Style
	searchBox, searchInput, searchTitle, searchResult, searchSelected, searchCount, searchHint                                lipgloss.Style
	copy                                                                                                                      terminal.CopyStyles
	settingsBox, settingsTitle, settingsText, settingsHint, settingsChoice, settingsRadio, settingsFocused, settingsError     lipgloss.Style
	sidebarSelected, sidebarMetadataSelected                                                                                  lipgloss.Style
	separatorDrag                                                                                                             lipgloss.Style
	formTitle, formText, formHint, formError                                                                                  lipgloss.Style
	dialog, backdrop                                                                                                          lipgloss.Style
	formInput                                                                                                                 textinput.Styles
	tab, tabActive, terminalFrame, terminalFocused                                                                            lipgloss.Style
	title, highlight, dim, error, text, key, separator, panelTitle                                                            lipgloss.Style
	menuDot, menuGap                                                                                                          string
}

func newStyles(t designTokens) styles {
	t = t.forDisplay()
	return styles{
		footerCuratedKey:        lipgloss.NewStyle().Foreground(t.colors.dim).Bold(t.bold),
		footerContext:           lipgloss.NewStyle().Foreground(t.footer.contextColor).Bold(t.bold),
		footerKey:               lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.accent).Bold(t.bold),
		footerFullKey:           lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.accent).Bold(t.bold).PaddingLeft(t.footer.fullKeyPadding).PaddingRight(t.footer.fullKeyPadding),
		footerDescription:       lipgloss.NewStyle().Foreground(t.colors.text),
		footerSeparator:         lipgloss.NewStyle().Foreground(t.colors.border),
		footerBadge:             lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.yellow).Bold(t.bold).PaddingLeft(t.footer.badgePadding).PaddingRight(t.footer.badgePadding),
		footerCompactBadge:      lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.yellow).Bold(t.bold),
		helpBox:                 lipgloss.NewStyle().Border(t.workspace.border).BorderForeground(t.help.keyColor).Background(t.colors.surface).Padding(t.help.paddingY, t.help.paddingX),
		helpTitle:               lipgloss.NewStyle().Foreground(t.colors.accent).Background(t.colors.surface).Bold(t.bold),
		helpSection:             lipgloss.NewStyle().Foreground(t.colors.cyan).Background(t.colors.surface).Bold(t.bold),
		helpKey:                 lipgloss.NewStyle().Foreground(t.help.keyColor).Background(t.colors.surface),
		helpText:                lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.surface),
		helpHint:                lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.surface).Italic(t.help.hintItalic),
		helpNote:                lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.surface),
		helpTab:                 lipgloss.NewStyle().Foreground(t.colors.cyan).Background(t.colors.surface),
		helpScroll:              lipgloss.NewStyle().Foreground(t.colors.yellow).Background(t.colors.surface).Bold(t.bold),
		confirmBox:              lipgloss.NewStyle().Border(t.workspace.border).BorderForeground(t.colors.red).Background(t.colors.bg).Padding(t.confirm.paddingY, t.confirm.paddingX),
		confirmNoticeBox:        lipgloss.NewStyle().Border(t.workspace.border).BorderForeground(t.colors.yellow).Background(t.colors.bg).Padding(t.confirm.paddingY, t.confirm.paddingX),
		confirmTitle:            lipgloss.NewStyle().Foreground(t.colors.red).Background(t.colors.bg).Bold(t.bold).MarginBottom(t.confirm.sectionGap),
		confirmNoticeTitle:      lipgloss.NewStyle().Foreground(t.colors.yellow).Background(t.colors.bg).Bold(t.bold).MarginBottom(t.confirm.sectionGap),
		confirmWarning:          lipgloss.NewStyle().Foreground(t.colors.yellow).Background(t.colors.bg).MarginBottom(t.confirm.sectionGap),
		confirmDetails:          lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.bg).MarginBottom(t.confirm.sectionGap),
		confirmHint:             lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.bg),
		confirmError:            lipgloss.NewStyle().Foreground(t.colors.red).Background(t.colors.bg),
		confirmDelete:           lipgloss.NewStyle().Foreground(t.colors.red).Background(t.colors.bg).Bold(t.bold).PaddingLeft(t.confirm.buttonPadding).PaddingRight(t.confirm.buttonPadding),
		confirmDeleteFocused:    lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.red).Bold(t.bold).PaddingLeft(t.confirm.buttonPadding).PaddingRight(t.confirm.buttonPadding),
		confirmCancel:           lipgloss.NewStyle().Foreground(t.colors.accent).Background(t.colors.bg).Bold(t.bold).PaddingLeft(t.confirm.buttonPadding).PaddingRight(t.confirm.buttonPadding),
		confirmCancelFocused:    lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.accent).Bold(t.bold).PaddingLeft(t.confirm.buttonPadding).PaddingRight(t.confirm.buttonPadding),
		groupInput:              groupInputStyles(t),
		groupLabel:              lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.surface),
		groupContext:            lipgloss.NewStyle().Foreground(t.colors.cyan).Background(t.colors.surface),
		groupMode:               lipgloss.NewStyle().Foreground(t.colors.accent).Background(t.colors.surface).Bold(t.bold),
		groupChoice:             lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.surface).PaddingLeft(t.group.choicePadding).PaddingRight(t.group.choicePadding),
		groupChoiceSelected:     lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.accent).Bold(t.bold).PaddingLeft(t.group.choicePadding).PaddingRight(t.group.choicePadding),
		groupName:               lipgloss.NewStyle().Foreground(t.colors.cyan).Bold(t.bold),
		groupPart:               lipgloss.NewStyle().Foreground(t.colors.text),
		groupHotkey:             lipgloss.NewStyle().Foreground(t.colors.dim),
		groupSelectedName:       lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.accent).Bold(t.bold),
		groupSelectedPart:       lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.accent),
		renameBox:               lipgloss.NewStyle().Border(t.workspace.border).BorderForeground(t.rename.borderColor).Background(t.colors.surface).Padding(t.rename.paddingY, t.rename.paddingX),
		renameTitle:             lipgloss.NewStyle().Foreground(t.rename.borderColor).Background(t.colors.surface).Bold(t.bold),
		renameHint:              lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.surface),
		renameInput:             renameInputStyles(t),
		searchTip:               lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.bg).Italic(t.search.hintItalic),
		searchBox:               lipgloss.NewStyle().Border(t.workspace.border).BorderForeground(t.colors.accent).Background(t.colors.bg).Padding(t.search.paddingY, t.search.paddingX),
		searchInput:             lipgloss.NewStyle().Border(t.workspace.border).BorderForeground(t.colors.accent).Background(t.colors.bg).Padding(t.search.inputPaddingY, t.search.inputPaddingX),
		searchTitle:             lipgloss.NewStyle().Foreground(t.colors.accent).Background(t.colors.bg).Bold(t.bold),
		searchResult:            lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.bg).PaddingLeft(t.search.resultPadding).PaddingRight(t.search.resultPadding),
		searchSelected:          lipgloss.NewStyle().Foreground(t.colors.bg).Background(t.colors.accent).PaddingLeft(t.search.resultPadding).PaddingRight(t.search.resultPadding),
		searchCount:             lipgloss.NewStyle().Foreground(t.search.countColor).Background(t.colors.bg),
		searchHint:              lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.bg),
		copy:                    newCopyStyles(t),
		settingsRadio:           lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.surface),
		settingsRadioIdle:       lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.bg),
		settingsChoiceIdle:      lipgloss.NewStyle().Foreground(t.colors.accent).Background(t.colors.bg).Bold(t.bold),
		settingsFocusMarker:     lipgloss.NewStyle().Foreground(t.colors.accent).Background(t.colors.surface).Bold(t.bold),
		settingsSection:         lipgloss.NewStyle().Foreground(t.colors.accent).Background(t.colors.bg).Bold(t.bold),
		settingsBox:             lipgloss.NewStyle().Border(t.workspace.border).BorderForeground(t.colors.cyan).Background(t.colors.bg).Padding(t.sidebar.settingsPaddingY, t.sidebar.settingsPaddingX),
		settingsTitle:           lipgloss.NewStyle().Foreground(t.colors.cyan).Background(t.colors.bg).Bold(t.bold),
		settingsText:            lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.bg),
		settingsHint:            lipgloss.NewStyle().Foreground(t.colors.dim).Background(t.colors.bg),
		settingsChoice:          lipgloss.NewStyle().Foreground(t.colors.accent).Background(t.colors.surface).Bold(t.bold).Underline(t.sidebar.underlineChoice),
		settingsFocused:         lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.surface),
		settingsError:           lipgloss.NewStyle().Foreground(t.colors.red).Background(t.colors.bg),
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

func renameInputStyles(t designTokens) textinput.Styles {
	text := lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.surface)
	state := textinput.StyleState{Text: text, Placeholder: text, Prompt: text, Suggestion: text}
	return textinput.Styles{Focused: state, Blurred: state, Cursor: textinput.CursorStyle{Color: t.colors.text, Shape: t.dialog.cursorShape, Blink: t.dialog.cursorBlink}}
}

func groupInputStyles(t designTokens) textinput.Styles {
	styles := renameInputStyles(t)
	placeholder := lipgloss.NewStyle().Foreground(t.group.placeholderColor).Background(t.colors.surface)
	styles.Focused.Placeholder, styles.Blurred.Placeholder = placeholder, placeholder
	return styles
}

func searchInputStyles(t designTokens) textinput.Styles {
	text := lipgloss.NewStyle().Foreground(t.colors.text).Background(t.colors.bg)
	dim := lipgloss.NewStyle().Foreground(t.search.placeholderColor).Background(t.colors.bg)
	state := textinput.StyleState{Text: text, Placeholder: dim, Suggestion: dim, Prompt: text}
	return textinput.Styles{Focused: state, Blurred: state, Cursor: textinput.CursorStyle{Color: t.colors.text, Shape: t.dialog.cursorShape, Blink: t.dialog.cursorBlink}}
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

func newCopyStyles(t designTokens) terminal.CopyStyles {
	style := func(background string) terminal.Run {
		return terminal.Run{Foreground: t.copy.foreground, Background: background}
	}
	return terminal.CopyStyles{Selection: style(t.copy.selection), Highlights: [4]terminal.Run{style(t.copy.match), style(t.copy.current), style(t.copy.mark), {Foreground: t.copy.mark, Background: t.copy.foreground}}}
}
