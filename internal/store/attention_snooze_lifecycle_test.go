package store_test

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestAttentionSnoozeRejectsStaleEditsAndResetsOnChangedQuestion(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "db"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := sample()
	c := replyCredential(v)
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	if err = db.BindAgent(ctx, c, v); err != nil {
		t.Fatal(err)
	}
	e := session.AttentionEvent{ID: strings.Repeat("d", 32), Version: 1, Kind: "question", State: "pending", Title: "Which option?"}
	old, err := db.ReportAttention(ctx, c, e)
	if err != nil {
		t.Fatal(err)
	}
	current, err := db.SnoozeAttention(ctx, old, 30*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.SnoozeAttention(ctx, old, 0); err == nil {
		t.Fatal("stale resume overwrote newer snooze")
	}
	other := current
	other.Source.RunID = strings.Repeat("e", 32)
	if _, err = db.SnoozeAttention(ctx, other, 0); err == nil {
		t.Fatal("another Run changed reminder")
	}
	if _, err = db.SnoozeAttention(ctx, current, time.Minute); err == nil {
		t.Fatal("unsupported duration accepted")
	}
	e.Version++
	same, err := db.ReportAttention(ctx, c, e)
	if err != nil || same.SnoozedUntil == nil {
		t.Fatal("same question update lost snooze", err)
	}
	e.Version++
	e.Title = "Changed operation"
	changed, err := db.ReportAttention(ctx, c, e)
	if err != nil || changed.SnoozedUntil != nil || !changed.ReminderDue(time.Now()) {
		t.Fatal("new question remained hidden", err)
	}
	current, err = db.SnoozeAttention(ctx, changed, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := db.SnoozeAttention(ctx, current, 0)
	if err != nil || !resumed.ReminderDue(time.Now()) {
		t.Fatal("resume failed", err)
	}
	if err = db.MarkStopped(ctx, v); err != nil {
		t.Fatal(err)
	}
	if _, err = db.SnoozeAttention(ctx, resumed, 10*time.Minute); err == nil {
		t.Fatal("ended Run request changed")
	}
}

func TestAttentionSnoozeV9MigrationRecoveryAndConflictPreservation(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "db")
	db, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	removeRecoveryV11(t, raw)
	if _, err = raw.Exec(`ALTER TABLE attention DROP COLUMN snoozed_ns; ALTER TABLE attention DROP COLUMN reminder_revision; PRAGMA user_version=9;`); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	check, err := store.Check(ctx, path)
	if err != nil || check.SchemaVersion != 9 {
		t.Fatal("v9 offline check", check, err)
	}
	backup := filepath.Join(t.TempDir(), "backup")
	if err = store.Snapshot(ctx, path, backup); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, backup, "a")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	check, err = store.Check(ctx, backup)
	if err != nil || check.SchemaVersion != 11 {
		t.Fatal("v9 migration", check, err)
	}
	raw, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`ALTER TABLE attention ADD COLUMN snoozed_ns INTEGER;`); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	before, _ := os.ReadFile(path)
	if opened, err := store.Open(ctx, path, "a"); err == nil {
		opened.Close()
		t.Fatal("conflicting migration accepted")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("migration conflict changed source")
	}
}
