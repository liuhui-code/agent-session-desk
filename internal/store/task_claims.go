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

const taskClaimSchema = `CREATE TABLE task_claims(
 scope TEXT NOT NULL,id TEXT NOT NULL,message_id TEXT NOT NULL,owner_run_id TEXT NOT NULL,
 token_hash TEXT NOT NULL CHECK(length(token_hash)=64),document TEXT NOT NULL CHECK(length(CAST(document AS BLOB))<=524288),
 state TEXT NOT NULL CHECK(state IN ('active','unknown','replied')),reason TEXT NOT NULL,
 lease_seconds INTEGER NOT NULL CHECK(lease_seconds BETWEEN 1 AND 3600),expires_ns INTEGER NOT NULL,
 created_ns INTEGER NOT NULL,updated_ns INTEGER NOT NULL,
 PRIMARY KEY(scope,id),UNIQUE(scope,message_id),
 FOREIGN KEY(scope,message_id) REFERENCES messages(scope,id),
 FOREIGN KEY(scope,owner_run_id) REFERENCES runs(scope,id));`

const taskClaimColumns = `SELECT id,message_id,owner_run_id,document,state,reason,lease_seconds,expires_ns,created_ns,updated_ns FROM task_claims`

func scanTaskClaim(row messageRow) (*session.TaskClaim, error) {
	var c session.TaskClaim
	var id, msg, run, doc, state, reason string
	var seconds int
	var expires, created, updated int64
	if err := row.Scan(&id, &msg, &run, &doc, &state, &reason, &seconds, &expires, &created, &updated); err != nil {
		return nil, err
	}
	if len(doc) > maxMessageDocument || json.Unmarshal([]byte(doc), &c) != nil || c.ID != id || c.MessageID != msg || c.Owner.RunID != run || !session.ValidID(id) || !session.ValidID(msg) || seconds < 1 || seconds > session.MaxClaimSeconds {
		return nil, fmt.Errorf("invalid task claim record")
	}
	c.State, c.Reason, c.LeaseSeconds = state, reason, seconds
	c.NewlyClaimed = false
	c.ExpiresAt, c.CreatedAt, c.UpdatedAt = time.Unix(0, expires).UTC(), time.Unix(0, created).UTC(), time.Unix(0, updated).UTC()
	return &c, nil
}

// Materialization never releases ownership or turns uncertainty into runnable work.
func (s *Store) refreshTaskClaim(ctx context.Context, tx *sql.Tx, msg string, now time.Time) (*session.TaskClaim, error) {
	_, err := tx.ExecContext(ctx, `UPDATE task_claims SET state='unknown',reason=CASE WHEN expires_ns<=? THEN 'lease_expired' ELSE 'owner_exit_observed' END,updated_ns=? WHERE scope=? AND message_id=? AND state='active' AND (expires_ns<=? OR EXISTS(SELECT 1 FROM runs r WHERE r.scope=task_claims.scope AND r.id=owner_run_id AND r.stopped_ns IS NOT NULL) OR EXISTS(SELECT 1 FROM pane_exits e WHERE e.scope=task_claims.scope AND e.run_id=owner_run_id))`, now.UnixNano(), now.UnixNano(), s.scope, msg, now.UnixNano())
	if err != nil {
		return nil, err
	}
	c, err := scanTaskClaim(tx.QueryRowContext(ctx, taskClaimColumns+` WHERE scope=? AND message_id=?`, s.scope, msg))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (s *Store) ownedTaskClaim(ctx context.Context, tx *sql.Tx, source session.Session, key session.TaskClaimCredential, now time.Time) (*session.TaskClaim, error) {
	if err := key.Validate(); err != nil {
		return nil, err
	}
	c, err := s.refreshTaskClaim(ctx, tx, key.MessageID, now)
	if err != nil {
		return nil, err
	}
	if c == nil || c.ID != key.ID || !sameMessageTarget(c.Owner, source) {
		return nil, fmt.Errorf("task claim is not owned by this caller")
	}
	var digest string
	if err = tx.QueryRowContext(ctx, `SELECT token_hash FROM task_claims WHERE scope=? AND id=?`, s.scope, key.ID).Scan(&digest); err != nil {
		return nil, err
	}
	if digest != key.Digest() {
		return nil, fmt.Errorf("task claim credential refused")
	}
	return c, nil
}

func (s *Store) TaskClaimStatus(ctx context.Context, id string) (*session.TaskClaim, error) {
	if _, err := s.Message(ctx, id); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	c, err := s.refreshTaskClaim(ctx, tx, id, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return c, tx.Commit()
}

func (s *Store) ClaimTask(ctx context.Context, agent session.AgentCredential, key session.TaskClaimCredential, seconds int) (*session.TaskClaim, error) {
	if err := key.Validate(); err != nil {
		return nil, err
	}
	if seconds < 1 || seconds > session.MaxClaimSeconds {
		return nil, fmt.Errorf("lease-seconds must be 1-3600")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	source, err := s.agentIdentity(ctx, tx, agent)
	if err != nil {
		return nil, err
	}
	msg, err := scanMessage(tx.QueryRowContext(ctx, messageColumns+` WHERE scope=? AND id=?`, s.scope, key.MessageID))
	if err != nil {
		return nil, err
	}
	if !sameMessageTarget(msg.Target, source) {
		return nil, fmt.Errorf("claim source does not match original target")
	}
	now := time.Now().UTC()
	old, err := s.refreshTaskClaim(ctx, tx, key.MessageID, now)
	if err != nil {
		return nil, err
	}
	if old != nil {
		owned, e := s.ownedTaskClaim(ctx, tx, source, key, now)
		if e != nil || old.LeaseSeconds != seconds {
			return nil, fmt.Errorf("request already claimed; no replacement allowed")
		}
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		if owned.State == "unknown" {
			return owned, fmt.Errorf("claim result uncertain; no processing or renewal allowed")
		}
		return owned, nil
	}
	var humanPending bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM attention WHERE scope=? AND run_id=? AND state='pending')`, s.scope, source.RunID).Scan(&humanPending); err != nil {
		return nil, err
	}
	if humanPending {
		return nil, fmt.Errorf("original Run has a pending human decision; new work paused")
	}
	if msg.State != "queued" {
		return nil, fmt.Errorf("terminal delivery already attempted; task ownership refused")
	}
	var replied bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM replies WHERE scope=? AND reply_to=?)`, s.scope, msg.ID).Scan(&replied); err != nil {
		return nil, err
	}
	if replied {
		return nil, fmt.Errorf("request already has a result")
	}
	c := session.TaskClaim{ID: key.ID, MessageID: key.MessageID, Owner: source, State: "active", Reason: "explicit_processing_claim", LeaseSeconds: seconds, ExpiresAt: now.Add(time.Duration(seconds) * time.Second), CreatedAt: now, UpdatedAt: now}
	data, err := json.Marshal(c)
	if err != nil || len(data) > maxMessageDocument {
		return nil, fmt.Errorf("task claim exceeds bound")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO task_claims VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, s.scope, c.ID, c.MessageID, c.Owner.RunID, key.Digest(), string(data), c.State, c.Reason, seconds, c.ExpiresAt.UnixNano(), now.UnixNano(), now.UnixNano())
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("claim commit unconfirmed; preserve claim file and query claim-status")
	}
	if !c.ExpiresAt.After(time.Now().UTC()) {
		status, e := s.TaskClaimStatus(ctx, c.MessageID)
		if e != nil {
			return nil, e
		}
		return status, fmt.Errorf("claim lease expired before receipt; no processing granted")
	}
	c.NewlyClaimed = true
	return &c, nil
}

func (s *Store) RenewTaskClaim(ctx context.Context, agent session.AgentCredential, key session.TaskClaimCredential) (*session.TaskClaim, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	source, err := s.agentIdentity(ctx, tx, agent)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	c, err := s.ownedTaskClaim(ctx, tx, source, key, now)
	if err != nil {
		return nil, err
	}
	if c.State != "active" {
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return c, fmt.Errorf("claim is not active; no renewal or reassignment")
	}
	now = time.Now().UTC()
	if !c.ExpiresAt.After(now) {
		expired, e := s.refreshTaskClaim(ctx, tx, key.MessageID, now)
		if e != nil {
			return nil, e
		}
		if e = tx.Commit(); e != nil {
			return nil, e
		}
		return expired, fmt.Errorf("claim lease expired; no renewal allowed")
	}
	expires := now.Add(time.Duration(c.LeaseSeconds) * time.Second)
	if expires.Before(c.ExpiresAt) {
		expires = c.ExpiresAt
	}
	_, err = tx.ExecContext(ctx, `UPDATE task_claims SET expires_ns=?,updated_ns=? WHERE scope=? AND id=? AND state='active'`, expires.UnixNano(), now.UnixNano(), s.scope, c.ID)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("renewal unconfirmed; query claim-status")
	}
	c.ExpiresAt, c.UpdatedAt = expires, now
	return c, nil
}
