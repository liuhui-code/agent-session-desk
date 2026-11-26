package store_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestAttentionViewedIsNotResolvedAndBlocksDelivery(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
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
	e := session.AttentionEvent{ID: strings.Repeat("a", 32), Version: 1, Kind: "approval", State: "pending", Title: "Allow operation?", Detail: "Choose in the original terminal"}
	item, err := db.ReportAttention(ctx, c, e)
	if err != nil || item.State != "pending" {
		t.Fatalf("event %+v %v", item, err)
	}
	if _, err = db.MarkAttentionViewed(ctx, e.ID, e.Version, v); err != nil {
		t.Fatal(err)
	}
	rows, err := db.PendingAttention(ctx)
	if err != nil || len(rows) != 1 || rows[0].ViewedAt == nil {
		t.Fatal("viewing resolved or lost the request")
	}
	msg := session.Message{ID: strings.Repeat("b", 32), Target: v, Body: "must not reach approval form", Fingerprint: strings.Repeat("c", 64)}
	if _, err = db.QueueMessage(ctx, msg); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ClaimMessage(ctx, msg.ID, v); err == nil {
		t.Fatal("automatic input entered a pending approval")
	}
	if _, err = db.ClaimTask(ctx, c, newClaimKey(t, msg.ID), 300); err == nil {
		t.Fatal("new automatic work started while a decision was pending")
	}
	e.Version = 2
	e.State = "resolved"
	if _, err = db.ReportAttention(ctx, c, e); err != nil {
		t.Fatal(err)
	}
	rows, err = db.PendingAttention(ctx)
	if err != nil || len(rows) != 0 {
		t.Fatal("source resolution did not clear pending")
	}
	if _, err = db.ClaimMessage(ctx, msg.ID, v); err != nil {
		t.Fatal("resolved attention still blocked explicit delivery")
	}
}

func TestAttentionMigrationConflictRollsBackAndSnapshotPreservesReadState(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "db.sqlite3")
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
	e := session.AttentionEvent{ID: strings.Repeat("a", 32), Version: 1, Kind: "question", State: "pending", Title: "Choose"}
	if _, err = db.ReportAttention(ctx, c, e); err != nil {
		t.Fatal(err)
	}
	if _, err = db.MarkAttentionViewed(ctx, e.ID, 1, v); err != nil {
		t.Fatal(err)
	}
	db.Close()
	backup := filepath.Join(t.TempDir(), "backup.sqlite3")
	if err = store.Snapshot(ctx, path, backup); err != nil {
		t.Fatal(err)
	}
	restored, err := store.Open(ctx, backup, "a")
	if err != nil {
		t.Fatal(err)
	}
	item, err := restored.Attention(ctx, e.ID)
	if err != nil || item.ViewedAt == nil || item.State != "pending" || item.TerminalAvailable {
		t.Fatal("snapshot invented resolution/liveness or lost read state", err)
	}
	restored.Close()
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`PRAGMA user_version=8`); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	before, _ := os.ReadFile(path)
	if opened, err := store.Open(ctx, path, "a"); err == nil {
		opened.Close()
		t.Fatal("conflicting v8 table accepted")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("failed migration changed source")
	}
}

func attentionFixture(t *testing.T) (*store.Store, session.AgentCredential, session.AttentionEvent) {
	t.Helper()
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	v := sample()
	c := replyCredential(v)
	if err = db.Save(context.Background(), v); err != nil {
		t.Fatal(err)
	}
	if err = db.BindAgent(context.Background(), c, v); err != nil {
		t.Fatal(err)
	}
	return db, c, session.AttentionEvent{ID: strings.Repeat("a", 32), Version: 1, Kind: "question", State: "pending", Title: "Choose in original terminal"}
}

func TestAttentionVersionOrderingTombstonesAndViewedRevision(t *testing.T) {
	db, c, e := attentionFixture(t)
	ctx := context.Background()
	first, err := db.ReportAttention(ctx, c, e)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.MarkAttentionViewed(ctx, e.ID, 1, sample()); err != nil {
		t.Fatal(err)
	}
	again, err := db.ReportAttention(ctx, c, e)
	if err != nil || again.ViewedAt == nil || !again.CreatedAt.Equal(first.CreatedAt) {
		t.Fatal("duplicate lost view receipt or request identity")
	}
	conflict := e
	conflict.Title = "different"
	if _, err = db.ReportAttention(ctx, c, conflict); err == nil {
		t.Fatal("same version changed")
	}
	e.Version = 3
	e.Title = "New choice"
	if item, err := db.ReportAttention(ctx, c, e); err != nil || item.ViewedAt != nil {
		t.Fatal("changed prompt did not become unread", err)
	}
	if item, err := db.ReportAttention(ctx, c, session.AttentionEvent{ID: e.ID, Version: 2, Kind: e.Kind, State: "resolved", Title: e.Title}); err != nil || item.Version != 3 || item.State != "pending" {
		t.Fatal("old closure hid new prompt")
	}
	if _, err = db.MarkAttentionViewed(ctx, e.ID, 1, sample()); err == nil {
		t.Fatal("stale notification marked newer prompt viewed")
	}
	e.Version = 4
	e.State = "withdrawn"
	if _, err = db.ReportAttention(ctx, c, e); err != nil {
		t.Fatal(err)
	}
	e.Version = 5
	e.State = "pending"
	if _, err = db.ReportAttention(ctx, c, e); err == nil {
		t.Fatal("ended request resurrected")
	}
	e.ID = strings.Repeat("b", 32)
	e.Version = 2
	e.State = "resolved"
	if _, err = db.ReportAttention(ctx, c, e); err != nil {
		t.Fatal(err)
	}
	e.Version = 1
	e.State = "pending"
	if item, err := db.ReportAttention(ctx, c, e); err != nil || item.State != "resolved" {
		t.Fatal("resolve-before-create resurrected")
	}
}

func TestAttentionSourceScopeExitAndDisconnect(t *testing.T) {
	db, c, e := attentionFixture(t)
	ctx := context.Background()
	if _, err := db.ReportAttention(ctx, c, e); err != nil {
		t.Fatal(err)
	}
	v := sample()
	other := v
	other.ID = strings.Repeat("3", 32)
	other.RunID = strings.Repeat("4", 32)
	other.BackendID = "other"
	if err := db.Save(ctx, other); err != nil {
		t.Fatal(err)
	}
	otherC, err := session.NewAgentCredential(other.ID, other.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.BindAgent(ctx, otherC, other); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ReportAttention(ctx, otherC, e); err == nil {
		t.Fatal("request stolen by other Run")
	}
	if _, err := db.MarkAttentionViewed(ctx, e.ID, 1, other); err == nil {
		t.Fatal("activation followed another Run")
	}
	v.TerminalState = session.TerminalUnknown
	if err := db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	if rows, err := db.PendingAttention(ctx); err != nil || len(rows) != 1 {
		t.Fatal("disconnect resolved request")
	}
	if err := db.MarkStopped(ctx, v); err != nil {
		t.Fatal(err)
	}
	if rows, err := db.PendingAttention(ctx); err != nil || len(rows) != 0 {
		t.Fatal("known Run end retained interactive request")
	}
	if item, err := db.Attention(ctx, e.ID); err != nil || item.State != "invalidated" {
		t.Fatal("Run-end tombstone missing")
	}
	if _, err := db.MarkAttentionViewed(ctx, e.ID, 1, v); err == nil {
		t.Fatal("ended Run navigable")
	}
}

func TestAttentionBoundedCapacityNeverPrunesClosedIdentity(t *testing.T) {
	db, c, e := attentionFixture(t)
	ctx := context.Background()
	e.State = "resolved"
	for i := range session.MaxAttentionItems {
		e.ID = fmt.Sprintf("%032x", i+1)
		if _, err := db.ReportAttention(ctx, c, e); err != nil {
			t.Fatal(i, err)
		}
	}
	e.ID = strings.Repeat("e", 32)
	if _, err := db.ReportAttention(ctx, c, e); err == nil {
		t.Fatal("unbounded attention history")
	}
	e.ID = fmt.Sprintf("%032x", 1)
	e.Version = 2
	if _, err := db.ReportAttention(ctx, c, e); err != nil {
		t.Fatal("capacity refused update to retained identity")
	}
}

func TestAttentionV8MigrationAndScopeIsolation(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "db.sqlite3")
	db, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec(`DROP TABLE attention; PRAGMA user_version=8`); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	if check, err := store.Check(ctx, path); err != nil || check.SchemaVersion != 8 {
		t.Fatal("read-only v8 validation changed", err)
	}
	db, err = store.Open(ctx, path, "a")
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
	e := session.AttentionEvent{ID: strings.Repeat("a", 32), Version: 1, Kind: "approval", State: "pending", Title: "Allow?"}
	if _, err = db.ReportAttention(ctx, c, e); err != nil {
		t.Fatal(err)
	}
	other, err := store.Open(ctx, path, "b")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if items, err := other.PendingAttention(ctx); err != nil || len(items) != 0 {
		t.Fatal("cross-scope list")
	}
	if _, err = other.Attention(ctx, e.ID); err == nil {
		t.Fatal("cross-scope lookup")
	}
	if check, err := store.Check(ctx, path); err != nil || check.SchemaVersion != 9 {
		t.Fatal("schema9 invalid", err)
	}
}
