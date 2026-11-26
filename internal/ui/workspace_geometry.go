package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type workspaceGeometry struct {
	x, y, width, height, left, outerWidth, bodyHeight, above, terminalHeight, listWidth, listHeight int
	bodyX, bodyY, bodyWidth, listX, listY                                                           int
}

func (m *Model) workspaceGeometry() workspaceGeometry {
	t := m.tokens
	inset := min(t.workspace.sideInset, max(0, (m.width-t.workspace.minPane)/2))
	top := min(t.workspace.topInsetRows, max(0, m.height-1))
	footer := 0
	if !m.workspace.focused {
		footer = t.workspace.footerRevealRows
	}
	bodyWidth := max(0, m.width-2*inset)
	bodyHeight := max(0, m.height-top-footer)
	left, gap := 0, lipgloss.Width(t.divider())
	if !m.workspace.sidebarHidden && bodyWidth >= t.layout.stackedBelow {
		left = max(t.workspace.minPane, bodyWidth*m.workspaceSidebarPercent()/100)
		left = min(left, max(0, bodyWidth-gap-t.workspace.minPane))
	} else {
		gap = 0
	}
	outer := max(0, bodyWidth-left-gap)
	style := m.styles.terminalFrame
	above, terminalHeight := 0, bodyHeight
	minList := t.layout.stackedListMin
	minFrame := style.GetVerticalFrameSize() + t.workspace.minTerminalRows
	if !m.workspace.sidebarHidden && bodyWidth < t.layout.stackedBelow && !m.workspace.focused {
		above, terminalHeight = bodyHeight, 0
		if bodyWidth >= t.layout.singleBelow && bodyHeight >= minList+t.layout.stackedGap+minFrame {
			listHeight := max(minList, bodyHeight*m.workspaceSidebarPercent()/100)
			listHeight = min(listHeight, bodyHeight-t.layout.stackedGap-minFrame)
			above = listHeight + t.layout.stackedGap
			terminalHeight = bodyHeight - above
		}
	}
	listWidth, listHeight := left, bodyHeight
	if above > 0 {
		listWidth, listHeight = bodyWidth, above
		if terminalHeight > 0 {
			listHeight -= t.layout.stackedGap
		}
	}
	return workspaceGeometry{
		x: inset + left + gap + style.GetBorderLeftSize(), y: top + above + style.GetBorderTopSize(),
		width: max(0, outer-style.GetHorizontalFrameSize()), height: max(0, terminalHeight-style.GetVerticalFrameSize()),
		left: left, outerWidth: outer, bodyHeight: bodyHeight, above: above, terminalHeight: terminalHeight, listWidth: listWidth, listHeight: listHeight,
		bodyX: inset, bodyY: top, bodyWidth: bodyWidth, listX: inset, listY: top,
	}
}

func (m *Model) padWorkspaceBody(body string, g workspaceGeometry) string {
	lines := strings.Split(fitFrame(body, g.bodyWidth, g.bodyHeight), "\n")
	frame := make([]string, m.height)
	for y := range frame {
		frame[y] = strings.Repeat(" ", m.width)
	}
	for y := 0; y < g.bodyHeight; y++ {
		frame[g.bodyY+y] = strings.Repeat(" ", g.bodyX) + lines[y]
	}
	return fitFrame(strings.Join(frame, "\n"), m.width, m.height)
}
