package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type messageDeliveryService interface {
	PendingMessage(context.Context, session.Session) (session.Message, error)
	DeliverMessage(context.Context, string, session.Session, func(context.Context, string) error) (session.Message, error)
}
type deliveryDialog struct {
	target                   session.Session
	message                  session.Message
	focus, offset            int
	loading, busy, submitted bool
	err                      error
}
type deliveryLoaded struct {
	dialog  *deliveryDialog
	message session.Message
	err     error
}
type deliveryWorkResult struct {
	message session.Message
	err     error
}
type deliveryDone struct {
	dialog  *deliveryDialog
	receipt tea.Msg
	result  deliveryWorkResult
}

func (m *Model) beginMessageDelivery() tea.Cmd {
	w := m.workspace
	service, ok := m.service.(messageDeliveryService)
	if !ok || m.sidebarGroupSelected() {
		return nil
	}
	if w.focused || w.busy || w.err != nil || w.conn == nil || w.transition != workspaceIdle || w.pending > 0 || w.clipboardReading || w.programPointer != nil || w.copyPointer != nil || m.groupOrder != nil || m.sessionOrder != nil {
		m.status = "Finish the current input or operation before reviewing a message"
		return nil
	}
	target, ok := m.selected()
	if !ok || len(w.tabs) == 0 || !sameRun(target, w.tabs[w.active].session) {
		m.status = "Open the target session first, then press m in navigation"
		return nil
	}
	if w.screen.CopyMode {
		m.status = "Exit terminal copy mode before reviewing a message"
		return nil
	}
	d := &deliveryDialog{target: target, focus: 1, loading: true}
	w.delivery = d
	m.status = ""
	return func() tea.Msg {
		message, err := service.PendingMessage(m.ctx, target)
		return deliveryLoaded{d, message, err}
	}
}
func (m *Model) closeMessageDelivery() tea.Cmd { m.workspace.delivery = nil; return m.workspaceSize() }

func (m *Model) deliveryEvent(msg tea.Msg) (tea.Cmd, bool) {
	w := m.workspace
	switch v := msg.(type) {
	case deliveryLoaded:
		if w.delivery == v.dialog {
			d := w.delivery
			d.loading = false
			d.message, d.err = v.message, v.err
		}
		return nil, true
	case deliveryDone:
		cmd, _ := m.workspaceEvent(v.receipt)
		if w.delivery == v.dialog {
			d := w.delivery
			d.busy = false
			d.submitted = true
			if v.result.message.ID != "" {
				d.message = v.result.message
			}
			d.err = v.result.err
		}
		return cmd, true
	}
	d := w.delivery
	if d == nil {
		return nil, false
	}
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, v.Width), max(0, v.Height)
		return m.workspaceSize(), true
	case tea.KeyReleaseMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.PasteMsg:
		return nil, true
	case tea.KeyPressMsg:
		if d.busy {
			return nil, true
		}
		if v.String() == "esc" || v.String() == "ctrl+c" {
			return m.closeMessageDelivery(), true
		}
		if d.loading || d.submitted || d.err != nil {
			return nil, true
		}
		switch v.String() {
		case "tab", "shift+tab", "left", "right":
			d.focus = 1 - d.focus
		case "up", "k":
			d.offset = max(0, d.offset-1)
		case "down", "j":
			d.offset++
		case "pgup":
			d.offset = max(0, d.offset-m.tokens.workspace.deliveryPreviewRows)
		case "pgdown":
			d.offset += m.tokens.workspace.deliveryPreviewRows
		case "home":
			d.offset = 0
		case "end":
			d.offset = len(d.message.Body)
		case "enter":
			if d.focus == 1 {
				return m.closeMessageDelivery(), true
			}
			return m.submitMessageDelivery(), true
		}
		d.offset = min(d.offset, max(0, len(m.messageDeliveryLines())-m.tokens.workspace.deliveryPreviewRows))
		return nil, true
	}
	return nil, false
}
func (m *Model) submitMessageDelivery() tea.Cmd {
	w, d := m.workspace, m.workspace.delivery
	_, fits := m.messageDeliveryBox()
	if !fits {
		d.err = fmt.Errorf("Enlarge the window and reopen to review the target before pasting")
		return nil
	}
	if w.busy || w.err != nil || w.conn == nil || w.transition != workspaceIdle || w.pending > 0 || len(w.tabs) == 0 || !sameRun(w.tabs[w.active].session, d.target) {
		d.err = fmt.Errorf("Original terminal is no longer ready; no input written")
		return nil
	}
	service := m.service.(messageDeliveryService)
	results := make(chan deliveryWorkResult, 1)
	id, target := d.message.ID, d.target
	action := workspaceInput{kind: "delivery", deliver: func(ctx context.Context, conn terminal.Connection) error {
		var inputErr error
		message, err := service.DeliverMessage(ctx, id, target, func(ctx context.Context, text string) error { inputErr = conn.Paste(ctx, text); return inputErr })
		results <- deliveryWorkResult{message, err}
		return inputErr
	}}
	d.busy = true
	cmd := m.workspaceWrite(action)
	if cmd == nil {
		d.busy = false
		d.err = fmt.Errorf("Paste did not start; check message status before continuing")
		return nil
	}
	return func() tea.Msg {
		receipt := cmd()
		var result deliveryWorkResult
		select {
		case result = <-results:
		default:
			result = deliveryWorkResult{message: session.Message{ID: id, Target: target, State: "unknown"}, err: fmt.Errorf("Input outcome unconfirmed; check asd status %s before any retry", id)}
		}
		return deliveryDone{d, receipt, result}
	}
}
