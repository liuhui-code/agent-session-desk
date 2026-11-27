package store

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const attentionSnoozeSchema = `ALTER TABLE attention ADD COLUMN snoozed_ns INTEGER;
ALTER TABLE attention ADD COLUMN reminder_revision INTEGER NOT NULL DEFAULT 0 CHECK(reminder_revision>=0);`

// SnoozeAttention changes human reminder policy, never lifecycle or input rights.
// Duration zero resumes reminders. The separate revision rejects competing edits.
func (s *Store) SnoozeAttention(ctx context.Context, expected session.AttentionItem, duration time.Duration) (session.AttentionItem, error) {
	if !session.ValidID(expected.ID) || (duration != 0 && duration != 10*time.Minute && duration != 30*time.Minute) {
		return session.AttentionItem{}, fmt.Errorf("snooze requires an exact item and 0, 10 or 30 minutes")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return session.AttentionItem{}, err
	}
	defer tx.Rollback()
	if err = s.invalidateAttention(ctx, tx); err != nil {
		return session.AttentionItem{}, err
	}
	item, _, err := scanAttention(tx.QueryRowContext(ctx, attentionColumns+` WHERE scope=? AND id=?`, s.scope, expected.ID))
	if err != nil {
		return item, err
	}
	if item.State != "pending" || item.Version != expected.Version || item.ReminderRevision != expected.ReminderRevision || !sameMessageTarget(item.Source, expected.Source) || item.ReminderRevision == math.MaxInt64 {
		return item, fmt.Errorf("reminder changed; refresh before editing")
	}
	var value any
	item.SnoozedUntil = nil
	if duration != 0 {
		until := time.Now().UTC().Add(duration)
		item.SnoozedUntil, value = &until, until.UnixNano()
	}
	_, err = tx.ExecContext(ctx, `UPDATE attention SET snoozed_ns=?,reminder_revision=reminder_revision+1 WHERE scope=? AND id=?`, value, s.scope, item.ID)
	if err != nil {
		return item, err
	}
	if err = tx.Commit(); err != nil {
		return item, fmt.Errorf("reminder commit unconfirmed; refresh without retry: %w", err)
	}
	item.ReminderRevision++
	return item, nil
}
