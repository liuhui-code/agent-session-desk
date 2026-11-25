package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func (m *Model) workspaceKey(msg tea.KeyPressMsg) tea.Cmd {
	return tea.Batch(m.finishProgramPointer(), m.finishWorkspaceSplit(), m.workspaceKeyNow(msg))
}

func (m *Model) workspaceKeyNow(msg tea.KeyPressMsg) tea.Cmd {
	w := m.workspace
	w.copyPointer = nil
	w.listClick = workspaceListClick{}
	if msg.Code == 's' && msg.Mod == tea.ModCtrl {
		return m.toggleWorkspaceSidebar()
	}
	if m.quickCreate != nil {
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quickCreate.quitting = true
			m.status = "Finishing creation before quitting…"
		}
		return nil
	}

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
			m.status = "Switching tabs; wait before changing focus"
			return nil
		}
		if len(w.tabs) == 0 || w.err != nil {
			w.focused = false
			m.status = ""
			return nil
		}
		w.focused = !w.focused
		if !w.focused {
			w.sidebarHidden = false
		}
		m.status = ""
		return m.workspaceSize()
	}
	if w.focused && len(w.tabs) > 0 {
		return m.workspaceTerminalKey(msg)
	}
	if m.groupOrder != nil || m.sessionOrder != nil {
		switch msg.String() {
		case "shift+up", "ctrl+up", "+", "K", "shift+down", "ctrl+down", "-", "J", "g", "M", "r", "n", "a", "h", "?", "S", "d":
			m.status = "Saving order; wait before continuing"
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
	case "V":
		return m.copyVisiblePane()
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
			m.status = "Terminal is not ready for selection/search"
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
			m.status = "Saving session order before quitting…"
			return nil
		}
		if m.groupOrder != nil {
			m.groupOrder.quitting = true
			m.status = "Saving group order before quitting…"
			return nil
		}
		if m.preferences.busy {
			w.transition = workspaceQuit
			m.status = "Saving sidebar preferences before quitting…"
			return nil
		}
		if w.pending > 0 || (w.busy && len(w.buffered) > 0) {
			w.transition = workspaceQuit
			m.status = "Finishing input before quitting…"
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
			m.status = "Tab is connecting or switching; wait before opening another terminal"
			return nil
		}
		selected, ok := m.selected()
		if !ok {
			return nil
		}
		if selected.TerminalState != session.TerminalAvailable {
			m.status = "Target terminal is not interactive"
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
	case "N":
		return m.navigateWorkspace(workspaceLaunchEdit)
	case "?":
		return m.navigateWorkspace(workspaceHelp)
	case "S":
		return m.navigateWorkspace(workspaceSettings)
	}
	return nil
}
