// Container removal semantics adapted from Agent Deck, Copyright (c) 2025
// Ashesh Goplani. SPDX-License-Identifier: MIT
// Source: internal/session/groups.go, DeleteGroup, at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: stable scoped metadata,
// reviewed snapshot CAS, historical members and atomic membership/tree updates.
package store

import (
	"context"
	"fmt"
	"slices"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (s *Store) RemoveGroup(ctx context.Context, expected session.GroupRemoval, observed []session.Session) (session.Workspace, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return session.Workspace{}, err
	}
	defer tx.Rollback()
	state, err := s.groupRemovalState(ctx, tx, expected.Root.ID)
	if err != nil {
		return session.Workspace{}, err
	}
	actual := state.plan
	if actual.Root != expected.Root || actual.Destination != expected.Destination || actual.MemberCount != expected.MemberCount || actual.Stamp != expected.Stamp || !slices.Equal(actual.Groups, expected.Groups) {
		return session.Workspace{}, ErrGroupConflict
	}
	if state.exhausted {
		return session.Workspace{}, fmt.Errorf("membership revision exhausted")
	}
	for _, g := range actual.Groups {
		if _, err = tx.ExecContext(ctx, `UPDATE session_groups SET group_id='default',revision=revision+1 WHERE scope=? AND group_id=?`, s.scope, g.ID); err != nil {
			return session.Workspace{}, err
		}
	}
	for i := len(actual.Groups) - 1; i >= 0; i-- {
		if _, err = tx.ExecContext(ctx, `DELETE FROM groups WHERE scope=? AND id=?`, s.scope, actual.Groups[i].ID); err != nil {
			return session.Workspace{}, err
		}
	}
	// Project only observed rows inside this same transaction. Terminal/Run
	// observations are copied, never replaced by invented online evidence.
	view, err := s.workspace(ctx, tx, observed)
	if err != nil {
		return session.Workspace{}, err
	}
	if err = tx.Commit(); err != nil {
		return session.Workspace{}, fmt.Errorf("group removal commit not confirmed; refresh before another edit: %w", err)
	}
	return view, nil
}
