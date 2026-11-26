package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const MaxAgentBindings = 128

const replySchema = `CREATE TABLE agent_credentials(
 scope TEXT NOT NULL,run_id TEXT NOT NULL,token_hash TEXT NOT NULL CHECK(length(token_hash)=64),
 PRIMARY KEY(scope,run_id),UNIQUE(scope,token_hash),
 FOREIGN KEY(scope,run_id) REFERENCES runs(scope,id));
 CREATE TABLE replies(
 scope TEXT NOT NULL,id TEXT NOT NULL,reply_to TEXT NOT NULL,source_run_id TEXT NOT NULL,
 document TEXT NOT NULL CHECK(length(CAST(document AS BLOB))<=524288),created_ns INTEGER NOT NULL,
 PRIMARY KEY(scope,id),UNIQUE(scope,reply_to),
 FOREIGN KEY(scope,reply_to) REFERENCES messages(scope,id),
 FOREIGN KEY(scope,source_run_id) REFERENCES runs(scope,id));`

const activeAgent = `SELECT r.session_id,r.id,r.backend_id,s.title,s.command FROM runs r
 JOIN sessions s ON s.scope=r.scope AND s.id=r.session_id
 WHERE r.scope=? AND r.id=? AND r.session_id=? AND r.stopped_ns IS NULL
 AND NOT EXISTS(SELECT 1 FROM pane_exits e WHERE e.scope=r.scope AND e.run_id=r.id)`

func sameMessageTarget(a, b session.Session) bool {
	return a.ID == b.ID && a.RunID == b.RunID && a.BackendID == b.BackendID && a.Command == b.Command
}

func (s *Store) BindAgent(ctx context.Context, c session.AgentCredential, target session.Session) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if c.SessionID != target.ID || c.RunID != target.RunID {
		return fmt.Errorf("credential target mismatch")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var raw session.Session
	err = tx.QueryRowContext(ctx, activeAgent, s.scope, c.RunID, c.SessionID).Scan(&raw.ID, &raw.RunID, &raw.BackendID, &raw.Title, &raw.Command)
	if err != nil || !sameMessageTarget(raw, target) {
		return fmt.Errorf("original Run is not bindable")
	}
	var old string
	err = tx.QueryRowContext(ctx, `SELECT token_hash FROM agent_credentials WHERE scope=? AND run_id=?`, s.scope, c.RunID).Scan(&old)
	if err == nil {
		if old != c.Digest() {
			return fmt.Errorf("Run already bound; reuse original credential file")
		}
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var n int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM agent_credentials`).Scan(&n); err != nil {
		return err
	}
	if n >= MaxAgentBindings {
		return fmt.Errorf("agent binding capacity reached; no credentials pruned")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO agent_credentials VALUES(?,?,?)`, s.scope, c.RunID, c.Digest()); err != nil {
		return fmt.Errorf("agent binding could not be saved")
	}
	return tx.Commit()
}

func (s *Store) agentIdentity(ctx context.Context, q rowReader, c session.AgentCredential) (session.Session, error) {
	var v session.Session
	if err := c.Validate(); err != nil {
		return v, err
	}
	query := activeAgent + ` AND EXISTS(SELECT 1 FROM agent_credentials a WHERE a.scope=r.scope AND a.run_id=r.id AND a.token_hash=?)`
	if err := q.QueryRowContext(ctx, query, s.scope, c.RunID, c.SessionID, c.Digest()).Scan(&v.ID, &v.RunID, &v.BackendID, &v.Title, &v.Command); err != nil {
		return v, fmt.Errorf("agent credential or original Run unavailable")
	}
	v.TerminalState = session.TerminalUnknown
	return v, nil
}

func (s *Store) AgentIdentity(ctx context.Context, c session.AgentCredential) (session.Session, error) {
	return s.agentIdentity(ctx, s.db, c)
}

// Inbox is a bounded metadata query. Reading never delivers terminal input.
func (s *Store) Inbox(ctx context.Context, c session.AgentCredential) ([]session.Message, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	v, err := s.agentIdentity(ctx, tx, c)
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, messageColumns+` WHERE scope=? AND target_run_id=? AND state='queued' AND NOT EXISTS(SELECT 1 FROM task_claims c WHERE c.scope=messages.scope AND c.message_id=messages.id) AND NOT EXISTS(SELECT 1 FROM replies p WHERE p.scope=messages.scope AND p.reply_to=messages.id) ORDER BY created_ns,id LIMIT ?`, s.scope, v.RunID, MaxMessages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []session.Message{}
	for rows.Next() {
		m, e := scanMessage(rows)
		if e != nil {
			return nil, e
		}
		if !sameMessageTarget(m.Target, v) {
			return nil, fmt.Errorf("inbox target identity conflict")
		}
		result = append(result, m)
	}
	return result, rows.Err()
}
