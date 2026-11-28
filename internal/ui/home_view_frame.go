// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/home_view_frame.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified for Agent Session Desk: retain responsive dual/stacked/single
// layout, panel titles and exact-size helpers; route product Tab terminals, remove
// filter/account/system panels and v1 renderer controls; use Bubble Tea v2.
// See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.

package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) View() tea.View {
	if m.recoveryUI.center != nil {
		view := tea.NewView(m.renderRecoveryCenter())
		view.AltScreen = true
		m.applyViewColors(&view)
		return view
	}
	if m.diagnostic.open {
		view := tea.NewView(m.renderDiagnostics())
		view.AltScreen = true
		m.applyViewColors(&view)
		return view
	}
	if m.workspace != nil && m.attention.center != nil {
		view := tea.NewView(m.renderAttentionCenter())
		view.AltScreen = true
		m.applyViewColors(&view)
		view.MouseMode = tea.MouseModeCellMotion
		return view
	}
	if m.workspace != nil && m.help {
		view := tea.NewView(m.renderHelpOverlay())
		view.AltScreen = true
		m.applyViewColors(&view)
		return view
	}
	if m.workspace != nil && m.sessionStop != nil {
		view := tea.NewView(m.renderSessionStop())
		view.AltScreen = true
		m.applyViewColors(&view)
		return view
	}
	if m.workspace != nil && m.groupRemoval != nil {
		view := tea.NewView(m.renderGroupRemoval())
		view.AltScreen = true
		m.applyViewColors(&view)
		return view
	}
	if m.workspace != nil && m.groupDialog != nil {
		content, cursor := m.renderGroupDialog()
		view := tea.NewView(content)
		view.AltScreen = true
		m.applyViewColors(&view)
		view.Cursor = cursor
		return view
	}
	if m.workspace != nil && m.rename != nil {
		content, cursor := m.renderSessionRename()
		view := tea.NewView(content)
		view.AltScreen = true
		m.applyViewColors(&view)
		view.Cursor = cursor
		return view
	}
	if m.workspace != nil && m.search != nil {
		content, cursor := m.renderSessionSearch()
		view := tea.NewView(content)
		view.AltScreen = true
		m.applyViewColors(&view)
		view.Cursor = cursor
		return view
	}
	if m.workspace != nil && m.densitySettings != nil {
		view := tea.NewView(m.renderDensitySettings())
		view.AltScreen = true
		m.applyViewColors(&view)
		return view
	}
	if m.workspace != nil && m.launch != nil {
		content, cursor := m.renderLaunchDialog()
		view := tea.NewView(content)
		view.AltScreen = true
		m.applyViewColors(&view)
		view.Cursor = cursor
		return view
	}
	if m.workspace != nil && !m.history.open && !m.help {
		view := tea.NewView(m.renderWorkspace())
		view.AltScreen = true
		m.applyViewColors(&view)
		view.MouseMode = tea.MouseModeAllMotion
		view.Cursor = m.workspaceCursor()
		return view
	}
	view := tea.NewView(m.renderFrame())
	view.AltScreen = true
	m.applyViewColors(&view)
	return view
}

func (m *Model) renderFrame() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	header := m.styles.title.Render("Agent Session Desk")
	count := m.styles.dim.Render(fmt.Sprintf("%d sessions", len(m.sessions)))
	if m.history.open {
		count = m.styles.dim.Render(fmt.Sprintf("%d history records", len(m.history.rows)))
	}
	gap := m.width - lipgloss.Width(header) - lipgloss.Width(count)
	if gap > m.tokens.layout.headerMinGap {
		header += strings.Repeat(" ", gap) + count
	}
	if m.height < m.tokens.layout.minFrameHeight {
		return fitFrame(header+"\n"+m.styles.dim.Render("Enlarge the window · q Quit"), m.width, m.height)
	}
	contentHeight := m.height - m.tokens.chromeRows() // header, breathing row, breathing row, footer
	var content string
	switch {
	case m.history.open:
		content = m.renderHistory(contentHeight)
	case m.help:
		content = m.renderHelp(contentHeight)
	case !m.previewVisible():
		content = m.renderSingleColumnLayout(contentHeight)
	case m.width < m.tokens.layout.stackedBelow:
		content = m.renderStackedLayout(contentHeight)
	default:
		content = m.renderDualColumnLayout(contentHeight)
	}
	footer := m.renderHelpBar()
	if m.history.open {
		footer = m.renderHistoryHelpBar()
	}
	if !m.history.open && m.status != "" {
		footer = m.styles.dim.Render(cellTruncate(label(m.status), max(0, m.width-m.tokens.layout.statusHelpWidth), m.tokens.marks.ellipsis)) + m.tokens.itemGap() + "? Help" + m.tokens.itemGap() + "q Quit"
	}
	if !m.history.open && m.listErr != nil {
		footer = m.styles.error.Render("Backend state unknown: "+label(m.listErr.Error())) + m.tokens.itemGap() + "r Retry"
	}
	return fitFrame(header+m.tokens.sectionBreak()+ensureExactHeight(content, contentHeight)+m.tokens.sectionBreak()+footer, m.width, m.height)
}

func (m *Model) renderPanelTitle(title string, width int) string {
	if width <= 0 {
		return ""
	}
	title = cellTruncate(title, width, m.tokens.marks.ellipsis)
	titleStyle := m.styles.panelTitle.Width(width)
	underlineStyle := m.styles.separator.Width(width)
	return titleStyle.Render(title) + "\n" + underlineStyle.Render(strings.Repeat(m.tokens.marks.rule, width))
}

// ensureExactHeight is retained from Agent Deck. Having exactly the requested
// number of rows prevents one panel from moving its neighbour when joined.
func ensureExactHeight(content string, n int) string {
	if n <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) > n {
		lines = lines[:n]
	} else {
		for len(lines) < n {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

// ensureExactWidth adapts Agent Deck's per-row width normalization. Truncate
// by display cells rather than bytes/runes so CJK and emoji stay in their pane.
func ensureExactWidth(content string, width int) string {
	if width <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) > width {
			line = cellTruncate(line, width, "")
		}
		padding := max(0, width-lipgloss.Width(line))
		lines[i] = line + strings.Repeat(" ", padding)
	}
	return strings.Join(lines, "\n")
}

func fitFrame(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	return ensureExactWidth(ensureExactHeight(content, height), width)
}

func (m *Model) renderDualColumnLayout(contentHeight int) string {
	dividerWidth := lipgloss.Width(m.tokens.divider())
	leftWidth := min(m.tokens.layout.listMax, max(m.tokens.layout.listMin, (m.width-dividerWidth)/m.tokens.layout.listFraction))
	rightWidth := m.width - leftWidth - dividerWidth
	panelContentHeight := max(0, contentHeight-m.tokens.layout.headingRows)
	leftPanel := m.renderPanelTitle("Sessions", leftWidth) + "\n" +
		ensureExactHeight(m.renderSessionList(leftWidth, panelContentHeight), panelContentHeight)
	rightPanel := m.renderPanelTitle("Screen preview", rightWidth) + "\n" +
		ensureExactHeight(m.renderPreviewPane(rightWidth, panelContentHeight), panelContentHeight)

	// Style the constant separator once, then repeat it for each exact row.
	separatorCell := m.styles.separator.Render(m.tokens.divider())
	separatorLines := make([]string, contentHeight)
	for i := range separatorLines {
		separatorLines[i] = separatorCell
	}
	leftPanel = fitFrame(leftPanel, leftWidth, contentHeight)
	rightPanel = fitFrame(rightPanel, rightWidth, contentHeight)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel,
		strings.Join(separatorLines, "\n"), rightPanel)
}

func (m *Model) renderStackedLayout(totalHeight int) string {
	// Unlike the upstream minimum floors, clamp both regions to the actual
	// available height; a short terminal must never overflow into the footer.
	listHeight := max(m.tokens.layout.stackedListMin, totalHeight/m.tokens.layout.stackedFraction)
	previewHeight := max(0, totalHeight-listHeight-m.tokens.layout.stackedGap)
	list := m.renderPanelTitle("Sessions", m.width) + "\n" +
		ensureExactHeight(m.renderSessionList(m.width, listHeight-m.tokens.layout.headingRows), listHeight-m.tokens.layout.headingRows)
	preview := m.renderPanelTitle("Screen preview", m.width) + "\n" +
		ensureExactHeight(m.renderPreviewPane(m.width, previewHeight-m.tokens.layout.headingRows), previewHeight-m.tokens.layout.headingRows)
	return ensureExactHeight(list, listHeight) + "\n" +
		ensureExactHeight(m.styles.separator.Render(strings.Repeat(m.tokens.marks.rule, m.width)), m.tokens.layout.stackedGap) + "\n" +
		ensureExactHeight(preview, previewHeight)
}

func (m *Model) renderSingleColumnLayout(totalHeight int) string {
	return m.renderPanelTitle("Sessions", m.width) + "\n" +
		ensureExactHeight(m.renderSessionList(m.width, totalHeight-m.tokens.layout.headingRows), totalHeight-m.tokens.layout.headingRows)
}

func (m *Model) renderPreviewPane(width, height int) string {
	s, ok := m.selected()
	if !ok {
		return m.styles.dim.Render("Select a session to see its current screen" + m.tokens.sectionBreak() + "? How to create a session")
	}
	if m.listErr != nil {
		return m.styles.dim.Render("Backend connection unconfirmed; press r to retry")
	}
	if m.previewErr != nil {
		return m.styles.error.Render("Preview unavailable:\n" + label(m.previewErr.Error()))
	}
	if s.TerminalState != session.TerminalAvailable && s.TerminalState != session.TerminalExited {
		return m.styles.dim.Render("Terminal" + terminalLabel(s.TerminalState) + "; CLI work state unknown")
	}
	if m.preview == "" {
		if m.previewBusy {
			return m.styles.dim.Render("Reading current screen…")
		}
		return m.styles.dim.Render("No text on the current screen")
	}
	// The capture is a snapshot, not durable history or a reconstructed
	// conversation. Keep its newest visible rows without retaining more copies.
	lines := strings.Split(m.preview, "\n")
	budget := max(0, height-1-m.tokens.layout.captionGap)
	clipped := len(lines) > budget
	if clipped {
		lines = lines[len(lines)-budget:]
	}
	caption := "Current screen · CLI work state unknown"
	if clipped || m.previewTruncated {
		caption = "Current screen excerpt · CLI state unknown"
	}
	if s.TerminalState == session.TerminalExited {
		caption = "Terminal exited · Screen retained, input disabled"
		if clipped || m.previewTruncated {
			caption = "Terminal exited · Retained screen excerpt"
		}
	}
	return m.styles.dim.Render(cellTruncate(caption, width, m.tokens.marks.ellipsis)) + strings.Repeat("\n", m.tokens.layout.captionGap+1) +
		strings.Join(lines, "\n")
}

func (m *Model) renderHelpBar() string {
	if m.width < m.tokens.layout.singleBelow {
		return m.styles.menuKey("↑↓", "Select") + m.tokens.itemGap() + m.styles.menuKey("↵", "Open") + m.tokens.itemGap() + "?" + m.tokens.itemGap() + "q"
	}
	parts := []string{
		m.styles.menuKey("↑↓", "Select"), m.styles.menuKey("Enter", "Open"),
		m.styles.menuKey("h", "History"), m.styles.menuKey("r", "Refresh"), m.styles.menuKey("?", "Help"), m.styles.menuKey("q", "Quit"),
	}
	return strings.Join(parts, m.tokens.itemGap())
}

func (m *Model) renderHelp(height int) string {
	content := m.renderPanelTitle("Getting started", m.width) + m.tokens.sectionBreak() +
		"Create a session in another terminal:\n" +
		m.styles.text.Render(m.tokens.itemGap()+"asd start --name Development -- codex") + "\n" +
		m.styles.text.Render(m.tokens.itemGap()+"asd start --name Review --cwd PATH -- PROGRAM ARGS") + m.tokens.sectionBreak() +
		"↑↓ / j k Select session; Enter Attach full screen.\n" +
		"Cancel IME composition and switch to English input before detaching.\n" +
		"Press Ctrl+b, release, then d to return.\n" +
		"h Read-only history; Enter Details, Esc Back.\n" +
		"q Quit the manager; background sessions continue." + m.tokens.sectionBreak() + "" +
		m.styles.dim.Render("Preview is the current screen only; CLI work state remains unknown.") + "\n" +
		"Esc / ? Back"
	return ensureExactHeight(content, height)
}
