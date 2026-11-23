// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/split_config.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: token-based bounds, management-only keys, one notice timer and
// in-memory view state; no implicit TOML writes. See THIRD_PARTY_NOTICES.md.
package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"time"
)

type workspaceSplit struct {
	dragging, moved, hovered bool
	dragStart                int
	override                 bool
	percent                  int
	noticeUntil              time.Time
	noticePending            bool
}
type workspaceSplitTick struct{}

func (m *Model) workspaceSidebarPercent() int {
	if m.workspace.split.override {
		return m.workspace.split.percent
	}
	if m.workspace.focused {
		return m.tokens.workspace.compactPercent
	}
	return m.tokens.workspace.sidebarPercent
}
func (m *Model) workspaceSplitKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	if msg.Mod&(tea.ModCtrl|tea.ModAlt) != 0 {
		return nil, false
	}
	key := msg.String()
	if msg.Text != "" {
		key = msg.Text
	}
	switch key {
	case "<":
		return m.setWorkspaceSplit(m.workspaceSidebarPercent() + m.tokens.workspace.splitStep), true
	case ">":
		return m.setWorkspaceSplit(m.workspaceSidebarPercent() - m.tokens.workspace.splitStep), true
	case "=":
		old := m.workspaceSidebarPercent()
		m.workspace.split.override = false
		m.status = ""
		var resize tea.Cmd
		if old != m.workspaceSidebarPercent() {
			resize = m.workspaceSize()
		}
		return tea.Batch(resize, m.markWorkspaceSplitNotice()), true
	}
	return nil, false
}
func (m *Model) setWorkspaceSplit(percent int) tea.Cmd {
	old := m.workspaceSidebarPercent()
	t := m.tokens.workspace
	percent = max(t.minSplitPercent, min(t.maxSplitPercent, percent))
	m.workspace.split.override = true
	m.workspace.split.percent = percent
	m.status = ""
	var resize tea.Cmd
	if old != percent {
		resize = m.workspaceSize()
	}
	return tea.Batch(resize, m.markWorkspaceSplitNotice())
}
func (m *Model) markWorkspaceSplitNotice() tea.Cmd {
	s := &m.workspace.split
	s.noticeUntil = time.Now().Add(m.tokens.workspace.splitNoticeDuration)
	if s.noticePending {
		return nil
	}
	return m.waitWorkspaceSplitNotice()
}
func (m *Model) waitWorkspaceSplitNotice() tea.Cmd {
	s := &m.workspace.split
	s.noticePending = true
	return tea.Tick(max(0, time.Until(s.noticeUntil)), func(time.Time) tea.Msg { return workspaceSplitTick{} })
}
func (m *Model) workspaceSplitExpired() tea.Cmd {
	s := &m.workspace.split
	s.noticePending = false
	if time.Now().Before(s.noticeUntil) {
		return m.waitWorkspaceSplitNotice()
	}
	s.noticeUntil = time.Time{}
	return nil
}

func (m *Model) workspaceDividerHit(x, y int) bool {
	g := m.workspaceGeometry()
	top := g.y - m.styles.terminalFrame.GetBorderTopSize()
	return g.left > 0 && y >= top && y < top+g.bodyHeight && x >= g.left-m.tokens.workspace.dividerGrabSlop && x < g.left+lipgloss.Width(m.tokens.divider())+m.tokens.workspace.dividerGrabSlop
}
func (m *Model) workspaceSplitMouseX(x int) tea.Cmd {
	if m.width <= 0 {
		return nil
	}
	x = max(0, min(m.width, x))
	return m.setWorkspaceSplit(x * 100 / m.width)
}
func (m *Model) workspaceSplitMotion(msg tea.MouseMotionMsg) tea.Cmd {
	s := &m.workspace.split
	s.hovered = msg.Button == tea.MouseNone && m.workspaceDividerHit(msg.X, msg.Y)
	if !s.dragging {
		return nil
	}
	if msg.Button != tea.MouseLeft {
		s.dragging = false
		return nil
	}
	if msg.X != s.dragStart {
		s.moved = true
	}
	if !s.moved {
		return nil
	}
	return m.workspaceSplitMouseX(msg.X)
}
func (m *Model) workspaceSplitRelease(msg tea.MouseReleaseMsg) tea.Cmd {
	s := &m.workspace.split
	s.hovered = m.workspaceDividerHit(msg.X, msg.Y)
	if !s.dragging {
		return nil
	}
	s.dragging = false
	if !s.moved {
		return nil
	}
	return m.workspaceSplitMouseX(msg.X)
}
