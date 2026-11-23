package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type workspaceTabCell struct {
	index, x, width int
	closeX          int
	text            string
}

func (m *Model) workspaceTabCells() []workspaceTabCell {
	w := m.workspace
	available := m.workspaceTabBudget()
	labels := make([]string, len(w.tabs))
	closeOffsets := make([]int, len(w.tabs))
	peers := m.titlePeers()
	for i, tab := range w.tabs {
		style := m.styles.tab
		if i == w.active {
			style = m.styles.tabActive
		}
		titleBudget := max(0, min(m.tokens.workspace.maxTabWidth, available)-style.GetHorizontalFrameSize()-m.tokens.layout.menuGap-lipgloss.Width(m.tokens.marks.close))
		title := m.sessionTitle(tab.session, peers, titleBudget)
		labels[i] = style.Render(title + strings.Repeat(" ", m.tokens.layout.menuGap) + m.tokens.marks.close)
		closeOffsets[i] = lipgloss.Width(labels[i]) - style.GetPaddingRight() - lipgloss.Width(m.tokens.marks.close)
	}
	start := w.active
	for start > 0 && lipgloss.Width(strings.Join(labels[start-1:w.active+1], m.tokens.itemGap())) <= available {
		start--
	}
	var cells []workspaceTabCell
	x := 0
	for i := start; i < len(labels) && x < available; i++ {
		width := min(lipgloss.Width(labels[i]), available-x)
		closeX := -1
		if width == lipgloss.Width(labels[i]) {
			closeX = x + closeOffsets[i]
		}
		cells = append(cells, workspaceTabCell{index: i, x: x, width: width, closeX: closeX, text: cellTruncate(labels[i], width, m.tokens.marks.ellipsis)})
		x += width + m.tokens.layout.itemGap
	}
	return cells
}
func (m *Model) workspaceNewTabCell() workspaceTabCell {
	text := m.styles.tab.Render(m.tokens.marks.add)
	width := min(m.width, lipgloss.Width(text))
	return workspaceTabCell{index: -1, x: max(0, m.width-width), width: width, text: cellTruncate(text, width, "")}
}
func (m *Model) workspaceTabBudget() int {
	return max(0, m.workspaceNewTabCell().x-m.tokens.layout.itemGap)
}
func (m *Model) workspaceTabs() string {
	var labels []string
	for _, cell := range m.workspaceTabCells() {
		labels = append(labels, cell.text)
	}
	button := m.workspaceNewTabCell()
	return ensureExactWidth(strings.Join(labels, m.tokens.itemGap()), button.x) + button.text
}
