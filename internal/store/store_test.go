package store_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func sample() session.Session {
	return session.Session{ID: "11111111111111111111111111111111", RunID: "22222222222222222222222222222222", BackendID: "asd_first", Title: "开发", Command: "agent.exe", TerminalState: session.TerminalAvailable}
}

func TestMetadataSurvivesReopenWithoutClaimingLiveness(t *testing.T) {
	path := filepath.Join(t.TempDir(), "中文 #%", "metadata.sqlite3")
	ctx := context.Background()
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.History(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Session.ID != sample().ID || rows[0].Session.RunID != sample().RunID || rows[0].Session.Title != "开发" || rows[0].Session.TerminalState != session.TerminalUnknown || rows[0].RecordedAt.IsZero() {
		t.Fatalf("metadata changed or invented liveness: %#v", rows)
	}
}

func TestScopeAndIdentityConflictsCannotOverwriteOrLeaveHalfRecords(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	a, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := store.Open(ctx, path, "scope-b")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	v := sample()
	if err = a.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	different := v
	different.ID = "33333333333333333333333333333333"
	different.Title = "different"
	if err = a.Save(ctx, different); err == nil {
		t.Fatal("Run identity rebound to a different Session")
	}
	if err = b.Save(ctx, different); err != nil {
		t.Fatal(err)
	}
	got, err := a.History(ctx, 10)
	if err != nil || len(got) != 1 || got[0].Session.ID != v.ID || got[0].Session.Title != v.Title {
		t.Fatalf("scope/transaction changed: %#v %v", got, err)
	}
	got, err = b.History(ctx, 10)
	if err != nil || len(got) != 1 || got[0].Session.ID != different.ID {
		t.Fatalf("scope isolation: %#v %v", got, err)
	}
	// A conflicting Save must not leave a Session row behind either.
	different.RunID = "44444444444444444444444444444444"
	different.Title = "legitimate"
	if err = a.Save(ctx, different); err != nil {
		t.Fatalf("failed transaction left a parent behind: %v", err)
	}
}

func TestConfirmedStopKeepsHistoryAndChecksExactRun(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := sample()
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	wrong := v
	wrong.RunID = "33333333333333333333333333333333"
	if err = db.MarkStopped(ctx, wrong); err == nil {
		t.Fatal("stopped unrelated Run")
	}
	if err = db.MarkStopped(ctx, v); err != nil {
		t.Fatal(err)
	}
	first, err := db.History(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	if err = db.MarkStopped(ctx, v); err != nil {
		t.Fatal(err)
	}
	rows, err := db.History(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].StopConfirmedAt == nil || !rows[0].StopConfirmedAt.Equal(*first[0].StopConfirmedAt) || rows[0].Session.TerminalState != session.TerminalUnknown {
		t.Fatalf("history stop evidence changed: %#v", rows)
	}
}

func TestHistoryReadIsBounded(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, limit := range []int{-1, 0, 1001} {
		if _, err = db.History(ctx, limit); err == nil {
			t.Fatalf("unbounded limit accepted: %d", limit)
		}
	}
}

func TestInvalidIdentityCannotBecomePersistentMetadata(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, kind := range []string{"empty-session", "bad-run", "empty-backend", "oversized-title"} {
		v := sample()
		switch kind {
		case "empty-session":
			v.ID = ""
		case "bad-run":
			v.RunID = "xyz"
		case "empty-backend":
			v.BackendID = ""
		case "oversized-title":
			v.Title = strings.Repeat("x", 4097)
		}
		if err = db.Save(ctx, v); err == nil {
			t.Fatalf("invalid metadata accepted: %s", kind)
		}
	}
	rows, err := db.History(ctx, 100)
	if err != nil || len(rows) != 0 {
		t.Fatalf("invalid records written: %#v %v", rows, err)
	}
}
