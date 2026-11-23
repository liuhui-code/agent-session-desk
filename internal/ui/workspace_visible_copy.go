// Visible-pane normalization/interaction adapted from Agent Deck, Copyright
// (c) 2025 Ashesh Goplani. SPDX-License-Identifier: MIT
// Source: pane_copy.go, home.go and hotkeys.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: original Run revalidation,
// bounded fresh capture, view cancellation and the existing OSC52 writer.
package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type visiblePaneService interface {
	CaptureVisible(context.Context, session.Session) (session.Preview, error)
}
type visiblePaneCopy struct {
	generation uint64
	target     session.Session
	cancel     context.CancelFunc
	ready      bool
	payload    string
}
type visiblePaneResult struct {
	operation *visiblePaneCopy
	capture   session.Preview
	err       error
}

func normalizeVisiblePane(raw string) string {
	lines := strings.Split(plainText(raw), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}
func (m *Model) copyVisiblePane() tea.Cmd {
	w := m.workspace
	service, ok := m.service.(visiblePaneService)
	target, selected := m.selected()
	if !ok || !selected || w.transition != workspaceIdle {
		return nil
	}
	if w.visibleCopy != nil || m.clipboard.inflight != nil {
		m.status = "正在处理复制请求；请稍后再复制"
		return nil
	}
	if target.TerminalState != session.TerminalAvailable && target.TerminalState != session.TerminalExited {
		m.status = "目标终端状态未知；请刷新后复制"
		return nil
	}
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	op := &visiblePaneCopy{generation: w.generation, target: target, cancel: cancel}
	w.visibleCopy = op
	m.status = "正在捕获所选原终端的可见页…"
	return func() tea.Msg {
		defer cancel()
		p, e := service.CaptureVisible(ctx, target)
		return visiblePaneResult{op, p, e}
	}
}
func (m *Model) visiblePaneEvent(msg tea.Msg) (tea.Cmd, bool) {
	result, ok := msg.(visiblePaneResult)
	if !ok {
		return nil, false
	}
	w := m.workspace
	op := result.operation
	if op == nil || w.visibleCopy != op || op.generation != w.generation {
		return nil, true
	}
	err := result.err
	if !m.visibleCopyTargetCurrent(op.target) {
		err = fmt.Errorf("原 Session/Run 已变化；未发送复制")
	}
	if err == nil && (result.capture.Truncated || len(result.capture.Text) > terminal.MaxFrameBytes || !terminal.ValidCaptureEncoding(result.capture.Text)) {
		err = fmt.Errorf("可见页不完整、超限或编码无效；未发送复制")
	}
	if err != nil {
		w.visibleCopy = nil
		m.status = "可见页复制失败：" + label(err.Error())
		return m.drainWorkspaceClipboard(), true
	}
	payload := normalizeVisiblePane(result.capture.Text)
	if payload == "" {
		w.visibleCopy = nil
		m.status = "所选终端没有可复制的可见文字"
		return m.drainWorkspaceClipboard(), true
	}
	if len(payload) > terminal.MaxClipboardBytes {
		w.visibleCopy = nil
		m.status = "可见页超过复制上限；未发送部分文字"
		return m.drainWorkspaceClipboard(), true
	}
	op.ready, op.payload = true, payload
	if m.clipboard.inflight != nil {
		m.status = "可见页已捕获；等待前一次复制请求完成"
		return nil, true
	}
	return m.drainWorkspaceClipboard(), true
}

func (m *Model) visibleCopyTargetCurrent(target session.Session) bool {
	for _, v := range m.sessions {
		if sameStopTarget(v, target) {
			return true
		}
	}
	return false
}
