package store

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (s *Store) Workspace(ctx context.Context, items []session.Session) (session.Workspace, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return session.Workspace{}, err
	}
	defer tx.Rollback()
	result, err := s.workspace(ctx, tx, items)
	if err != nil {
		return session.Workspace{}, err
	}
	if err = tx.Commit(); err != nil {
		return session.Workspace{}, err
	}
	return result, nil
}

func (s *Store) workspace(ctx context.Context, q rowsReader, items []session.Session) (session.Workspace, error) {
	groups, err := s.groups(ctx, q)
	if err != nil {
		return session.Workspace{}, err
	}
	result := session.Workspace{Groups: groups, Sessions: append([]session.Session(nil), items...)}
	for i := range result.Sessions {
		v := &result.Sessions[i]
		var positionGroup string
		err = q.QueryRowContext(ctx, `SELECT coalesce(n.title,s.title),coalesce(n.revision,0),coalesce(g.group_id,''),coalesce(g.revision,0),coalesce(p.ordinal,0),coalesce(p.revision,0),coalesce(p.group_id,'') FROM sessions s LEFT JOIN session_names n ON n.scope=s.scope AND n.session_id=s.id LEFT JOIN session_groups g ON g.scope=s.scope AND g.session_id=s.id LEFT JOIN session_positions p ON p.scope=s.scope AND p.session_id=s.id WHERE s.scope=? AND s.id=?`, s.scope, v.ID).Scan(&v.Title, &v.TitleRevision, &v.GroupID, &v.GroupRevision, &v.Order, &v.OrderRevision, &positionGroup)
		if err != nil {
			return session.Workspace{}, err
		}
		if v.Order < 0 || v.OrderRevision < 0 || (v.OrderRevision == 0 && v.Order != 0) || (v.OrderRevision > 0 && positionGroup != session.GroupOf(*v)) {
			return session.Workspace{}, fmt.Errorf("invalid stored session position")
		}
		if v.TitleRevision < 0 {
			return session.Workspace{}, fmt.Errorf("invalid stored name revision")
		}
		if v.TitleRevision > 0 {
			name, e := session.NormalizeRenameTitle(v.Title)
			if e != nil || name != v.Title {
				return session.Workspace{}, fmt.Errorf("invalid stored display name")
			}
		}
		if v.GroupRevision < 0 || (v.GroupID != "" && v.GroupRevision == 0) || (v.GroupID == "" && v.GroupRevision > 0) {
			return session.Workspace{}, fmt.Errorf("invalid stored membership revision")
		}
		found := false
		for _, g := range groups {
			if g.ID == session.GroupOf(*v) {
				found = true
			}
		}
		if !found {
			return session.Workspace{}, fmt.Errorf("stored membership group missing")
		}
		if v.GroupID == session.DefaultGroupID {
			v.GroupID = ""
		}
	}
	result.OrderStamps, err = s.sessionOrderStamps(ctx, q, groups)
	if err != nil {
		return session.Workspace{}, err
	}
	session.SortSessions(result.Sessions)
	return result, nil
}

func (s *Store) MoveToGroup(ctx context.Context, expected session.Session, target string) (session.Session, error) {
	if target == "" {
		target = session.DefaultGroupID
	}
	if !validGroupID(target) {
		return session.Session{}, fmt.Errorf("invalid target group identity")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return session.Session{}, err
	}
	defer tx.Rollback()
	var current string
	var revision int64
	err = tx.QueryRowContext(ctx, `SELECT coalesce(g.group_id,'default'),coalesce(g.revision,0) FROM runs r JOIN sessions s ON s.scope=r.scope AND s.id=r.session_id LEFT JOIN session_groups g ON g.scope=s.scope AND g.session_id=s.id WHERE r.scope=? AND r.id=? AND r.session_id=? AND r.backend_id=? AND s.command=?`, s.scope, expected.RunID, expected.ID, expected.BackendID, expected.Command).Scan(&current, &revision)
	if err != nil {
		return session.Session{}, fmt.Errorf("exact metadata Run required for move: %w", err)
	}
	if current != session.GroupOf(expected) || revision != expected.GroupRevision {
		return session.Session{}, ErrGroupConflict
	}
	var exists int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM groups WHERE scope=? AND id=?`, s.scope, target).Scan(&exists); err != nil {
		return session.Session{}, err
	}
	if exists != 1 {
		return session.Session{}, fmt.Errorf("target group not found")
	}
	result := expected
	if current == target {
		return result, nil
	}
	if revision == math.MaxInt64 {
		return session.Session{}, fmt.Errorf("membership revision exhausted")
	}
	result.GroupID, result.GroupRevision = target, revision+1
	if target == session.DefaultGroupID {
		result.GroupID = ""
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO session_groups VALUES(?,?,?,?) ON CONFLICT(scope,session_id) DO UPDATE SET group_id=excluded.group_id,revision=excluded.revision`, s.scope, expected.ID, target, result.GroupRevision)
	if err != nil {
		return session.Session{}, err
	}
	if err = s.clearSessionPosition(ctx, tx, expected.ID); err != nil {
		return session.Session{}, err
	}
	result.Order, result.OrderRevision = 0, 0
	if err = tx.Commit(); err != nil {
		return session.Session{}, fmt.Errorf("membership commit not confirmed; refresh before another move: %w", err)
	}
	return result, nil
}
