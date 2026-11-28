package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (s *Store) RecoveryAttempt(ctx context.Context, id string) (recovery.Attempt, error) {
	return s.recoveryAttempt(ctx, s.db, id)
}
func (s *Store) recoveryAttempt(ctx context.Context, q rowReader, id string) (recovery.Attempt, error) {
	var a recovery.Attempt
	var result sql.NullString
	err := q.QueryRowContext(ctx, `SELECT id,session_id,run_id,conversation_id,revision,state,result FROM recovery_attempts WHERE scope=? AND session_id=?`, s.scope, id).Scan(&a.ID, &a.SessionID, &a.RunID, &a.ConversationID, &a.Revision, &a.State, &result)
	if err != nil {
		return a, fmt.Errorf("recovery attempt unavailable")
	}
	if result.Valid {
		r, e := recovery.DecodeResult([]byte(result.String))
		if e != nil {
			return a, e
		}
		a.Result = &r
	}
	return a, nil
}

// ReserveRecovery commits identity, credential and the no-retry fence before
// any external launch. A crash leaves launch_unknown, never permission to replay.
func (s *Store) ReserveRecovery(ctx context.Context, r recovery.Reservation) (recovery.Attempt, error) {
	a := r.Attempt
	var empty recovery.Attempt
	if !session.ValidID(a.ID) || !session.ValidID(a.SessionID) || !session.ValidID(a.RunID) || !session.ValidID(r.PreviousRunID) || a.RunID == r.PreviousRunID || r.Target.ID != a.SessionID || r.Target.RunID != a.RunID || r.Target.BackendID != "asd_"+a.SessionID+"_"+a.RunID || r.Credential.SessionID != a.SessionID || r.Credential.RunID != a.RunID {
		return empty, fmt.Errorf("invalid recovery reservation")
	}
	if err := r.Credential.Validate(); err != nil {
		return empty, err
	}
	b, err := launchDocument(r.Launch)
	if err != nil {
		return empty, err
	}
	if r.Target.Command != r.Launch.Executable || r.Target.Title != r.Launch.Name {
		return empty, fmt.Errorf("recovery launch snapshot mismatch")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback()
	p, err := s.recoveryProfile(ctx, tx, a.SessionID)
	if err != nil {
		return empty, err
	}
	if p.Revision != a.Revision || p.Declaration.ConversationID != a.ConversationID {
		return empty, fmt.Errorf("recovery profile changed")
	}
	previous, err := s.latestRecoveryRun(ctx, tx, a.SessionID)
	if err != nil || previous.RunID != r.PreviousRunID {
		return empty, fmt.Errorf("latest Run changed; recovery not started")
	}
	var pending int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM recovery_attempts a JOIN runs r ON r.scope=a.scope AND r.id=a.run_id WHERE a.scope=? AND a.session_id=? AND a.state='launch_unknown' AND r.stopped_ns IS NULL AND NOT EXISTS(SELECT 1 FROM pane_exits e WHERE e.scope=r.scope AND e.run_id=r.id)`, s.scope, a.SessionID).Scan(&pending); err != nil {
		return empty, err
	}
	if pending != 0 {
		return empty, fmt.Errorf("previous recovery launch remains unknown; no new attempt allowed")
	}
	var count int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM agent_credentials`).Scan(&count); err != nil {
		return empty, err
	}
	if count >= MaxAgentBindings {
		return empty, fmt.Errorf("agent binding capacity reached")
	}
	var previousTime int64
	if err = tx.QueryRowContext(ctx, `SELECT recorded_ns FROM runs WHERE scope=? AND id=?`, s.scope, previous.RunID).Scan(&previousTime); err != nil {
		return empty, err
	}
	now := time.Now().UnixNano()
	if now <= previousTime {
		now = previousTime + 1
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO runs(scope,id,session_id,backend_id,recorded_ns,launch_title,launch_command) VALUES(?,?,?,?,?,?,?)`, s.scope, a.RunID, a.SessionID, r.Target.BackendID, now, r.Target.Title, r.Target.Command); err != nil {
		return empty, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO run_launches VALUES(?,?,?,?) ON CONFLICT(scope,session_id) DO UPDATE SET run_id=excluded.run_id,document=excluded.document`, s.scope, a.SessionID, a.RunID, b); err != nil {
		return empty, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO agent_credentials VALUES(?,?,?)`, s.scope, a.RunID, r.Credential.Digest()); err != nil {
		return empty, fmt.Errorf("recovery credential could not be saved")
	}
	a.State = "launch_unknown"
	if _, err = tx.ExecContext(ctx, `INSERT INTO recovery_attempts VALUES(?,?,?,?,?,?,?,NULL) ON CONFLICT(scope,session_id) DO UPDATE SET id=excluded.id,run_id=excluded.run_id,conversation_id=excluded.conversation_id,revision=excluded.revision,state=excluded.state,result=NULL`, s.scope, a.SessionID, a.ID, a.RunID, a.ConversationID, a.Revision, a.State); err != nil {
		return empty, err
	}
	return a, tx.Commit()
}

func (s *Store) RecoveryStarted(ctx context.Context, v session.Session) error {
	result, err := s.db.ExecContext(ctx, `UPDATE recovery_attempts SET state=CASE WHEN state='launch_unknown' THEN 'started' ELSE state END WHERE scope=? AND session_id=? AND run_id=? AND EXISTS(SELECT 1 FROM runs r WHERE r.scope=? AND r.id=? AND r.backend_id=? AND r.launch_command=?)`, s.scope, v.ID, v.RunID, s.scope, v.RunID, v.BackendID, v.Command)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("recovery start receipt not accepted")
	}
	return nil
}

func (s *Store) RecordRecoveryResult(ctx context.Context, c session.AgentCredential, r recovery.Result) (recovery.Attempt, error) {
	var empty recovery.Attempt
	if err := r.Validate(); err != nil {
		return empty, err
	}
	if c.SessionID != r.SessionID || c.RunID != r.RunID {
		return empty, fmt.Errorf("recovery result credential mismatch")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback()
	if _, err = s.agentIdentity(ctx, tx, c); err != nil {
		return empty, err
	}
	var aid, rid, cid, state string
	var revision int64
	var old sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT id,run_id,conversation_id,revision,state,result FROM recovery_attempts WHERE scope=? AND session_id=?`, s.scope, r.SessionID).Scan(&aid, &rid, &cid, &revision, &state, &old)
	if err != nil || aid != r.AttemptID || rid != r.RunID || cid != r.ConversationID {
		return empty, fmt.Errorf("recovery result attempt mismatch")
	}
	b, _ := json.Marshal(r)
	if old.Valid && old.String != string(b) {
		return empty, fmt.Errorf("conflicting recovery final result")
	}
	if !old.Valid {
		if state != "started" {
			return empty, fmt.Errorf("launch not confirmed; preserve result and resubmit after start receipt")
		}
		if _, err = tx.ExecContext(ctx, `UPDATE recovery_attempts SET state=?,result=? WHERE scope=? AND session_id=?`, r.Status, string(b), s.scope, r.SessionID); err != nil {
			return empty, err
		}
	}
	if err = tx.Commit(); err != nil {
		return empty, err
	}
	// Return the exact transaction receipt. A subsequent explicit recovery may
	// replace the current row immediately after commit; never return its IDs.
	return recovery.Attempt{ID: aid, SessionID: r.SessionID, RunID: rid, ConversationID: cid, Revision: revision, State: r.Status, Result: &r}, nil
}

func launchDocument(l session.Launch) (string, error) {
	if err := l.Validate(); err != nil {
		return "", err
	}
	b, err := json.Marshal(l)
	if err != nil || len(b) > 65536 {
		return "", fmt.Errorf("launch description exceeds 64 KiB")
	}
	return string(b), nil
}

// SaveLaunch retains one complete explicit launch per logical Session. Legacy
// discovery alone cannot reconstruct argv/cwd/env and does not invent them.
func (s *Store) SaveLaunch(ctx context.Context, v session.Session, l session.Launch) error {
	b, err := launchDocument(l)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO run_launches SELECT ?,?,?,? WHERE EXISTS(SELECT 1 FROM runs WHERE scope=? AND id=? AND session_id=? AND launch_command=?) ON CONFLICT(scope,session_id) DO UPDATE SET run_id=excluded.run_id,document=excluded.document`, s.scope, v.ID, v.RunID, b, s.scope, v.RunID, v.ID, v.Command)
	return err
}
