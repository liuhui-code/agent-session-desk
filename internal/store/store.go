// Package store persists metadata and bounded screen records, never readiness.
package store

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	_ "modernc.org/sqlite"
)

type Store struct {
	db    *sql.DB
	scope string
	path  string
}

// Record remains a compatible name for callers of the storage boundary.
type Record = session.HistoryRecord

func Open(ctx context.Context, path, scope string) (*Store, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	fresh := err == nil
	if fresh {
		if err = f.Close(); err != nil {
			return nil, err
		}
	} else if !os.IsExist(err) {
		return nil, err
	}
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	uri := url.URL{Scheme: "file", Path: p}
	q := url.Values{"mode": {"rw"}, "_txlock": {"immediate"}, "_pragma": {"busy_timeout(1000)", "foreign_keys(1)", "synchronous(FULL)", "cache_size(-1024)"}}
	uri.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	s := &Store{db: db, scope: scope, path: path}
	if err = s.initialize(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("open metadata: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) initialize(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var app, version, tables int
	if err = tx.QueryRowContext(ctx, "PRAGMA application_id").Scan(&app); err != nil {
		return err
	}
	if err = tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if err = tx.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_schema WHERE name NOT LIKE 'sqlite_%'").Scan(&tables); err != nil {
		return err
	}
	const applicationID = 0x41534431
	if app == 0 && version == 0 && tables == 0 {
		if _, err = tx.ExecContext(ctx, `
 CREATE TABLE sessions(scope TEXT NOT NULL,id TEXT NOT NULL,title TEXT NOT NULL,command TEXT NOT NULL,PRIMARY KEY(scope,id));
 CREATE TABLE runs(scope TEXT NOT NULL,id TEXT NOT NULL,session_id TEXT NOT NULL,backend_id TEXT NOT NULL,recorded_ns INTEGER NOT NULL,stopped_ns INTEGER,PRIMARY KEY(scope,id),FOREIGN KEY(scope,session_id) REFERENCES sessions(scope,id));
 CREATE INDEX runs_recent ON runs(scope,recorded_ns DESC,id DESC);
 PRAGMA application_id=1095975985; PRAGMA user_version=2;`); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, exitSchema); err != nil {
			return err
		}
	} else if app != applicationID || (version != 1 && version != 2) {
		return fmt.Errorf("unsupported metadata application/schema; file preserved")
	}
	var journal string
	if err = tx.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journal); err != nil {
		return err
	}
	if journal != "delete" {
		return fmt.Errorf("metadata requires DELETE journal mode")
	}
	rows, err := tx.QueryContext(ctx, `SELECT s.scope,s.id,s.title,s.command,r.id,r.session_id,r.backend_id,r.recorded_ns,r.stopped_ns FROM sessions s JOIN runs r ON s.scope=r.scope AND s.id=r.session_id LIMIT 0`)
	if err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	if version == 1 {
		if _, err = tx.ExecContext(ctx, exitSchema+" PRAGMA user_version=2;"); err != nil {
			return err
		}
	}
	rows, err = tx.QueryContext(ctx, `SELECT scope,run_id,pane_id,source,observed_ns FROM pane_exits LIMIT 0`)
	if err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Save(ctx context.Context, v session.Session) error {
	if err := validateExit(v); err != nil {
		return err
	}
	for _, id := range []string{v.ID, v.RunID} {
		if len(id) != 32 {
			return fmt.Errorf("invalid metadata identity")
		}
		if _, err := hex.DecodeString(id); err != nil {
			return fmt.Errorf("invalid metadata identity")
		}
	}
	if v.BackendID == "" {
		return fmt.Errorf("backend identity required")
	}
	for _, value := range []string{v.BackendID, v.Title, v.Command} {
		if len(value) > 4096 || strings.ContainsRune(value, 0) {
			return fmt.Errorf("metadata field exceeds transport bounds")
		}
	}

	if found, err := s.matches(ctx, s.db, v); err != nil {
		return err
	} else if found {
		if match, err := s.matchesExit(ctx, s.db, v); err != nil || match {
			return err
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO sessions VALUES(?,?,?,?) ON CONFLICT DO NOTHING`, s.scope, v.ID, v.Title, v.Command); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO runs VALUES(?,?,?,?,?,NULL) ON CONFLICT DO NOTHING`, s.scope, v.RunID, v.ID, v.BackendID, time.Now().UnixNano()); err != nil {
		return err
	}
	if found, err := s.matches(ctx, tx, v); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("metadata Run missing after insert")
	}
	if err = s.saveExit(ctx, tx, v); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) History(ctx context.Context, limit int) ([]Record, error) {
	if limit < 1 || limit > 1000 {
		return nil, fmt.Errorf("history limit must be 1–1000")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT s.id,r.id,r.backend_id,s.title,s.command,r.recorded_ns,r.stopped_ns,e.pane_id,e.source,e.observed_ns FROM runs r JOIN sessions s ON s.scope=r.scope AND s.id=r.session_id LEFT JOIN pane_exits e ON e.scope=r.scope AND e.run_id=r.id WHERE r.scope=? ORDER BY r.recorded_ns DESC,r.id DESC LIMIT ?`, s.scope, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Record{}
	for rows.Next() {
		var r Record
		var at int64
		var stopped, exitAt sql.NullInt64
		var pane, source sql.NullString
		if err = rows.Scan(&r.Session.ID, &r.Session.RunID, &r.Session.BackendID, &r.Session.Title, &r.Session.Command, &at, &stopped, &pane, &source, &exitAt); err != nil {
			return nil, err
		}
		r.Session.TerminalState = session.TerminalUnknown
		r.RecordedAt = time.Unix(0, at).UTC()
		if stopped.Valid {
			t := time.Unix(0, stopped.Int64).UTC()
			r.StopConfirmedAt = &t
		}
		if pane.Valid {
			r.PaneExit = &session.PaneExit{PaneID: pane.String, Source: source.String, ObservedAt: time.Unix(0, exitAt.Int64).UTC()}
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

// MarkStopped records only an explicit successful stop of this backend target.
// It is not an agent exit code or an inference from a missing list entry.
func (s *Store) MarkStopped(ctx context.Context, v session.Session) error {
	r, err := s.db.ExecContext(ctx, `UPDATE runs SET stopped_ns=COALESCE(stopped_ns,?) WHERE scope=? AND id=? AND session_id=? AND backend_id=?`, time.Now().UnixNano(), s.scope, v.RunID, v.ID, v.BackendID)
	if err != nil {
		return err
	}
	n, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("no exact metadata Run to mark stopped")
	}
	return nil
}

// CheckWritable catches current busy/read-only failures before process creation.
// There is deliberately no SQL transaction held across an external backend call.
func (s *Store) CheckWritable(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	return tx.Rollback()
}

type rowReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s *Store) matches(ctx context.Context, q rowReader, v session.Session) (bool, error) {
	var title, command, owner, backendID string
	err := q.QueryRowContext(ctx, `SELECT s.title,s.command,r.session_id,r.backend_id FROM runs r JOIN sessions s ON s.scope=r.scope AND s.id=r.session_id WHERE r.scope=? AND r.id=?`, s.scope, v.RunID).Scan(&title, &command, &owner, &backendID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if title != v.Title || command != v.Command || owner != v.ID || backendID != v.BackendID {
		return true, fmt.Errorf("metadata identity conflict; existing record preserved")
	}
	return true, nil
}
