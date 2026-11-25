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
		m.status = "Copy request in progress; wait before copying again"
		return nil
	}
	if target.TerminalState != session.TerminalAvailable && target.TerminalState != session.TerminalExited {
		m.status = "Terminal state unknown; refresh before copying"
		return nil
	}
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	op := &visiblePaneCopy{generation: w.generation, target: target, cancel: cancel}
	w.visibleCopy = op
	m.status = "Capturing the selected terminal screen…"
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
		err = fmt.Errorf("Session/Run changed; copy was not sent")
	}
	if err == nil && (result.capture.Truncated || len(result.capture.Text) > terminal.MaxFrameBytes || !terminal.ValidCaptureEncoding(result.capture.Text)) {
		err = fmt.Errorf("Screen incomplete, too large or invalidly encoded; copy not sent")
	}
	if err != nil {
		w.visibleCopy = nil
		m.status = "Screen copy failed: " + label(err.Error())
		return m.drainWorkspaceClipboard(), true
	}
	payload := normalizeVisiblePane(result.capture.Text)
	if payload == "" {
		w.visibleCopy = nil
		m.status = "No visible text to copy from the selected terminal"
		return m.drainWorkspaceClipboard(), true
	}
	if len(payload) > terminal.MaxClipboardBytes {
		w.visibleCopy = nil
		m.status = "Screen exceeds copy limit; no partial text was sent"
		return m.drainWorkspaceClipboard(), true
	}
	op.ready, op.payload = true, payload
	if m.clipboard.inflight != nil {
		m.status = "Screen captured; waiting for the previous copy request"
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
