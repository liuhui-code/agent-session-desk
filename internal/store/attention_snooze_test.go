package store_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestAttentionSnoozeSurvivesReopenWithoutResolvingOrUnblocking(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	v := sample()
	c := replyCredential(v)
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	if err = db.BindAgent(ctx, c, v); err != nil {
		t.Fatal(err)
	}
	item, err := db.ReportAttention(ctx, c, session.AttentionEvent{ID: strings.Repeat("a", 32), Version: 1, Kind: "approval", State: "pending", Title: "Allow?"})
	if err != nil {
		t.Fatal(err)
	}
	before := time.Now()
	item, err = db.SnoozeAttention(ctx, item, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if item.State != "pending" || item.ViewedAt != nil || item.SnoozedUntil == nil || item.ReminderDue(before) || !item.ReminderDue(before.Add(11*time.Minute)) {
		t.Fatal("snooze changed the request or did not expire")
	}
	db.Close()
	db, err = store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.PendingAttention(ctx)
	if err != nil || len(rows) != 1 || rows[0].SnoozedUntil == nil || !rows[0].SnoozedUntil.Equal(*item.SnoozedUntil) {
		t.Fatal("persistent pending/snooze lost", err)
	}
	msg := session.Message{ID: strings.Repeat("b", 32), Target: v, Body: "do not answer approval", Fingerprint: strings.Repeat("c", 64)}
	if _, err = db.QueueMessage(ctx, msg); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ClaimMessage(ctx, msg.ID, v); err == nil {
		t.Fatal("snooze unblocked automatic input")
	}
	if _, err = db.ClaimTask(ctx, c, newClaimKey(t, msg.ID), 300); err == nil {
		t.Fatal("snooze unblocked a new task")
	}
}
