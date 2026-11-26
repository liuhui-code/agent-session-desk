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

const replyColumns = `SELECT id,reply_to,source_run_id,document,created_ns FROM replies`

func scanReply(row messageRow) (session.Reply, error) {
	var r session.Reply
	var id, parent, run, doc string
	var ns int64
	if err := row.Scan(&id, &parent, &run, &doc, &ns); err != nil {
		return r, err
	}
	if len(doc) > maxMessageDocument || json.Unmarshal([]byte(doc), &r) != nil || r.ID != id || r.ReplyTo != parent || r.Source.RunID != run || !session.ValidID(r.Source.ID) || (session.ReplyRequest{ID: r.ID, ReplyTo: r.ReplyTo, Body: r.Body}).Validate() != nil {
		return r, fmt.Errorf("reply identity/content conflict")
	}
	r.RecordedAt = time.Unix(0, ns).UTC()
	return r, nil
}

func (s *Store) RecordReply(ctx context.Context, c session.AgentCredential, request session.ReplyRequest) (session.Reply, error) {
	var zero session.Reply
	if err := request.Validate(); err != nil {
		return zero, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()
	source, err := s.agentIdentity(ctx, tx, c)
	if err != nil {
		return zero, err
	}
	parent, err := scanMessage(tx.QueryRowContext(ctx, messageColumns+` WHERE scope=? AND id=?`, s.scope, request.ReplyTo))
	if err != nil {
		return zero, err
	}
	if !sameMessageTarget(parent.Target, source) {
		return zero, fmt.Errorf("reply source does not match original request target")
	}
	old, err := scanReply(tx.QueryRowContext(ctx, replyColumns+` WHERE scope=? AND id=?`, s.scope, request.ID))
	if err == nil {
		if old.ReplyTo != request.ReplyTo || old.Body != request.Body || !sameMessageTarget(old.Source, source) {
			return zero, fmt.Errorf("reply ID already used for different intent")
		}
		return old, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return zero, err
	}
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM replies WHERE scope=? AND reply_to=?)`, s.scope, request.ReplyTo).Scan(&exists); err != nil {
		return zero, err
	}
	if exists {
		return zero, fmt.Errorf("request already has a reply; query replies instead of replacing it")
	}
	r := session.Reply{ID: request.ID, ReplyTo: request.ReplyTo, Source: source, Body: request.Body, RecordedAt: time.Now().UTC()}
	data, err := json.Marshal(r)
	if err != nil || len(data) > maxMessageDocument {
		return zero, fmt.Errorf("reply document exceeds bound")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO replies VALUES(?,?,?,?,?,?)`, s.scope, r.ID, r.ReplyTo, r.Source.RunID, string(data), r.RecordedAt.UnixNano())
	if err != nil {
		return zero, err
	}
	if err = tx.Commit(); err != nil {
		return zero, fmt.Errorf("reply commit unconfirmed; query replies using original IDs")
	}
	return r, nil
}

// Replies remains available offline, including after the source Run exits.
func (s *Store) Replies(ctx context.Context, id string) ([]session.Reply, error) {
	if _, err := s.Message(ctx, id); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, replyColumns+` WHERE scope=? AND reply_to=? ORDER BY created_ns,id`, s.scope, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []session.Reply{}
	for rows.Next() {
		r, err := scanReply(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
