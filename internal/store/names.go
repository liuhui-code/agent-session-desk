package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const nameSchema = `CREATE TABLE session_names(scope TEXT NOT NULL,session_id TEXT NOT NULL,title TEXT NOT NULL CHECK(length(title) BETWEEN 1 AND 128 AND length(CAST(title AS BLOB))<=512),revision INTEGER NOT NULL CHECK(revision>0),PRIMARY KEY(scope,session_id),FOREIGN KEY(scope,session_id) REFERENCES sessions(scope,id));`

var ErrNameConflict = errors.New("session name changed; refresh before renaming")

// WithNames projects only the requested live rows, not all archived labels.
func (s *Store) WithNames(ctx context.Context, items []session.Session) ([]session.Session, error) {
	result := append([]session.Session(nil), items...)
	for i := range result {
		var title string
		var revision int64
		err := s.db.QueryRowContext(ctx, `SELECT title,revision FROM session_names WHERE scope=? AND session_id=?`, s.scope, result[i].ID).Scan(&title, &revision)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if _, err = session.NormalizeRenameTitle(title); err != nil || revision < 1 {
			return nil, fmt.Errorf("invalid stored display name")
		}
		result[i].Title, result[i].TitleRevision = title, revision
	}
	return result, nil
}

// Rename commits one product-owned label. Backend identifiers/launch snapshots
// never change and no terminal callback runs while holding the SQL transaction.
func (s *Store) Rename(ctx context.Context, expected session.Session, value string) (session.Session, error) {
	title, err := session.NormalizeRenameTitle(value)
	if err != nil {
		return session.Session{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return session.Session{}, err
	}
	defer tx.Rollback()
	var current string
	var revision int64
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(n.title,s.title),COALESCE(n.revision,0) FROM runs r JOIN sessions s ON s.scope=r.scope AND s.id=r.session_id LEFT JOIN session_names n ON n.scope=s.scope AND n.session_id=s.id WHERE r.scope=? AND r.id=? AND r.session_id=? AND r.backend_id=? AND s.command=?`, s.scope, expected.RunID, expected.ID, expected.BackendID, expected.Command).Scan(&current, &revision)
	if err != nil {
		return session.Session{}, fmt.Errorf("exact metadata Run required for rename: %w", err)
	}
	if revision != expected.TitleRevision || current != expected.Title {
		return session.Session{}, ErrNameConflict
	}
	result := expected
	result.Title = title
	if title == current {
		return result, nil
	}
	if revision == math.MaxInt64 {
		return session.Session{}, fmt.Errorf("name revision exhausted")
	}
	result.TitleRevision = revision + 1
	_, err = tx.ExecContext(ctx, `INSERT INTO session_names(scope,session_id,title,revision) VALUES(?,?,?,?) ON CONFLICT(scope,session_id) DO UPDATE SET title=excluded.title,revision=excluded.revision`, s.scope, expected.ID, title, result.TitleRevision)
	if err != nil {
		return session.Session{}, err
	}
	if err = tx.Commit(); err != nil {
		return session.Session{}, fmt.Errorf("name commit not confirmed; refresh to reconcile: %w", err)
	}
	return result, nil
}
