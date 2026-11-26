package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) workspaceFooterVisible() bool {
	w := m.workspace
	_, prompt := m.workspaceCopyPrompt()
	return !w.focused || m.status != "" || w.err != nil || m.listErr != nil || prompt != nil || (w.screen.CopyMode && w.screen.Status != "")
}

func (m *Model) workspaceFooterHit(x, y int) bool {
	g := m.workspaceGeometry()
	return x >= g.bodyX && x < g.bodyX+g.bodyWidth && y >= m.height-m.tokens.workspace.footerRevealRows && y < m.height && m.workspaceFooterVisible()
}

func (m *Model) workspaceFooterMotion(msg tea.MouseMotionMsg) bool {
	w := m.workspace
	w.footerHovered = false
	if msg.Button != tea.MouseNone || msg.Mod != 0 || w.programPointer != nil || w.copyPointer != nil || w.split.dragging {
		return false
	}
	if m.workspaceFooterHit(msg.X, msg.Y) {
		w.programHover.latest = nil
		return true
	}
	return false
}

func (m *Model) renderWorkspaceOverlay(body string) string {
	if !m.workspaceFooterVisible() || m.height < 1 {
		return body
	}
	w := m.workspace
	text := m.tokens.workspace.terminalHint
	if !w.focused {
		view := *m
		view.width = m.workspaceGeometry().bodyWidth
		text = view.renderManagementFooter()
	} else if w.screen.CopyMode {
		text = m.tokens.workspace.copyHint
		if w.screen.Status != "" {
			text = "History · " + label(w.screen.Status)
		}
	}
	if m.status != "" {
		text = label(m.status)
	}
	if m.listErr != nil {
		text = "Backend state unknown: " + label(m.listErr.Error())
	}
	if w.err != nil {
		text = "Terminal not interactive · r Retry: " + label(w.err.Error())
		if w.screen.Width > 0 {
			text = "Read-only last observed screen; disconnected. r Retry. " + label(w.err.Error())
		}
	}
	if prompt, cursor := m.workspaceCopyPrompt(); cursor != nil {
		text = prompt
	}
	lines := strings.Split(body, "\n")
	g := m.workspaceGeometry()
	line := strings.Repeat(" ", g.bodyX) + ensureExactWidth(cellTruncate(text, g.bodyWidth, ""), g.bodyWidth)
	lines[m.height-1] = m.styles.dim.Background(m.tokens.colors.bg).Render(ensureExactWidth(line, m.width))
	return strings.Join(lines, "\n")
}
