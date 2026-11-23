// Confirmation interaction adapted from Agent Deck, Copyright (c) 2025 Ashesh
// Goplani. SPDX-License-Identifier: MIT
// Source: confirm_dialog.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: stop only, named original Run review, drain and one-shot receipts.
package ui

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type sessionStopService interface {
	StopReviewed(context.Context, session.Session) error
}
type sessionStopDialog struct {
	target          session.Session
	focus           int
	busy, submitted bool
	err             error
}
type sessionStopResult struct {
	dialog *sessionStopDialog
	err    error
}

func sameStopTarget(a, b session.Session) bool {
	return sameRun(a, b) && a.BackendID == b.BackendID && a.Command == b.Command
}

func (m *Model) beginSessionStop() tea.Cmd {
	target, ok := m.selected()
	if _, supported := m.service.(sessionStopService); !supported || !ok || m.listErr != nil || m.workspace.transition != workspaceIdle || m.groupOrder != nil || m.sessionOrder != nil {
		return nil
	}
	if target.TerminalState != session.TerminalAvailable && target.TerminalState != session.TerminalExited {
		m.status = "目标终端状态未知；请刷新核对后再停止"
		return nil
	}
	m.sessionStopIntent = &sessionStopDialog{target: target, focus: 1}
	return m.navigateWorkspace(workspaceStop)
}

func (m *Model) openSessionStop() tea.Cmd {
	m.sessionStop, m.sessionStopIntent = m.sessionStopIntent, nil
	m.workspace.focused = false
	if m.sessionStop == nil {
		return m.connectWorkspace()
	}
	return nil
}

func (m *Model) closeSessionStop() tea.Cmd {
	m.sessionStop = nil
	return tea.Batch(m.connectLiveWorkspace(), m.loadSessions())
}

func (m *Model) sessionStopEvent(msg tea.Msg) (tea.Cmd, bool) {
	if result, ok := msg.(sessionStopResult); ok {
		d := m.sessionStop
		if d == nil || d != result.dialog {
			return nil, true
		}
		d.busy = false
		m.listGeneration++
		if m.listCancel != nil {
			m.listCancel()
		}
		if result.err != nil {
			d.err = fmt.Errorf("停止结果未确认；请返回刷新核对，勿直接重复停止：%w", result.err)
			return m.loadSessions(), true
		}
		// Preserve the stopped view and all unrelated/replacement Runs. Never turn
		// a successful stop into an agent completion or a fabricated exit code.
		for i := range m.workspace.tabs {
			tab := &m.workspace.tabs[i]
			if sameStopTarget(tab.session, d.target) {
				tab.live, tab.session.TerminalState = false, session.TerminalExited
				if i == m.workspace.active {
					m.failWorkspace(fmt.Errorf("该 Run 已显式停止；保留视图与历史"))
				}
			}
		}
		rows := m.sessions[:0]
		for _, v := range m.sessions {
			if !sameStopTarget(v, d.target) {
				rows = append(rows, v)
			}
		}
		m.sessions = rows
		m.cursor = min(m.cursor, len(rows)-1)
		m.workspace.listClick = workspaceListClick{}
		m.sessionStop = nil
		m.status = "原 Run 已停止；Tab 视图和历史保留"
		return tea.Batch(m.connectLiveWorkspace(), m.loadSessions()), true
	}
	d := m.sessionStop
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
		switch msg.String() {
		case "esc", "n", "N", "ctrl+c":
			return m.closeSessionStop(), true
		case "y", "Y":
			return m.submitSessionStop(), true
		case "enter":
			if d.focus == 1 {
				return m.closeSessionStop(), true
			}
			return m.submitSessionStop(), true
		case "left", "h":
			d.focus = 0
		case "right", "l", "tab":
			d.focus = 1
		}
		return nil, true
	}
	return nil, false
}

func (m *Model) connectLiveWorkspace() tea.Cmd {
	w := m.workspace
	if len(w.tabs) == 0 || !w.tabs[w.active].live || w.tabs[w.active].session.TerminalState != session.TerminalAvailable {
		return nil
	}
	return m.connectWorkspace()
}

func (m *Model) submitSessionStop() tea.Cmd {
	d := m.sessionStop
	if d == nil || d.busy || d.submitted {
		return nil
	}
	if len(d.target.Title) > maxLabelBytes || len(d.target.Command) > maxLabelBytes {
		d.err = fmt.Errorf("名称或程序字段过长，无法完整核对；请返回处理后重新确认")
		return nil
	}
	if _, fits := m.sessionStopBox(); !fits {
		d.err = fmt.Errorf("请放大窗口核对完整停止范围")
		return nil
	}
	d.busy, d.submitted, d.err = true, true, nil
	service, ctx, target := m.service.(sessionStopService), m.ctx, d.target
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return sessionStopResult{d, service.StopReviewed(ctx, target)}
	}
}
