// Adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/group_dialog.go and home.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: v2 inputs, stable
// product identities, bounded metadata effects and asynchronous reconciliation.
package ui

import (
	"context"
	"fmt"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type groupRenameService interface {
	RenameGroup(context.Context, session.Group, string) (session.Group, error)
}
type groupDialog struct {
	move, rename, busy, submitted bool
	groupTarget                   session.Group
	contextID, parentID           string
	target                        session.Session
	groups                        []session.Group
	inputs                        [2]textinput.Model
	focus, choice                 int
	err                           error
}

func (m *Model) beginGroupRename() tea.Cmd {
	if _, ok := m.service.(groupRenameService); !ok || m.workspace.transition != workspaceIdle || !m.sidebarGroupSelected() {
		return nil
	}
	m.groupIntent = &groupDialog{rename: true, groupTarget: m.selectedGroup()}
	return m.navigateWorkspace(workspaceGroup)
}

type groupResult struct {
	dialog  *groupDialog
	group   session.Group
	session session.Session
	err     error
}

func (m *Model) beginGroupDialog(move bool) tea.Cmd {
	if _, ok := m.service.(groupService); !ok || m.workspace.transition != workspaceIdle {
		return nil
	}
	d := &groupDialog{move: move, groups: append([]session.Group(nil), m.sidebarGroups()...)}
	if move {
		v, ok := m.selected()
		if !ok {
			return nil
		}
		d.target = v
		for i, g := range d.groups {
			if g.ID == session.GroupOf(v) {
				d.choice = i
			}
		}
	} else {
		d.contextID = m.selectedGroup().ID
		if m.sidebarGroupSelected() {
			d.parentID = d.contextID
		}
	}
	m.groupIntent = d
	return m.navigateWorkspace(workspaceGroup)
}
func (m *Model) openGroupDialog() tea.Cmd {
	d := m.groupIntent
	m.groupIntent = nil
	if d == nil {
		return m.connectWorkspace()
	}
	m.groupDialog = d
	m.workspace.focused = false
	if d.move {
		return nil
	}
	for i := range d.inputs {
		if d.rename && i > 0 {
			break
		}
		input := textinput.New()
		input.Placeholder = []string{"分组名称", "可选工作目录"}[i]
		input.CharLimit = session.MaxTitleInputBytes
		input.Prompt = m.tokens.rename.prompt
		input.SetVirtualCursor(false)
		input.SetStyles(m.styles.groupInput)
		input.KeyMap.Paste.SetEnabled(false)
		d.inputs[i] = input
	}
	if d.rename {
		d.inputs[0].SetStyles(m.styles.renameInput)
		d.inputs[0].SetValue(d.groupTarget.Name)
		d.inputs[0].CursorEnd()
	}
	m.sizeGroupDialog()
	return d.inputs[0].Focus()
}
func (m *Model) groupDialogEvent(msg tea.Msg) (tea.Cmd, bool) {
	if r, ok := msg.(groupResult); ok {
		if m.groupDialog != r.dialog {
			return nil, true
		}
		d := m.groupDialog
		d.busy, d.err = false, r.err
		if r.err != nil {
			d.err = fmt.Errorf("请返回并刷新核对分组后再操作：%w", r.err)
			return nil, true
		}
		m.listGeneration++
		if m.listCancel != nil {
			m.listCancel()
		}
		if d.move {
			m.acceptGroupMembership(r.session)
			m.acceptSessionName(r.session)
			m.selectWorkspaceSession(r.session)
		} else {
			groups := append([]session.Group(nil), m.sidebarGroups()...)
			found := false
			for i, g := range groups {
				if g.ID == r.group.ID {
					if g.Revision <= r.group.Revision {
						groups[i] = r.group
					}
					found = true
				}
			}
			if !found {
				groups = append(groups, r.group)
			}
			m.groups = groups
			m.cursor, m.selectedGroupID = -1, r.group.ID
			if !d.rename {
				m.revealGroup(r.group.ParentID)
			}
		}
		m.groupDialog = nil
		m.status = ""
		return tea.Batch(m.connectWorkspace(), m.loadSessions()), true
	}
	d := m.groupDialog
	if d == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		m.sizeGroupDialog()
		return nil, true
	case tea.KeyReleaseMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
		return nil, true
	case tea.KeyPressMsg:
		if d.busy {
			return nil, true
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			m.groupDialog = nil
			return tea.Batch(m.connectWorkspace(), m.loadSessions()), true
		case "enter":
			return m.submitGroupDialog(), true
		}
		if d.submitted {
			return nil, true
		}
		if d.move {
			switch msg.String() {
			case "up", "k":
				d.choice = max(0, d.choice-1)
			case "down", "j":
				d.choice = min(len(d.groups)-1, d.choice+1)
			}
			return nil, true
		}
		if d.rename {
			switch msg.String() {
			case "tab", "shift+tab", "up", "down":
				return nil, true
			}
		}
		switch msg.String() {
		case "tab", "shift+tab":
			d.inputs[d.focus].Blur()
			d.focus = 1 - d.focus
			return d.inputs[d.focus].Focus(), true
		case "up", "down":
			if d.contextID != "" {
				if d.parentID == "" {
					d.parentID = d.contextID
				} else {
					d.parentID = ""
				}
			}
			return nil, true
		case "ctrl+pgup", "ctrl+pgdown", "f6":
			return nil, true
		}
		if !renameInputSafe(d.inputs[d.focus].Value(), msg.Text) {
			return nil, true
		}
		var cmd tea.Cmd
		d.inputs[d.focus], cmd = d.inputs[d.focus].Update(msg)
		d.err = nil
		return cmd, true
	case tea.PasteMsg:
		if d.busy || d.submitted || d.move {
			return nil, true
		}
		if !renameInputSafe(d.inputs[d.focus].Value(), msg.Content) {
			d.err = fmt.Errorf("分组字段只接受有界单行文本")
			return nil, true
		}
		var cmd tea.Cmd
		d.inputs[d.focus], cmd = d.inputs[d.focus].Update(msg)
		d.err = nil
		return cmd, true
	}
	return nil, false
}
func (m *Model) submitGroupDialog() tea.Cmd {
	d := m.groupDialog
	if d.busy || d.submitted {
		return nil
	}
	var name, dir, target string
	var err error
	if d.move {
		if d.choice < 0 || d.choice >= len(d.groups) {
			return nil
		}
		target = d.groups[d.choice].ID
		found := false
		for _, v := range m.sessions {
			if sameRun(v, d.target) && v.BackendID == d.target.BackendID && v.Command == d.target.Command {
				found = true
			}
		}
		if !found {
			d.err = fmt.Errorf("原 Session/Run 已变化；请返回后重新选择")
			return nil
		}
	} else {
		name, err = session.NormalizeGroupName(d.inputs[0].Value())
		if err == nil && !d.rename {
			dir, err = session.NormalizeGroupDir(d.inputs[1].Value())
		}
		if err != nil {
			d.err = err
			return nil
		}
	}
	d.busy, d.submitted, d.err = true, true, nil
	ctx, parent, expected, groupTarget := m.ctx, d.parentID, d.target, d.groupTarget
	service := m.service
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if d.move {
			v, err := service.(groupService).MoveToGroup(ctx, expected, target)
			return groupResult{dialog: d, session: v, err: err}
		}
		if d.rename {
			g, err := service.(groupRenameService).RenameGroup(ctx, groupTarget, name)
			return groupResult{dialog: d, group: g, err: err}
		}
		g, err := service.(groupService).CreateGroup(ctx, parent, name, dir)
		return groupResult{dialog: d, group: g, err: err}
	}
}
func (m *Model) acceptGroupMembership(moved session.Session) {
	apply := func(v *session.Session) {
		if v.ID == moved.ID && v.GroupRevision <= moved.GroupRevision {
			v.GroupID, v.GroupRevision = moved.GroupID, moved.GroupRevision
		}
	}
	for i := range m.sessions {
		apply(&m.sessions[i])
	}
	for i := range m.workspace.tabs {
		apply(&m.workspace.tabs[i].session)
	}
}
