package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) workspaceFooterVisible() bool {
	w := m.workspace
	_, prompt := m.workspaceCopyPrompt()
	return w.footerHovered || m.status != "" || w.err != nil || m.listErr != nil || prompt != nil
}

func (m *Model) workspaceFooterHit(x, y int) bool {
	return x >= 0 && x < m.width && y >= m.height-m.tokens.workspace.footerRevealRows && y < m.height && m.workspaceFooterVisible()
}

func (m *Model) workspaceFooterMotion(msg tea.MouseMotionMsg) bool {
	w := m.workspace
	// An existing native gesture always retains its original delivery path.
	if msg.Button != tea.MouseNone || msg.Mod != 0 || w.programPointer != nil || w.copyPointer != nil || w.split.dragging {
		w.footerHovered = false
		return false
	}
	w.footerHovered = msg.X >= 0 && msg.X < m.width && msg.Y >= m.height-m.tokens.workspace.footerRevealRows && msg.Y < m.height
	if w.footerHovered {
		w.programHover.latest = nil
	}
	return w.footerHovered
}

func (m *Model) renderWorkspaceOverlay(body string) string {
	if !m.workspaceFooterVisible() || m.height < 1 {
		return body
	}
	w := m.workspace
	text := m.tokens.workspace.terminalHint
	if !w.focused {
		text = ansi.Strip(m.renderManagementFooter())
	} else if w.screen.CopyMode {
		text = m.tokens.workspace.copyHint
		if w.screen.Status != "" {
			text = "滚动历史 · " + label(w.screen.Status)
		}
	}
	if m.status != "" {
		text = label(m.status)
	}
	if m.listErr != nil {
		text = "后台状态未知：" + label(m.listErr.Error())
	}
	if w.err != nil {
		text = "终端不可交互 · r 重查：" + label(w.err.Error())
		if w.screen.Width > 0 {
			text = "只读：上次观察屏幕；连接不可用，r 重查。" + label(w.err.Error())
		}
	}
	if prompt, cursor := m.workspaceCopyPrompt(); cursor != nil {
		text = prompt
	}
	lines := strings.Split(body, "\n")
	lines[m.height-1] = m.styles.dim.Background(m.tokens.colors.bg).Render(ensureExactWidth(cellTruncate(text, m.width, ""), m.width))
	return strings.Join(lines, "\n")
}
