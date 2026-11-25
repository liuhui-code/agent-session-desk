// Session-row ordering keys adapted from Agent Deck, Copyright (c) 2025
// Ashesh Goplani. SPDX-License-Identifier: MIT
// Source: internal/ui/home.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: asynchronous scoped snapshots, no terminal effects and a quit barrier.
package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"strings"
	"time"
)

type sessionOrderService interface {
	ReorderSession(context.Context, session.SessionOrder, int) (session.Workspace, error)
}
type sessionOrderMutation struct {
	quitting       bool
	groupID, stamp string
}
type sessionOrderResult struct {
	operation *sessionOrderMutation
	view      session.Workspace
	err       error
}

func (m *Model) reorderSelectedSession(delta int) tea.Cmd {
	if m.sidebarGroupSelected() {
		return m.reorderSelectedGroup(delta)
	}
	service, ok := m.service.(sessionOrderService)
	if !ok || m.sessionOrder != nil || m.groupOrder != nil || m.listErr != nil || m.workspace.transition != workspaceIdle {
		return nil
	}
	if m.sessionOrderFailed {
		m.status = "Session order unconfirmed; refresh before sorting again"
		return nil
	}
	selected, ok := m.selected()
	if !ok {
		return nil
	}
	peers := session.OrderPeers(m.sessions, selected.ID)
	index := -1
	for i, v := range peers {
		if sameRun(v, selected) {
			index = i
		}
	}
	if index < 0 || index+delta < 0 || index+delta >= len(peers) {
		return nil
	}
	operation := &sessionOrderMutation{groupID: session.GroupOf(selected), stamp: m.orderStamps[session.GroupOf(selected)]}
	m.sessionOrder = operation
	m.status = "Saving session order…"
	expected := session.SessionOrder{TargetID: selected.ID, Siblings: peers, Stamp: operation.stamp}
	ctx := m.ctx
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		view, err := service.ReorderSession(ctx, expected, delta)
		return sessionOrderResult{operation, view, err}
	}
}

func (m *Model) sessionOrderEvent(msg tea.Msg) (tea.Cmd, bool) {
	if result, ok := msg.(sessionOrderResult); ok {
		if m.sessionOrder == nil || m.sessionOrder != result.operation {
			return nil, true
		}
		quitting := m.sessionOrder.quitting
		m.sessionOrder = nil
		m.listGeneration++
		if m.listCancel != nil {
			m.listCancel()
		}
		if result.err != nil {
			m.sessionOrderFailed = true
			m.status = "Session order unconfirmed; refresh to verify: " + label(result.err.Error())
			return m.loadSessions(), true
		}
		m.acceptSessionOrder(result.view.Sessions)
		if m.orderStamps[result.operation.groupID] == result.operation.stamp {
			if m.orderStamps == nil {
				m.orderStamps = map[string]string{}
			}
			m.orderStamps[result.operation.groupID] = result.view.OrderStamps[result.operation.groupID]
		}
		if strings.HasPrefix(m.status, "Saving session order") {
			m.status = ""
		}
		if quitting {
			return m.workspaceKey(tea.KeyPressMsg{Code: 'q', Text: "q"}), true
		}
		return m.loadSessions(), true
	}
	if m.sessionOrder != nil && m.sessionOrder.quitting {
		switch msg.(type) {
		case tea.KeyPressMsg, tea.KeyReleaseMsg, tea.PasteMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
			return nil, true
		}
	}
	return nil, false
}

func (m *Model) acceptSessionOrder(rows []session.Session) {
	selected, hadSelection := m.selected()
	byID := map[string]session.Session{}
	for _, row := range rows {
		byID[row.ID] = row
	}
	for i, v := range m.sessions {
		incoming, ok := byID[v.ID]
		if ok && sameRun(v, incoming) && v.BackendID == incoming.BackendID && v.Command == incoming.Command && v.Title == incoming.Title && session.GroupOf(v) == session.GroupOf(incoming) && v.GroupRevision == incoming.GroupRevision && v.TitleRevision == incoming.TitleRevision && incoming.OrderRevision > v.OrderRevision {
			m.sessions[i].Order, m.sessions[i].OrderRevision = incoming.Order, incoming.OrderRevision
		}
	}
	session.SortSessions(m.sessions)
	if hadSelection {
		for i, v := range m.sessions {
			if sameRun(v, selected) {
				m.cursor = i
				break
			}
		}
	}
	m.workspace.listClick = workspaceListClick{}
}
