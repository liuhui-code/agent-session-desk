package store_test

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestMessageOutboxPersistsAndDeduplicatesExactIntent(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	target := sample()
	if err = db.Save(ctx, target); err != nil {
		t.Fatal(err)
	}
	msg := session.Message{ID: strings.Repeat("a", 32), Target: target, Body: "literal $() ` ;\n\u4e2d\u6587", Fingerprint: strings.Repeat("b", 64)}
	first, err := db.QueueMessage(ctx, msg)
	if err != nil {
		t.Fatal(err)
	}
	again, err := db.QueueMessage(ctx, msg)
	if err != nil || again.ID != first.ID || again.State != "queued" {
		t.Fatalf("dedup: %+v %v", again, err)
	}
	msg.Fingerprint = strings.Repeat("c", 64)
	if _, err = db.QueueMessage(ctx, msg); err == nil {
		t.Fatal("conflicting idempotency key accepted")
	}
	db.Close()
	db, err = store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	got, err := db.Message(ctx, first.ID)
	if err != nil || got.Body != first.Body || got.State != "queued" || got.CreatedAt.IsZero() {
		t.Fatalf("reopen: %+v %v", got, err)
	}
	other, err := store.Open(ctx, path, "scope-b")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if _, err = other.Message(ctx, first.ID); err == nil {
		t.Fatal("cross-scope message exposed")
	}
}

func TestMessageClaimIsExclusiveAndNeverReplaysUncertainWrite(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	a, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	target := sample()
	if err = a.Save(ctx, target); err != nil {
		t.Fatal(err)
	}
	msg := session.Message{ID: strings.Repeat("a", 32), Target: target, Body: "draft", Fingerprint: strings.Repeat("b", 64)}
	if _, err = a.QueueMessage(ctx, msg); err != nil {
		t.Fatal(err)
	}
	wrong := target
	wrong.RunID = strings.Repeat("9", 32)
	if _, err = a.ClaimMessage(ctx, msg.ID, wrong); err == nil {
		t.Fatal("claim retargeted Run")
	}
	results := make(chan session.DeliveryAttempt, 2)
	var wg sync.WaitGroup
	for _, db := range []*store.Store{a, b} {
		wg.Add(1)
		go func(db *store.Store) {
			defer wg.Done()
			v, e := db.ClaimMessage(ctx, msg.ID, target)
			if e == nil {
				results <- v
			}
		}(db)
	}
	wg.Wait()
	close(results)
	var claims []session.DeliveryAttempt
	for v := range results {
		claims = append(claims, v)
	}
	if len(claims) != 1 {
		t.Fatalf("claim winners: %d", len(claims))
	}
	got, err := a.Message(ctx, msg.ID)
	if err != nil || got.State != "unknown" {
		t.Fatalf("claim must persist uncertainty before I/O: %+v %v", got, err)
	}
	if _, err = b.ClaimMessage(ctx, msg.ID, target); err == nil {
		t.Fatal("uncertain input replayed")
	}
	attempt := claims[0]
	bad := attempt
	bad.Token = strings.Repeat("0", 32)
	if _, err = a.FinishMessage(ctx, bad, true); err == nil {
		t.Fatal("foreign attempt wrote receipt")
	}
	written, err := a.FinishMessage(ctx, attempt, true)
	if err != nil || written.State != "input_written" {
		t.Fatalf("receipt: %+v %v", written, err)
	}
	if _, err = b.ClaimMessage(ctx, msg.ID, target); err == nil {
		t.Fatal("written input replayed")
	}
	if _, err = a.FinishMessage(ctx, attempt, true); err == nil {
		t.Fatal("receipt committed twice")
	}
}
