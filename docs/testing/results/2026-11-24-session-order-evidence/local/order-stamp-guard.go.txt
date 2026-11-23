// Sibling swap/normalization adapted from Agent Deck, Copyright (c) 2025
// Ashesh Goplani. SPDX-License-Identifier: MIT
// Source: internal/session/groups.go, MoveSessionUp/Down at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: scoped logical metadata,
// full observed-row CAS, transactional versions and no child/archived model.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"math"
)

const positionSchema = `CREATE TABLE session_positions(scope TEXT NOT NULL,session_id TEXT NOT NULL,group_id TEXT NOT NULL,ordinal INTEGER NOT NULL CHECK(ordinal>=0),revision INTEGER NOT NULL CHECK(revision>0),PRIMARY KEY(scope,session_id),FOREIGN KEY(scope,session_id) REFERENCES sessions(scope,id),FOREIGN KEY(scope,group_id) REFERENCES groups(scope,id) ON DELETE CASCADE);`
const positionColumns = `SELECT scope,session_id,group_id,ordinal,revision FROM session_positions LIMIT 0`

var ErrSessionOrderConflict = errors.New("session order snapshot changed; refresh before editing")

func validatePositions(ctx context.Context, q rowsReader) error {
	var invalid int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM session_positions p LEFT JOIN sessions s ON s.scope=p.scope AND s.id=p.session_id LEFT JOIN groups g ON g.scope=p.scope AND g.id=p.group_id LEFT JOIN session_groups m ON m.scope=p.scope AND m.session_id=p.session_id WHERE s.id IS NULL OR g.id IS NULL OR p.ordinal<0 OR p.revision<1 OR p.group_id<>coalesce(m.group_id,'default')`).Scan(&invalid)
	if err != nil {
		return err
	}
	if invalid != 0 {
		return fmt.Errorf("invalid stored session position")
	}
	return nil
}

func (s *Store) ReorderSession(ctx context.Context, expected session.SessionOrder, delta int, observed []session.Session) (session.Workspace, error) {
	if delta != -1 && delta != 1 {
		return session.Workspace{}, fmt.Errorf("session order direction must be -1 or 1")
	}
	seen := map[string]bool{}
	for _, v := range observed {
		if seen[v.ID] {
			return session.Workspace{}, ErrSessionOrderConflict
		}
		seen[v.ID] = true
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return session.Workspace{}, err
	}
	defer tx.Rollback()
	view, err := s.workspace(ctx, tx, observed)
	if err != nil {
		return session.Workspace{}, err
	}
	peers := session.OrderPeers(view.Sessions, expected.TargetID)
	if len(peers) == 0 || expected.Stamp == "" || view.OrderStamps[session.GroupOf(peers[0])] != expected.Stamp || !session.SameOrderSnapshot(peers, expected.Siblings) {
		return session.Workspace{}, ErrSessionOrderConflict
	}
	index := -1
	for i, v := range peers {
		var count int
		err = tx.QueryRowContext(ctx, `SELECT count(*) FROM runs r JOIN sessions s ON s.scope=r.scope AND s.id=r.session_id WHERE r.scope=? AND r.id=? AND r.session_id=? AND r.backend_id=? AND s.command=? AND r.stopped_ns IS NULL`, s.scope, v.RunID, v.ID, v.BackendID, v.Command).Scan(&count)
		if err != nil {
			return session.Workspace{}, err
		}
		if count != 1 {
			return session.Workspace{}, ErrSessionOrderConflict
		}
		if v.ID == expected.TargetID {
			index = i
		}
	}
	neighbor := index + delta
	if neighbor < 0 || neighbor >= len(peers) {
		return view, nil
	}
	for i, v := range peers {
		ordinal := i
		if i == index {
			ordinal = neighbor
		} else if i == neighbor {
			ordinal = index
		}
		if v.OrderRevision > 0 && v.Order == int64(ordinal) && i != index && i != neighbor {
			continue
		}
		if v.OrderRevision == math.MaxInt64 {
			return session.Workspace{}, fmt.Errorf("session order revision exhausted")
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO session_positions VALUES(?,?,?,?,?) ON CONFLICT(scope,session_id) DO UPDATE SET group_id=excluded.group_id,ordinal=excluded.ordinal,revision=excluded.revision`, s.scope, v.ID, session.GroupOf(v), ordinal, v.OrderRevision+1)
		if err != nil {
			return session.Workspace{}, err
		}
	}
	view, err = s.workspace(ctx, tx, observed)
	if err != nil {
		return session.Workspace{}, err
	}
	if err = tx.Commit(); err != nil {
		return session.Workspace{}, fmt.Errorf("session order commit not confirmed; refresh before another edit: %w", err)
	}
	return view, nil
}

// A group move invalidates the old per-group position in the same transaction.
func (s *Store) clearSessionPosition(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM session_positions WHERE scope=? AND session_id=?`, s.scope, id)
	return err
}
