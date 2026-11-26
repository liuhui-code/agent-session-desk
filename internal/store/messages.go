package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// Engineering retention bounds; no automatic pruning of idempotency receipts.
const MaxMessages = 128
const maxMessageDocument = 512 * 1024

var ErrMessageNotFound = errors.New("message not found in this scope")

const messageSchema = `CREATE TABLE messages(
 scope TEXT NOT NULL,id TEXT NOT NULL,target_run_id TEXT NOT NULL,fingerprint TEXT NOT NULL,
 document TEXT NOT NULL CHECK(length(CAST(document AS BLOB))<=524288),
 state TEXT NOT NULL CHECK(state IN ('queued','unknown','input_written')),
 reason TEXT NOT NULL,created_ns INTEGER NOT NULL,updated_ns INTEGER NOT NULL,attempt TEXT NOT NULL DEFAULT '',
 PRIMARY KEY(scope,id),FOREIGN KEY(scope,target_run_id) REFERENCES runs(scope,id));
 CREATE INDEX messages_pending ON messages(scope,target_run_id,state,created_ns,id);`
const messageColumns = `SELECT id,target_run_id,fingerprint,document,state,reason,created_ns,updated_ns FROM messages`

type messageRow interface{ Scan(...any) error }

func scanMessage(row messageRow) (session.Message, error) {
	var m session.Message
	var run, doc string
	var created, updated int64
	err := row.Scan(&m.ID, &run, &m.Fingerprint, &doc, &m.State, &m.Reason, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return m, ErrMessageNotFound
	}
	if err != nil {
		return m, err
	}
	id, fp, state, reason := m.ID, m.Fingerprint, m.State, m.Reason
	if len(doc) > maxMessageDocument {
		return m, fmt.Errorf("message document exceeds bound")
	}
	if err = json.Unmarshal([]byte(doc), &m); err != nil {
		return m, fmt.Errorf("invalid message document")
	}
	if m.ID != id || m.Target.RunID != run || session.ValidateMessageBody(m.Body) != nil {
		return m, fmt.Errorf("message identity/content conflict")
	}
	m.Fingerprint, m.State, m.Reason = fp, state, reason
	m.CreatedAt = time.Unix(0, created).UTC()
	m.UpdatedAt = time.Unix(0, updated).UTC()
	return m, nil
}
func (s *Store) Message(ctx context.Context, id string) (session.Message, error) {
	if !session.ValidID(id) {
		return session.Message{}, fmt.Errorf("full message ID required")
	}
	return scanMessage(s.db.QueryRowContext(ctx, messageColumns+` WHERE scope=? AND id=?`, s.scope, id))
}
func (s *Store) PendingMessage(ctx context.Context, target session.Session) (session.Message, error) {
	return scanMessage(s.db.QueryRowContext(ctx, messageColumns+` WHERE scope=? AND target_run_id=? AND state='queued' ORDER BY created_ns,id LIMIT 1`, s.scope, target.RunID))
}
func (s *Store) QueueMessage(ctx context.Context, m session.Message) (session.Message, error) {
	if !session.ValidID(m.ID) || len(m.Fingerprint) != 64 || !session.ValidID(m.Target.ID) || !session.ValidID(m.Target.RunID) {
		return m, fmt.Errorf("invalid message identities")
	}
	if err := session.ValidateMessageBody(m.Body); err != nil {
		return m, err
	}
	if m.Source != nil && (!session.ValidID(m.Source.SessionID) || !session.ValidID(m.Source.RunID) || !session.ValidID(m.Source.RecordID) || m.Source.StartLine < 1 || m.Source.EndLine < m.Source.StartLine || m.Source.Kind != "screen_snapshot" || m.Source.Complete || m.Source.RecordedAt.IsZero()) {
		return m, fmt.Errorf("invalid output provenance")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return m, err
	}
	defer tx.Rollback()
	old, err := scanMessage(tx.QueryRowContext(ctx, messageColumns+` WHERE scope=? AND id=?`, s.scope, m.ID))
	if err == nil {
		if old.Fingerprint != m.Fingerprint {
			return m, fmt.Errorf("request ID already used for different content or target")
		}
		return old, nil
	}
	if !errors.Is(err, ErrMessageNotFound) {
		return m, err
	}
	var n int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM messages`).Scan(&n); err != nil {
		return m, err
	}
	if n >= MaxMessages {
		return m, fmt.Errorf("message capacity reached; no receipts pruned")
	}
	var exact bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM runs r JOIN sessions s ON s.scope=r.scope AND s.id=r.session_id WHERE r.scope=? AND r.id=? AND r.session_id=? AND r.backend_id=? AND s.command=? AND r.stopped_ns IS NULL)`, s.scope, m.Target.RunID, m.Target.ID, m.Target.BackendID, m.Target.Command).Scan(&exact); err != nil {
		return m, err
	}
	if !exact {
		return m, fmt.Errorf("no exact live metadata Run for message target")
	}
	m.State, m.Reason = "queued", "explicit_delivery_required"
	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = m.CreatedAt
	data, err := json.Marshal(m)
	if err != nil {
		return m, err
	}
	if len(data) > maxMessageDocument {
		return m, fmt.Errorf("message document exceeds bound")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO messages(scope,id,target_run_id,fingerprint,document,state,reason,created_ns,updated_ns) VALUES(?,?,?,?,?,?,?,?,?)`, s.scope, m.ID, m.Target.RunID, m.Fingerprint, string(data), m.State, m.Reason, m.CreatedAt.UnixNano(), m.UpdatedAt.UnixNano())
	if err != nil {
		return m, err
	}
	if err = tx.Commit(); err != nil {
		return session.Message{}, err
	}
	return m, nil
}

// Claim commits uncertainty before any terminal side effect. A crash cannot
// turn this back into a safely replayable queued request.
func (s *Store) ClaimMessage(ctx context.Context, id string, target session.Session) (session.DeliveryAttempt, error) {
	var zero session.DeliveryAttempt
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()
	m, err := scanMessage(tx.QueryRowContext(ctx, messageColumns+` WHERE scope=? AND id=?`, s.scope, id))
	if err != nil {
		return zero, err
	}
	if m.State != "queued" || m.Target.ID != target.ID || m.Target.RunID != target.RunID || m.Target.BackendID != target.BackendID || m.Target.Command != target.Command {
		return zero, fmt.Errorf("message is not queued for this exact Run; no replay")
	}
	token, err := session.NewID()
	if err != nil {
		return zero, err
	}
	now := time.Now().UTC()
	result, err := tx.ExecContext(ctx, `UPDATE messages SET state='unknown',reason='delivery_started_result_unconfirmed',attempt=?,updated_ns=? WHERE scope=? AND id=? AND state='queued' AND EXISTS(SELECT 1 FROM runs WHERE scope=? AND id=? AND stopped_ns IS NULL) AND NOT EXISTS(SELECT 1 FROM task_claims c WHERE c.scope=messages.scope AND c.message_id=messages.id) AND NOT EXISTS(SELECT 1 FROM attention a WHERE a.scope=messages.scope AND a.run_id=messages.target_run_id AND a.state='pending')`, token, now.UnixNano(), s.scope, id, s.scope, target.RunID)
	if err != nil {
		return zero, err
	}
	n, err := result.RowsAffected()
	if err != nil || n != 1 {
		return zero, fmt.Errorf("message claim refused; target may have stopped")
	}
	if err = tx.Commit(); err != nil {
		return zero, err
	}
	m.State, m.Reason, m.UpdatedAt = "unknown", "delivery_started_result_unconfirmed", now
	return session.DeliveryAttempt{Message: m, Token: token}, nil
}
func (s *Store) FinishMessage(ctx context.Context, a session.DeliveryAttempt, written bool) (session.Message, error) {
	state, reason := "unknown", "transport_result_uncertain"
	if written {
		state, reason = "input_written", "agent_acceptance_unknown"
	}
	result, err := s.db.ExecContext(ctx, `UPDATE messages SET state=?,reason=?,attempt='',updated_ns=? WHERE scope=? AND id=? AND state='unknown' AND attempt=? AND attempt<>''`, state, reason, time.Now().UTC().UnixNano(), s.scope, a.Message.ID, a.Token)
	if err != nil {
		return session.Message{}, err
	}
	n, err := result.RowsAffected()
	if err != nil || n != 1 {
		return session.Message{}, fmt.Errorf("delivery receipt does not own the pending attempt")
	}
	return s.Message(ctx, a.Message.ID)
}
