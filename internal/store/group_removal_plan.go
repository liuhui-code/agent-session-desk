package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type groupRemovalState struct {
	plan      session.GroupRemoval
	exhausted bool
}

func (s *Store) PreviewGroupRemoval(ctx context.Context, id string) (session.GroupRemoval, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return session.GroupRemoval{}, err
	}
	defer tx.Rollback()
	state, err := s.groupRemovalState(ctx, tx, id)
	if err != nil {
		return session.GroupRemoval{}, err
	}
	if err = tx.Commit(); err != nil {
		return session.GroupRemoval{}, err
	}
	return state.plan, nil
}

func (s *Store) groupRemovalState(ctx context.Context, q rowsReader, id string) (groupRemovalState, error) {
	state := groupRemovalState{}
	if id == session.DefaultGroupID {
		return state, fmt.Errorf("default group cannot be removed")
	}
	groups, err := s.groups(ctx, q)
	if err != nil {
		return state, err
	}
	removed := map[string]bool{}
	for _, g := range groups {
		if g.ID == session.DefaultGroupID {
			state.plan.Destination = g
		}
		if g.ID == id {
			state.plan.Root = g
		}
		if g.ID == id || removed[g.ParentID] {
			removed[g.ID] = true
			state.plan.Groups = append(state.plan.Groups, g)
		}
	}
	if state.plan.Root.ID == "" {
		return state, ErrGroupConflict
	}
	hash := sha256.New()
	encoder := json.NewEncoder(hash)
	// Scope/tree names and versions make both the selected structure and the
	// named destination part of review. The digest is not authorization.
	if err = encoder.Encode([]string{s.scope, id}); err != nil {
		return state, err
	}
	if err = encoder.Encode(groups); err != nil {
		return state, err
	}
	rows, err := q.QueryContext(ctx, `SELECT session_id,group_id,revision FROM session_groups WHERE scope=? ORDER BY session_id`, s.scope)
	if err != nil {
		return state, err
	}
	defer rows.Close()
	for rows.Next() {
		var member struct {
			SessionID, GroupID string
			Revision           int64
		}
		if err = rows.Scan(&member.SessionID, &member.GroupID, &member.Revision); err != nil {
			return state, err
		}
		if !removed[member.GroupID] {
			continue
		}
		if member.SessionID == session.DefaultGroupID || !validGroupID(member.SessionID) || member.Revision < 1 {
			return state, fmt.Errorf("invalid affected membership")
		}
		if state.plan.MemberCount == math.MaxInt64 {
			return state, fmt.Errorf("member count exhausted")
		}
		state.plan.MemberCount++
		state.exhausted = state.exhausted || member.Revision == math.MaxInt64
		if err = encoder.Encode(member); err != nil {
			return state, err
		}
	}
	if err = rows.Err(); err != nil {
		return state, err
	}
	state.plan.Stamp = hex.EncodeToString(hash.Sum(nil))
	return state, nil
}
