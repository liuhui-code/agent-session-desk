package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func (m *Model) closeWorkspaceTabNow() tea.Cmd {
	w := m.workspace
	if len(w.tabs) == 0 {
		return nil
	}
	w.transition = workspaceIdle
	w.disconnect()
	w.tabs = append(w.tabs[:w.active], w.tabs[w.active+1:]...)
	w.active = min(w.active, max(0, len(w.tabs)-1))
	w.err = nil
	w.screen = terminal.Screen{}
	w.buffered = nil
	w.transition = workspaceIdle
	m.status = "Tab 已关闭；运行实例继续保留，可从会话列表重新打开"
	if len(w.tabs) == 0 {
		return nil
	}
	return m.activateWorkspaceNow(w.active)
}

func (m *Model) closeWorkspaceTab() tea.Cmd {
	w := m.workspace
	if w.pending > 0 || (w.busy && len(w.buffered) > 0) {
		w.transition = workspaceClose
		m.status = "正在完成已输入的内容后关闭 Tab…"
		return nil
	}
	return m.closeWorkspaceTabNow()
}

func (m *Model) closeWorkspaceView(index int) tea.Cmd {
	w := m.workspace
	if index < 0 || index >= len(w.tabs) {
		return nil
	}
	if index == w.active {
		return m.closeWorkspaceTab()
	}
	w.tabs = append(w.tabs[:index], w.tabs[index+1:]...)
	if index < w.active {
		w.active--
	}
	m.status = "Tab 已关闭；运行实例继续保留，可从会话列表重新打开"
	return nil
}
