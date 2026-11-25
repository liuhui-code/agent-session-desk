// Ordering interaction adapted from Agent Deck, Copyright (c) 2025 Ashesh
// Goplani. SPDX-License-Identifier: MIT
// Source: internal/ui/home.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: stable product identities, asynchronous metadata CAS, bounded
// single-flight writes and independent terminal/quit barriers.
package ui

import (
	"context"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type groupOrderService interface {
	ReorderGroup(context.Context, session.GroupOrder, int) ([]session.Group, error)
}
type groupOrderMutation struct {
	quitting bool
	observed uint64
}
type groupOrderResult struct {
	operation *groupOrderMutation
	groups    []session.Group
	err       error
}

func (m *Model) reorderSelectedGroup(delta int) tea.Cmd {
	if m.groupOrderFailed {
		m.status = "Group order unconfirmed; refresh before sorting again"
		return nil
	}
	service, ok := m.service.(groupOrderService)
	if !ok || m.groupOrder != nil || m.sessionOrder != nil || m.listErr != nil || !m.sidebarGroupSelected() || m.workspace.transition != workspaceIdle {
		return nil
	}
	target := m.selectedGroup()
	expected := session.GroupOrder{TargetID: target.ID}
	index := -1
	for _, g := range m.sidebarGroups() {
		if g.ParentID == target.ParentID {
			if g.ID == target.ID {
				index = len(expected.Siblings)
			}
			expected.Siblings = append(expected.Siblings, g)
		}
	}
	neighbor := index + delta
	if index < 0 || target.ID == session.DefaultGroupID || neighbor < 0 || neighbor >= len(expected.Siblings) || expected.Siblings[neighbor].ID == session.DefaultGroupID {
		return nil
	}
	operation := &groupOrderMutation{observed: m.groupsObserved}
	m.groupOrder = operation
	m.status = "Saving group order…"
	ctx := m.ctx
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		groups, err := service.ReorderGroup(ctx, expected, delta)
		return groupOrderResult{operation, groups, err}
	}
}

func (m *Model) groupOrderEvent(msg tea.Msg) (tea.Cmd, bool) {
	if result, ok := msg.(groupOrderResult); ok {
		if m.groupOrder != result.operation || m.groupOrder == nil {
			return nil, true
		}
		quitting := m.groupOrder.quitting
		m.groupOrder = nil
		m.listGeneration++
		if m.listCancel != nil {
			m.listCancel()
		}
		if result.err != nil {
			m.groupOrderFailed = true
			m.status = "Group order unconfirmed; refresh to verify: " + label(result.err.Error())
			return m.loadSessions(), true
		}
		m.acceptGroupOrder(result.groups, m.groupsObserved == result.operation.observed)
		if strings.HasPrefix(m.status, "Saving group order") {
			m.status = ""
		}
		if quitting {
			return m.workspaceKey(tea.KeyPressMsg{Code: 'q', Text: "q"}), true
		}
		return m.loadSessions(), true
	}
	if m.groupOrder != nil && m.groupOrder.quitting {
		switch msg.(type) {
		case tea.KeyPressMsg, tea.KeyReleaseMsg, tea.PasteMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
			return nil, true
		}
	}
	return nil, false
}

func (m *Model) acceptGroupOrder(incoming []session.Group, allowNew bool) {
	byID := map[string]session.Group{}
	for _, g := range m.sidebarGroups() {
		byID[g.ID] = g
	}
	for _, g := range incoming {
		old, ok := byID[g.ID]
		if !ok && !allowNew {
			continue
		}
		if !ok || old.Revision <= g.Revision {
			byID[g.ID] = g
		}
	}
	byParent := map[string][]session.Group{}
	for _, g := range byID {
		byParent[g.ParentID] = append(byParent[g.ParentID], g)
	}
	var ordered []session.Group
	var visit func(string, int)
	visit = func(parent string, depth int) {
		if depth >= session.MaxGroupDepth {
			return
		}
		children := byParent[parent]
		sort.Slice(children, func(i, j int) bool {
			if children[i].ID == session.DefaultGroupID || children[j].ID == session.DefaultGroupID {
				return children[i].ID == session.DefaultGroupID && children[j].ID != session.DefaultGroupID
			}
			if children[i].Order == children[j].Order {
				return children[i].ID < children[j].ID
			}
			return children[i].Order < children[j].Order
		})
		for _, g := range children {
			ordered = append(ordered, g)
			visit(g.ID, depth+1)
		}
	}
	visit("", 0)
	m.groups = ordered
	m.workspace.listClick = workspaceListClick{}
}
