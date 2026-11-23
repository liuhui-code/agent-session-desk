// Sibling normalization/swap adapted from Agent Deck, Copyright (c) 2025
// Ashesh Goplani. SPDX-License-Identifier: MIT
// Source: internal/session/groups.go, moveGroupAmongSiblings, at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: stable IDs, whole-sibling
// CAS, scoped transactional metadata and the product's reserved default group.
package store

import (
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (s *Store) ReorderGroup(ctx context.Context, expected session.GroupOrder, delta int) ([]session.Group, error) {
	if delta != -1 && delta != 1 {
		return nil, fmt.Errorf("group order direction must be -1 or 1")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	groups, err := s.groups(ctx, tx)
	if err != nil {
		return nil, err
	}
	var target session.Group
	for _, g := range groups {
		if g.ID == expected.TargetID {
			target = g
			break
		}
	}
	if target.ID == "" {
		return nil, ErrGroupConflict
	}
	var siblings []session.Group
	index := -1
	for _, g := range groups {
		if g.ParentID == target.ParentID {
			if g.ID == target.ID {
				index = len(siblings)
			}
			siblings = append(siblings, g)
		}
	}
	if !slices.Equal(siblings, expected.Siblings) {
		return nil, ErrGroupConflict
	}
	neighbor := index + delta
	if target.ID == session.DefaultGroupID || neighbor < 0 || neighbor >= len(siblings) || siblings[neighbor].ID == session.DefaultGroupID {
		return groups, nil
	}
	var updates []session.Group
	for i, g := range siblings {
		order := i
		if i == index {
			order = neighbor
		} else if i == neighbor {
			order = index
		}
		if g.Order != int64(order) || i == index || i == neighbor {
			if g.Revision == math.MaxInt64 {
				return nil, fmt.Errorf("group revision exhausted")
			}
			g.Order, g.Revision = int64(order), g.Revision+1
			updates = append(updates, g)
		}
	}
	for _, g := range updates {
		if _, err = tx.ExecContext(ctx, `UPDATE groups SET ordinal=?,revision=? WHERE scope=? AND id=?`, g.Order, g.Revision, s.scope, g.ID); err != nil {
			return nil, err
		}
	}
	ordered, err := s.groups(ctx, tx)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("group order commit not confirmed; refresh before another edit: %w", err)
	}
	return ordered, nil
}
