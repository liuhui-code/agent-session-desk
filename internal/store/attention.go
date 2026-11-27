package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const attentionSchema = `CREATE TABLE attention(
 scope TEXT NOT NULL,id TEXT NOT NULL,run_id TEXT NOT NULL,version INTEGER NOT NULL CHECK(version>0),
 fingerprint TEXT NOT NULL,document TEXT NOT NULL CHECK(length(CAST(document AS BLOB))<=24576),
 state TEXT NOT NULL CHECK(state IN ('pending','resolved','withdrawn','invalidated')),
 viewed_ns INTEGER,created_ns INTEGER NOT NULL,updated_ns INTEGER NOT NULL,
 PRIMARY KEY(scope,id),FOREIGN KEY(scope,run_id) REFERENCES runs(scope,id));`
const attentionV9Columns = `SELECT id,run_id,version,fingerprint,document,state,viewed_ns,created_ns,updated_ns FROM attention`
const attentionColumns = `SELECT id,run_id,version,fingerprint,document,state,viewed_ns,created_ns,updated_ns,snoozed_ns,reminder_revision FROM attention`

func scanAttention(row messageRow) (session.AttentionItem, string, error) {
	var item session.AttentionItem
	var id, run, fp, doc, state string
	var version, created, updated int64
	var viewed, snoozed sql.NullInt64
	var revision int64
	if err := row.Scan(&id, &run, &version, &fp, &doc, &state, &viewed, &created, &updated, &snoozed, &revision); err != nil {
		return item, fp, err
	}
	if len(doc) > 24576 || json.Unmarshal([]byte(doc), &item) != nil || item.ID != id || item.Source.RunID != run {
		return item, fp, fmt.Errorf("attention identity/document conflict")
	}
	item.Version, item.State = version, state
	item.EventID = id + "." + strconv.FormatInt(version, 10)
	item.CreatedAt, item.UpdatedAt = time.Unix(0, created).UTC(), time.Unix(0, updated).UTC()
	item.ViewedAt = nil
	item.SnoozedUntil, item.ReminderRevision = nil, revision
	if revision < 0 || (snoozed.Valid && snoozed.Int64 <= 0) {
		return item, fp, fmt.Errorf("invalid attention reminder metadata")
	}
	if snoozed.Valid {
		v := time.Unix(0, snoozed.Int64).UTC()
		item.SnoozedUntil = &v
	}
	if viewed.Valid {
		v := time.Unix(0, viewed.Int64).UTC()
		item.ViewedAt = &v
	}
	item.TerminalAvailable = false
	return item, fp, nil
}

func (s *Store) invalidateAttention(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `UPDATE attention SET state='invalidated',updated_ns=? WHERE scope=? AND state='pending' AND (EXISTS(SELECT 1 FROM runs r WHERE r.scope=attention.scope AND r.id=attention.run_id AND r.stopped_ns IS NOT NULL) OR EXISTS(SELECT 1 FROM pane_exits e WHERE e.scope=attention.scope AND e.run_id=attention.run_id))`, time.Now().UTC().UnixNano(), s.scope)
	return err
}

func (s *Store) ReportAttention(ctx context.Context, c session.AgentCredential, e session.AttentionEvent) (session.AttentionItem, error) {
	var zero session.AttentionItem
	fp, err := e.Fingerprint()
	if err != nil {
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
	old, oldFP, err := scanAttention(tx.QueryRowContext(ctx, attentionColumns+` WHERE scope=? AND id=?`, s.scope, e.ID))
	now := time.Now().UTC()
	created := now
	var viewed, snoozed any
	var revision int64
	if err == nil {
		if !sameMessageTarget(old.Source, source) || old.Kind != e.Kind {
			return zero, fmt.Errorf("attention identity already belongs to another source/kind")
		}
		if e.Version < old.Version {
			return old, nil
		}
		if e.Version == old.Version {
			if fp != oldFP {
				return zero, fmt.Errorf("attention version has conflicting content")
			}
			return old, nil
		}
		if old.State != "pending" && e.State == "pending" {
			return zero, fmt.Errorf("ended attention cannot be reopened; new request ID required")
		}
		created = old.CreatedAt
		if old.ReminderRevision == int64(^uint64(0)>>1) {
			return zero, fmt.Errorf("attention reminder revision exhausted")
		}
		revision = old.ReminderRevision + 1
		if e.State == "pending" && old.Title == e.Title && old.Detail == e.Detail && old.SnoozedUntil != nil {
			snoozed = old.SnoozedUntil.UnixNano()
		}
		if old.ViewedAt != nil && old.Title == e.Title && old.Detail == e.Detail {
			viewed = old.ViewedAt.UnixNano()
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return zero, err
	} else {
		var n int
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM attention`).Scan(&n); err != nil {
			return zero, err
		}
		if n >= session.MaxAttentionItems {
			return zero, fmt.Errorf("attention capacity reached; no evidence pruned")
		}
	}
	item := session.AttentionItem{AttentionEvent: e, Source: source, CreatedAt: created, UpdatedAt: now}
	data, err := json.Marshal(item)
	if err != nil || len(data) > 24576 {
		return zero, fmt.Errorf("attention document exceeds bound")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO attention VALUES(?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(scope,id) DO UPDATE SET version=excluded.version,fingerprint=excluded.fingerprint,document=excluded.document,state=excluded.state,viewed_ns=excluded.viewed_ns,updated_ns=excluded.updated_ns,snoozed_ns=excluded.snoozed_ns,reminder_revision=excluded.reminder_revision`, s.scope, e.ID, source.RunID, e.Version, fp, string(data), e.State, viewed, created.UnixNano(), now.UnixNano(), snoozed, revision)
	if err != nil {
		return zero, err
	}
	if err = tx.Commit(); err != nil {
		return zero, fmt.Errorf("attention commit unconfirmed; query original attention ID")
	}
	return s.Attention(ctx, e.ID)
}

func (s *Store) Attention(ctx context.Context, id string) (session.AttentionItem, error) {
	if !session.ValidID(id) {
		return session.AttentionItem{}, fmt.Errorf("full attention ID required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return session.AttentionItem{}, err
	}
	defer tx.Rollback()
	if err = s.invalidateAttention(ctx, tx); err != nil {
		return session.AttentionItem{}, err
	}
	item, _, err := scanAttention(tx.QueryRowContext(ctx, attentionColumns+` WHERE scope=? AND id=?`, s.scope, id))
	if err != nil {
		return item, err
	}
	return item, tx.Commit()
}

func (s *Store) PendingAttention(ctx context.Context) ([]session.AttentionItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = s.invalidateAttention(ctx, tx); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, attentionColumns+` WHERE scope=? AND state='pending' ORDER BY created_ns,id LIMIT ?`, s.scope, session.MaxAttentionItems)
	if err != nil {
		return nil, err
	}
	items := []session.AttentionItem{}
	for rows.Next() {
		item, _, e := scanAttention(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		items = append(items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	return items, tx.Commit()
}

func (s *Store) MarkAttentionViewed(ctx context.Context, id string, version int64, target session.Session) (session.AttentionItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return session.AttentionItem{}, err
	}
	defer tx.Rollback()
	if err = s.invalidateAttention(ctx, tx); err != nil {
		return session.AttentionItem{}, err
	}
	item, _, err := scanAttention(tx.QueryRowContext(ctx, attentionColumns+` WHERE scope=? AND id=?`, s.scope, id))
	if err != nil {
		return item, err
	}
	if item.State != "pending" || item.Version != version || !sameMessageTarget(item.Source, target) {
		return item, fmt.Errorf("attention changed or original Run unavailable")
	}
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE attention SET viewed_ns=? WHERE scope=? AND id=?`, now.UnixNano(), s.scope, id); err != nil {
		return item, err
	}
	if err = tx.Commit(); err != nil {
		return item, err
	}
	item.ViewedAt = &now
	return item, nil
}
