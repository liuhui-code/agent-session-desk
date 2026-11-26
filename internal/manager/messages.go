package manager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func (m *Manager) Read(ctx context.Context, target session.Session, lines int) (session.OutputRecord, error) {
	if lines < 1 || lines > 200 {
		return session.OutputRecord{}, fmt.Errorf("read requires 1-200 snapshot lines")
	}
	p, err := m.Terminal.Preview(ctx, target, lines)
	if err != nil {
		return session.OutputRecord{}, err
	}
	return m.metadata.RecordOutput(ctx, target, p)
}
func (m *Manager) Message(ctx context.Context, id string) (session.Message, error) {
	return m.metadata.Message(ctx, id)
}
func (m *Manager) PendingMessage(ctx context.Context, target session.Session) (session.Message, error) {
	return m.metadata.PendingMessage(ctx, target)
}

func (m *Manager) messageTarget(ctx context.Context, sid, rid string) (session.Session, error) {
	rows, err := m.List(ctx)
	if err != nil {
		return session.Session{}, err
	}
	for _, v := range rows {
		if v.ID == sid && v.RunID == rid && v.TerminalState == session.TerminalAvailable {
			return v, nil
		}
	}
	return session.Session{}, fmt.Errorf("exact target Session/Run is not available; no input written")
}
func (m *Manager) QueueMessage(ctx context.Context, r session.MessageRequest) (session.Message, error) {
	fp, err := r.Fingerprint()
	if err != nil {
		return session.Message{}, err
	}
	old, err := m.metadata.Message(ctx, r.RequestID)
	if err == nil {
		if old.Fingerprint != fp {
			return old, fmt.Errorf("request ID already used for different content or target")
		}
		return old, nil
	}
	if !errors.Is(err, store.ErrMessageNotFound) {
		return session.Message{}, err
	}
	target, err := m.messageTarget(ctx, r.TargetSessionID, r.TargetRunID)
	if err != nil {
		return session.Message{}, err
	}
	msg := session.Message{ID: r.RequestID, Target: target, Body: r.Body, Fingerprint: fp}
	if r.SourceRecordID != "" {
		records, e := m.metadata.Output(ctx, r.SourceRunID)
		if e != nil {
			return msg, e
		}
		found := false
		for _, record := range records {
			if record.ID != r.SourceRecordID {
				continue
			}
			found = true
			lines := strings.Split(record.Text, "\n")
			start, end := r.StartLine, r.EndLine
			if start == 0 {
				start = 1
			}
			if end == 0 {
				end = len(lines)
			}
			if start < 1 || end < start || end > len(lines) {
				return msg, fmt.Errorf("snapshot-local line range is outside the retained record")
			}
			msg.Body = strings.Join(lines[start-1:end], "\n")
			msg.Source = &session.OutputReference{SessionID: record.SessionID, RunID: record.RunID, RecordID: record.ID, RecordedAt: record.RecordedAt, StartLine: start, EndLine: end, Kind: record.Source, Truncated: record.Truncated, Complete: false}
			break
		}
		if !found {
			return msg, fmt.Errorf("source snapshot not retained; no message created")
		}
	}
	return m.metadata.QueueMessage(ctx, msg)
}

// DeliverMessage coordinates a durable attempt with an explicit backend transport.
// It does not depend on a TUI, focused view or confirmation dialog.
func (m *Manager) DeliverMessage(ctx context.Context, id string, target session.Session, paste func(context.Context, string) error) (session.Message, error) {
	if paste == nil {
		return session.Message{}, fmt.Errorf("no paste transport; backend delivery required")
	}
	live, err := m.messageTarget(ctx, target.ID, target.RunID)
	if err != nil {
		return session.Message{}, err
	}
	if live.BackendID != target.BackendID || live.Command != target.Command {
		return session.Message{}, fmt.Errorf("original target identity changed; no input written")
	}
	attempt, err := m.metadata.ClaimMessage(ctx, id, target)
	if err != nil {
		return session.Message{}, err
	}
	if err = ctx.Err(); err == nil {
		err = paste(ctx, attempt.Message.Body)
	}
	if err != nil {
		// The durable claim is already unknown, including cancellation and partial writes.
		if saved, e := m.metadata.FinishMessage(ctx, attempt, false); e == nil {
			attempt.Message = saved
		}
		return attempt.Message, fmt.Errorf("input result uncertain; check message %s; do not replay", id)
	}
	saved, err := m.metadata.FinishMessage(ctx, attempt, true)
	if err != nil {
		return attempt.Message, fmt.Errorf("input write completed but receipt unconfirmed for %s; do not replay", id)
	}
	return saved, nil
}

// DeliverQueued is the headless protocol operation used by agents and the CLI.
func (m *Manager) DeliverQueued(ctx context.Context, id string) (session.Message, error) {
	message, err := m.metadata.Message(ctx, id)
	if err != nil {
		return message, err
	}
	if message.State != "queued" {
		return message, fmt.Errorf("message is not queued; no replay attempted")
	}
	transport, ok := m.Terminal.(interface {
		Paste(context.Context, session.Session, string) error
	})
	if !ok {
		return message, fmt.Errorf("backend does not support headless delivery")
	}
	if capability, ok := m.Terminal.(interface{ SupportsMessageDelivery() bool }); ok && !capability.SupportsMessageDelivery() {
		return message, fmt.Errorf("headless delivery not available for this backend")
	}
	return m.DeliverMessage(ctx, id, message.Target, func(ctx context.Context, text string) error { return transport.Paste(ctx, message.Target, text) })
}
