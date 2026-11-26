package store_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestClaimReplyCannotBypassOwnership(t *testing.T) {
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
	m := session.Message{ID: strings.Repeat("a", 32), Target: v, Body: "request", Fingerprint: strings.Repeat("b", 64)}
	if _, err = db.QueueMessage(ctx, m); err != nil {
		t.Fatal(err)
	}
	if _, err = db.RecordReply(ctx, c, session.ReplyRequest{ID: strings.Repeat("c", 32), ReplyTo: m.ID, Body: "unclaimed result"}); err == nil {
		t.Fatal("reply accepted without task ownership")
	}
	rows, err := db.Replies(ctx, m.ID)
	if err != nil || len(rows) != 0 {
		t.Fatal("unclaimed result became durable")
	}
}

func claimFixture(t *testing.T) (*store.Store, string, session.AgentCredential, session.Message) {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	v := sample()
	c := replyCredential(v)
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	if err = db.BindAgent(ctx, c, v); err != nil {
		t.Fatal(err)
	}
	m := session.Message{ID: strings.Repeat("a", 32), Target: v, Body: "task", Fingerprint: strings.Repeat("b", 64)}
	if _, err = db.QueueMessage(ctx, m); err != nil {
		t.Fatal(err)
	}
	return db, path, c, m
}

func newClaimKey(t *testing.T, msg string) session.TaskClaimCredential {
	t.Helper()
	id, err := session.NewID()
	if err != nil {
		t.Fatal(err)
	}
	s, err := session.NewAgentCredential(id, msg)
	if err != nil {
		t.Fatal(err)
	}
	return session.TaskClaimCredential{ID: id, MessageID: msg, Secret: s.Secret}
}

func TestClaimSixteenConsumersOnlyOneNewGrant(t *testing.T) {
	db, path, c, msg := claimFixture(t)
	ctx := context.Background()
	start := make(chan struct{})
	results := make(chan session.TaskClaimCredential, 16)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		other, err := store.Open(ctx, path, "a")
		if err != nil {
			t.Fatal(err)
		}
		defer other.Close()
		key := newClaimKey(t, msg.ID)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			got, err := other.ClaimTask(ctx, c, key, 300)
			if err == nil && got.NewlyClaimed && got.State == "active" {
				results <- key
			}
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	var winners []session.TaskClaimCredential
	for k := range results {
		winners = append(winners, k)
	}
	if len(winners) != 1 {
		t.Fatalf("new grants %d", len(winners))
	}
	key := winners[0]
	again, err := db.ClaimTask(ctx, c, key, 300)
	if err != nil || again.NewlyClaimed {
		t.Fatal("receipt recovery granted execution again")
	}
	if _, err = db.ClaimMessage(ctx, msg.ID, msg.Target); err == nil {
		t.Fatal("structured ownership allowed terminal delivery")
	}
	if rows, err := db.Inbox(ctx, c); err != nil || len(rows) != 0 {
		t.Fatal("claimed work remained in ordinary inbox")
	}
	bad := key
	bad.Secret = strings.Repeat("0", 64)
	if _, err = db.RenewTaskClaim(ctx, c, bad); err == nil {
		t.Fatal("public claim ID authorised renewal")
	}
	if _, err = db.RecordReply(ctx, c, session.ReplyRequest{ID: strings.Repeat("c", 32), ReplyTo: msg.ID, Body: "forged", Claim: bad}); err == nil {
		t.Fatal("public claim ID authorised result")
	}
	renewed, err := db.RenewTaskClaim(ctx, c, key)
	if err != nil || renewed.ExpiresAt.Before(again.ExpiresAt) || renewed.NewlyClaimed {
		t.Fatal("renewal regressed or granted new execution")
	}
	r, err := db.RecordReply(ctx, c, session.ReplyRequest{ID: strings.Repeat("c", 32), ReplyTo: msg.ID, Body: "result", Claim: key})
	if err != nil || r.ClaimID != key.ID {
		t.Fatalf("owned reply %+v %v", r, err)
	}
	status, err := db.TaskClaimStatus(ctx, msg.ID)
	if err != nil || status.State != "replied" || status.NewlyClaimed {
		t.Fatal("result did not atomically seal ownership")
	}
}

func TestClaimExpiredOwnerNeverReactivatesAfterReopen(t *testing.T) {
	db, path, c, msg := claimFixture(t)
	ctx := context.Background()
	key := newClaimKey(t, msg.ID)
	if _, err := db.ClaimTask(ctx, c, key, 300); err != nil {
		t.Fatal(err)
	}
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	// A controlled timestamp fault, not a change to the host clock.
	if _, err = raw.Exec(`UPDATE task_claims SET expires_ns=?`, time.Now().Add(-time.Hour).UnixNano()); err != nil {
		t.Fatal(err)
	}
	status, err := db.TaskClaimStatus(ctx, msg.ID)
	if err != nil || status.State != "unknown" || status.Reason != "lease_expired" {
		t.Fatalf("expiry %+v %v", status, err)
	}
	// Once materialized, even a later clock/record jump cannot revive it.
	if _, err = raw.Exec(`UPDATE task_claims SET expires_ns=?`, time.Now().Add(time.Hour).UnixNano()); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	db.Close()
	db, err = store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if status, err = db.ClaimTask(ctx, c, key, 300); err == nil || status.State != "unknown" || status.NewlyClaimed {
		t.Fatal("same-file recovery reactivated uncertain work")
	}
	if _, err = db.ClaimTask(ctx, c, newClaimKey(t, msg.ID), 300); err == nil {
		t.Fatal("new owner replaced uncertain work")
	}
	if _, err = db.RenewTaskClaim(ctx, c, key); err == nil {
		t.Fatal("unknown owner renewed")
	}
	if _, err = db.RecordReply(ctx, c, session.ReplyRequest{ID: strings.Repeat("c", 32), ReplyTo: msg.ID, Body: "late result", Claim: key}); err == nil {
		t.Fatal("expired worker published a result")
	}
	if rows, err := db.Replies(ctx, msg.ID); err != nil || len(rows) != 0 {
		t.Fatal("late result became durable")
	}
	if _, err = db.ClaimMessage(ctx, msg.ID, msg.Target); err == nil {
		t.Fatal("unknown ownership fell back to terminal delivery")
	}
}

func TestClaimTerminalRaceHasOneOwnershipPath(t *testing.T) {
	db, path, c, base := claimFixture(t)
	ctx := context.Background()
	other, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	for i := 0; i < 12; i++ {
		msg := base
		msg.ID = fmt.Sprintf("%032x", i+1)
		if _, err = db.QueueMessage(ctx, msg); err != nil {
			t.Fatal(err)
		}
		key := newClaimKey(t, msg.ID)
		start := make(chan struct{})
		results := make(chan bool, 2)
		go func() { <-start; _, err := db.ClaimTask(ctx, c, key, 300); results <- err == nil }()
		go func() { <-start; _, err := other.ClaimMessage(ctx, msg.ID, msg.Target); results <- err == nil }()
		close(start)
		wins := 0
		for j := 0; j < 2; j++ {
			if <-results {
				wins++
			}
		}
		if wins != 1 {
			t.Fatalf("ownership paths %d", wins)
		}
	}
}

func TestClaimObservedOwnerStopBecomesUnknown(t *testing.T) {
	db, _, c, msg := claimFixture(t)
	ctx := context.Background()
	key := newClaimKey(t, msg.ID)
	if _, err := db.ClaimTask(ctx, c, key, 300); err != nil {
		t.Fatal(err)
	}
	if err := db.MarkStopped(ctx, msg.Target); err != nil {
		t.Fatal(err)
	}
	status, err := db.TaskClaimStatus(ctx, msg.ID)
	if err != nil || status.State != "unknown" || status.Reason != "owner_exit_observed" {
		t.Fatal("stopped owner retained active claim")
	}
}

func claimForReply(t *testing.T, db *store.Store, c session.AgentCredential, id string) session.TaskClaimCredential {
	t.Helper()
	keyID, err := session.NewID()
	if err != nil {
		t.Fatal(err)
	}
	secret, err := session.NewAgentCredential(keyID, id)
	if err != nil {
		t.Fatal(err)
	}
	key := session.TaskClaimCredential{ID: keyID, MessageID: id, Secret: secret.Secret}
	if _, err = db.ClaimTask(context.Background(), c, key, session.DefaultClaimSeconds); err != nil {
		t.Fatal(err)
	}
	return key
}
