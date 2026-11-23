// Confirmation interaction adapted from Agent Deck, Copyright (c) 2025 Ashesh
// Goplani. SPDX-License-Identifier: MIT
// Source: confirm_dialog.go, home.go and home_keys_dialogs.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: read-only impact preview,
// stable IDs, metadata CAS, original input drain and asynchronous reconciliation.
package ui

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type groupRemovalService interface {
	PreviewGroupRemoval(context.Context, string) (session.GroupRemoval, error)
	RemoveGroup(context.Context, session.GroupRemoval, []session.Session) (session.Workspace, error)
}
type groupRemovalDialog struct {
	target                   session.Group
	plan                     session.GroupRemoval
	focus                    int
	loading, busy, submitted bool
	err                      error
	cancel                   context.CancelFunc
	observed                 uint64
}
type groupRemovalPreview struct {
	dialog *groupRemovalDialog
	plan   session.GroupRemoval
	err    error
}
type groupRemovalResult struct {
	dialog *groupRemovalDialog
	view   session.Workspace
	err    error
}

func (m *Model) beginGroupRemoval() tea.Cmd {
	if _, ok := m.service.(groupRemovalService); !ok || !m.sidebarGroupSelected() || m.workspace.transition != workspaceIdle || m.groupOrder != nil {
		return nil
	}
	m.groupRemovalIntent = &groupRemovalDialog{target: m.selectedGroup(), focus: 1}
	return m.navigateWorkspace(workspaceRemoval)
}
func (m *Model) openGroupRemoval() tea.Cmd {
	d := m.groupRemovalIntent
	m.groupRemovalIntent = nil
	if d == nil {
		return m.connectWorkspace()
	}
	m.groupRemoval = d
	m.workspace.focused = false
	if d.target.ID == session.DefaultGroupID {
		return nil
	}
	d.loading = true
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	d.cancel = cancel
	service := m.service.(groupRemovalService)
	id := d.target.ID
	return func() tea.Msg {
		defer cancel()
		plan, err := service.PreviewGroupRemoval(ctx, id)
		return groupRemovalPreview{d, plan, err}
	}
}
func (m *Model) closeGroupRemoval() tea.Cmd {
	if d := m.groupRemoval; d != nil && d.cancel != nil {
		d.cancel()
	}
	m.groupRemoval = nil
	return tea.Batch(m.connectWorkspace(), m.loadSessions())
}
func (m *Model) groupRemovalEvent(msg tea.Msg) (tea.Cmd, bool) {
	switch result := msg.(type) {
	case groupRemovalPreview:
		if m.groupRemoval != result.dialog {
			return nil, true
		}
		d := m.groupRemoval
		d.loading, d.err = false, result.err
		if result.err == nil {
			if result.plan.Root.ID != d.target.ID {
				d.err = fmt.Errorf("预览目标变化，请返回刷新")
			} else {
				d.plan = result.plan
			}
		}
		return nil, true
	case groupRemovalResult:
		if m.groupRemoval != result.dialog {
			return nil, true
		}
		d := m.groupRemoval
		d.busy = false
		m.listGeneration++
		if m.listCancel != nil {
			m.listCancel()
		}
		if result.err != nil {
			d.err = fmt.Errorf("结果未确认，请返回刷新核对：%w", result.err)
			return m.loadSessions(), true
		}
		removed := map[string]bool{}
		for _, g := range d.plan.Groups {
			removed[g.ID] = true
		}
		var retained []session.Group
		for _, g := range m.sidebarGroups() {
			if !removed[g.ID] {
				retained = append(retained, g)
			}
		}
		m.groups = retained
		m.acceptGroupOrder(result.view.Groups, m.groupsObserved == d.observed)
		for _, v := range result.view.Sessions {
			m.acceptGroupMembership(v)
			m.acceptSessionName(v)
		}
		m.cursor, m.selectedGroupID = -1, session.DefaultGroupID
		preferences := m.forgetRemovedGroupFolds(removed)
		m.groupRemoval = nil
		return tea.Batch(preferences, m.connectWorkspace(), m.loadSessions()), true
	}
	d := m.groupRemoval
	if d == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		return nil, true
	case tea.KeyReleaseMsg, tea.PasteMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
		return nil, true
	case tea.KeyPressMsg:
		if d.busy {
			return nil, true
		}
		key := msg.String()
		if d.target.ID == session.DefaultGroupID {
			if key == "enter" || key == "esc" || key == "o" || key == "O" || key == "ctrl+c" {
				return m.closeGroupRemoval(), true
			}
			return nil, true
		}
		switch key {
		case "esc", "n", "N", "ctrl+c":
			return m.closeGroupRemoval(), true
		case "y", "Y":
			return m.submitGroupRemoval(), true
		case "enter":
			if d.focus == 1 {
				return m.closeGroupRemoval(), true
			}
			return m.submitGroupRemoval(), true
		case "left", "h":
			d.focus = 0
		case "right", "l", "tab":
			d.focus = 1
		}
		return nil, true
	}
	return nil, false
}
func (m *Model) submitGroupRemoval() tea.Cmd {
	d := m.groupRemoval
	if d == nil || d.loading || d.busy || d.submitted || d.plan.Stamp == "" || d.target.ID == session.DefaultGroupID {
		return nil
	}
	if _, fits := m.groupRemovalBox(); !fits {
		d.err = fmt.Errorf("请放大窗口核对完整影响范围")
		return nil
	}
	d.busy, d.submitted, d.err = true, true, nil
	d.observed = m.groupsObserved
	service, ctx, plan := m.service.(groupRemovalService), m.ctx, d.plan
	observed := append([]session.Session(nil), m.sessions...)
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		view, err := service.RemoveGroup(ctx, plan, observed)
		return groupRemovalResult{d, view, err}
	}
}
