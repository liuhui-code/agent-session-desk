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
	if cmd, handled := m.workspaceSplitKey(msg); handled {
		return cmd
	}
	switch msg.String() {
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
	case "home", "g":
		return m.move(-len(m.sessions))
	case "end", "G":
		return m.move(len(m.sessions))
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
		w.err = nil
		m.status = ""
		if w.conn == nil {
			return tea.Batch(m.connectWorkspace(), m.loadSessions())
		}
		return m.workspaceWrite(workspaceInput{kind: "refresh"})
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
