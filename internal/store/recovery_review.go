package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// RecoveryReview is metadata-only and reads all review facts in one transaction.
func (s *Store) RecoveryReview(ctx context.Context, id string) (recovery.Review, error) {
	var review recovery.Review
	if !session.ValidID(id) {
		return review, fmt.Errorf("full Session ID required")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return review, err
	}
	defer tx.Rollback()
	review.Source, err = s.latestRecoveryRun(ctx, tx, id)
	if err != nil {
		return review, err
	}
	var present bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM recovery_profiles WHERE scope=? AND session_id=?)`, s.scope, id).Scan(&present); err != nil {
		return review, err
	}
	if present {
		p, e := s.recoveryProfile(ctx, tx, id)
		if e != nil {
			return review, e
		}
		review.Profile = &p
	}
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM recovery_attempts WHERE scope=? AND session_id=?)`, s.scope, id).Scan(&present); err != nil {
		return review, err
	}
	if present {
		a, e := s.recoveryAttempt(ctx, tx, id)
		if e != nil {
			return review, e
		}
		review.Attempt = &a
	}
	var run, document string
	err = tx.QueryRowContext(ctx, `SELECT run_id,document FROM run_launches WHERE scope=? AND session_id=?`, s.scope, id).Scan(&run, &document)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return review, err
	}
	if err == nil && run == review.Source.RunID {
		if len(document) > 65536 {
			return review, fmt.Errorf("stored launch exceeds review limit")
		}
		var launch session.Launch
		decoder := json.NewDecoder(strings.NewReader(document))
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&launch); err != nil {
			return review, fmt.Errorf("stored launch invalid")
		}
		var tail any
		if decoder.Decode(&tail) != io.EOF {
			return review, fmt.Errorf("stored launch has trailing data")
		}
		if err = launch.Validate(); err != nil {
			return review, fmt.Errorf("stored launch invalid")
		}
		// Names and group membership are current logical metadata, not launch args.
		launch.Name = review.Source.Title
		var group string
		err = tx.QueryRowContext(ctx, `SELECT group_id FROM session_groups WHERE scope=? AND session_id=?`, s.scope, id).Scan(&group)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return review, err
		}
		launch.GroupID = group
		review.Launch = &launch
	}
	return review, tx.Commit()
}
