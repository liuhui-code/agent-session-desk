package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func (m *Model) workspaceKey(msg tea.KeyPressMsg) tea.Cmd {
	w := m.workspace
	w.tabDrag = workspaceTabDrag{}
	w.listClick = workspaceListClick{}
	if w.transition != workspaceIdle && w.transition != workspaceSwitch {
		return nil
	}
	if msg.Mod&tea.ModCtrl != 0 && msg.Mod&(tea.ModAlt|tea.ModShift) == 0 {
		switch msg.Code {
		case tea.KeyPgDown:
			return m.activateWorkspace(w.active + 1)
		case tea.KeyPgUp:
			return m.activateWorkspace(w.active - 1)
		}
	}
	if msg.Code == tea.KeyF6 && msg.Mod == 0 {
		if w.transition == workspaceSwitch {
			m.status = "正在切换 Tab；完成后可改变输入焦点"
			return nil
		}
		if len(w.tabs) == 0 || w.err != nil {
			w.focused = false
			m.status = ""
			return nil
		}
		w.focused = !w.focused
		m.status = ""
		return m.workspaceSize()
	}
	if w.focused && len(w.tabs) > 0 {
		return m.workspaceTerminalKey(msg)
	}
	if m.groupOrder != nil || m.sessionOrder != nil {
		switch msg.String() {
		case "shift+up", "ctrl+up", "+", "K", "shift+down", "ctrl+down", "-", "J", "g", "M", "r", "n", "a", "h", "?", "S", "d":
			m.status = "管理模式 · 正在保存排序；完成后可继续操作"
			return nil
		}
	}
	if cmd, handled := m.sidebarNavigationKey(msg); handled {
		return cmd
	}
	if cmd, handled := m.workspaceSplitKey(msg); handled {
		return cmd
	}
	switch msg.String() {
	case "d":
		if m.sidebarGroupSelected() {
			return m.beginGroupRemoval()
		}
		return m.beginSessionStop()
	case "shift+up", "ctrl+up", "+", "K":
		return m.reorderSelectedSession(-1)
	case "shift+down", "ctrl+down", "-", "J":
		return m.reorderSelectedSession(1)
	case "/":
		return m.navigateWorkspace(workspaceSearch)
	case "v":
		if _, ok := w.conn.(terminal.CopyModeConnection); !ok || w.busy || w.err != nil || w.transition != workspaceIdle {
			m.status = "当前终端尚不可进入选择／搜索"
			return nil
		}
		w.focused = true
		m.status = ""
		return tea.Batch(m.workspaceSize(), m.workspaceWrite(workspaceInput{kind: "copy-mode"}))
	case "[", "]":
		if len(w.tabs) > 0 {
			destination := w.active - 1
			if msg.String() == "]" {
				destination = w.active + 1
			}
			m.moveWorkspaceTab(w.tabs[w.active].viewID, destination)
		}
		return nil
	case "q", "ctrl+c":
		if m.sessionOrder != nil {
			m.sessionOrder.quitting = true
			m.status = "管理模式 · 正在保存会话顺序后退出…"
			return nil
		}
		if m.groupOrder != nil {
			m.groupOrder.quitting = true
			m.status = "管理模式 · 正在保存分组顺序后退出…"
			return nil
		}
		if m.preferences.busy {
			w.transition = workspaceQuit
			m.status = "管理模式 · 正在保存已选择的侧栏偏好后退出…"
			return nil
		}
		if w.pending > 0 || (w.busy && len(w.buffered) > 0) {
			w.transition = workspaceQuit
			m.status = "正在完成已输入的内容后退出…"
			return nil
		}
		m.Close()
		return tea.Quit
	case "up", "k":
		return m.move(-1)
	case "down", "j":
		return m.move(1)
	case "g":
		if _, ok := m.service.(groupService); ok {
			return m.beginGroupDialog(false)
		}
		fallthrough
	case "home":
		return m.move(-len(m.sessions) - len(m.sidebarGroups()))
	case "M":
		return m.beginGroupDialog(true)
	case "end", "G":
		return m.move(len(m.sessions) + len(m.sidebarGroups()))
	case "x":
		return m.closeWorkspaceTab()
	case "enter":
		if w.transition == workspaceSwitch {
			selected, ok := m.selected()
			for _, tab := range w.tabs {
				if ok && tab.viewID == w.switchTarget && sameRun(tab.session, selected) {
					w.focused = true
					return m.workspaceSize()
				}
			}
		}
		if w.busy || w.transition != workspaceIdle {
			m.status = "当前 Tab 正在连接或切换；完成后可打开所选终端"
			return nil
		}
		selected, ok := m.selected()
		if !ok {
			return nil
		}
		if selected.TerminalState != session.TerminalAvailable {
			m.status = "目标终端当前不可交互"
			return nil
		}
		for i, tab := range w.tabs {
			if sameRun(tab.session, selected) {
				w.focused = true
				if i == w.active && w.conn != nil {
					return m.workspaceSize()
				}
				return m.activateWorkspace(i)
			}
		}
		w.nextID++
		w.tabs = append(w.tabs, workspaceTab{viewID: w.nextID, session: selected, live: true})
		w.focused = true
		return m.activateWorkspace(len(w.tabs) - 1)
	case "r":
		if w.err == nil && m.listErr == nil {
			if m.sidebarGroupSelected() {
				return m.beginGroupRename()
			}
			return m.beginSessionRename()
		}
		fallthrough
	case "ctrl+r":
		w.err = nil
		m.status = ""
		if w.conn == nil {
			return tea.Batch(m.connectWorkspace(), m.loadSessions())
		}
		return tea.Batch(m.workspaceWrite(workspaceInput{kind: "refresh"}), m.loadSessions())
	case "a":
		return m.navigateWorkspace(workspaceAttach)
	case "h":
		return m.navigateWorkspace(workspaceHistory)
	case "n":
		return m.navigateWorkspace(workspaceLaunch)
	case "?":
		return m.navigateWorkspace(workspaceHelp)
	case "S":
		return m.navigateWorkspace(workspaceSettings)
	}
	return nil
}
