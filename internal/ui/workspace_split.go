// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Sources: internal/ui/split_config.go and home_keys_main.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: token-based bounds, management-only keys, one notice timer and
// explicit serialized UI preferences and transient drag preview under ADR-0027.
// See THIRD_PARTY_NOTICES.md.
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
	preview                  int
	noticeUntil              time.Time
	noticePending            bool
}
type workspaceSplitTick struct{}

func (m *Model) workspaceSidebarPercent() int {
	if m.workspace.split.dragging && m.workspace.split.moved {
		return m.workspace.split.preview
	}
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
		m.workspace.split.percent = 0
		m.status = ""
		var resize tea.Cmd
		if old != m.workspaceSidebarPercent() {
			resize = m.workspaceSize()
		}
		return tea.Batch(resize, m.markWorkspaceSplitNotice(), m.saveWorkspaceSplit(0)), true
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
	return tea.Batch(resize, m.markWorkspaceSplitNotice(), m.saveWorkspaceSplit(percent))
}

func (m *Model) saveWorkspaceSplit(percent int) tea.Cmd {
	if m.preferences.wanted.SidebarPercent == percent {
		return nil
	}
	m.preferences.wanted.SidebarPercent = percent
	return m.saveUIPreferences()
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
	top := g.bodyY
	dividerX := g.bodyX + g.left
	return g.left > 0 && y >= top && y < top+g.bodyHeight && x >= dividerX-m.tokens.workspace.dividerGrabSlop && x < dividerX+lipgloss.Width(m.tokens.divider())+m.tokens.workspace.dividerGrabSlop
}
func (m *Model) workspaceSplitMouseX(x int) tea.Cmd {
	g := m.workspaceGeometry()
	if g.bodyWidth <= 0 {
		return nil
	}
	x = max(0, min(g.bodyWidth, x-g.bodyX))
	old := m.workspaceSidebarPercent()
	t := m.tokens.workspace
	m.workspace.split.preview = max(t.minSplitPercent, min(t.maxSplitPercent, x*100/g.bodyWidth))
	m.status = ""
	var resize tea.Cmd
	if old != m.workspaceSidebarPercent() {
		resize = m.workspaceSize()
	}
	return tea.Batch(resize, m.markWorkspaceSplitNotice())
}
func (m *Model) workspaceSplitMotion(msg tea.MouseMotionMsg) tea.Cmd {
	s := &m.workspace.split
	s.hovered = msg.Button == tea.MouseNone && m.workspaceDividerHit(msg.X, msg.Y)
	if !s.dragging {
		return nil
	}
	if msg.Button != tea.MouseLeft {
		return m.finishWorkspaceSplit()
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
	// X10 releases may have no button and no reliable pointer coordinate.
	// Like the source, save the last motion rather than recomputing from release.
	return m.finishWorkspaceSplit()
}

func (m *Model) finishWorkspaceSplit() tea.Cmd {
	s := &m.workspace.split
	if !s.dragging {
		return nil
	}
	s.dragging = false
	if s.moved {
		return m.setWorkspaceSplit(s.preview)
	}
	return nil
}

func (m *Model) restoreWorkspaceSplit(percent int) tea.Cmd {
	s := &m.workspace.split
	old := m.workspaceSidebarPercent()
	s.dragging, s.override, s.percent = false, percent != 0, percent
	if old != m.workspaceSidebarPercent() {
		return m.workspaceSize()
	}
	return nil
}
