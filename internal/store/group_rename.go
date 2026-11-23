package store

import (
	"context"
	"fmt"
	"math"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// RenameGroup edits display metadata only. The complete observed group is a
// compare-and-swap token; neither a display path nor a terminal is addressed.
func (s *Store) RenameGroup(ctx context.Context, expected session.Group, name string) (session.Group, error) {
	name, err := session.NormalizeGroupName(name)
	if err != nil {
		return session.Group{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return session.Group{}, err
	}
	defer tx.Rollback()
	groups, err := s.groups(ctx, tx)
	if err != nil {
		return session.Group{}, err
	}
	found := false
	for _, g := range groups {
		if g.ID == expected.ID {
			if g != expected {
				return session.Group{}, ErrGroupConflict
			}
			found = true
		} else if g.ParentID == expected.ParentID && g.Name == name {
			return session.Group{}, ErrGroupConflict
		}
	}
	if !found {
		return session.Group{}, ErrGroupConflict
	}
	if expected.Name == name {
		return expected, nil
	}
	if expected.Revision == math.MaxInt64 {
		return session.Group{}, fmt.Errorf("group revision exhausted")
	}
	named := expected
	named.Name, named.Revision = name, expected.Revision+1
	if _, err = tx.ExecContext(ctx, `UPDATE groups SET name=?,revision=? WHERE scope=? AND id=? AND revision=?`, named.Name, named.Revision, s.scope, expected.ID, expected.Revision); err != nil {
		return session.Group{}, err
	}
	if err = tx.Commit(); err != nil {
		return session.Group{}, fmt.Errorf("group rename commit not confirmed; refresh before another edit: %w", err)
	}
	return named, nil
}
