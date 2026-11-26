package manager_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestMessageSnapshotTransferCopiesRangeAndWritesOnlyAfterDurableClaim(t *testing.T) {
	ctx := context.Background()
	db, e := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	source := observed()
	target := source
	target.ID = strings.Repeat("3", 32)
	target.RunID = strings.Repeat("4", 32)
	target.BackendID = "asd_target"
	f := &outputTerminal{terminal: terminal{items: []session.Session{source, target}}, capture: session.Preview{Text: "one\n\u4e2d\u6587\nliteral $() ` ;"}}
	m := manager.New(f, db)
	if _, e = m.List(ctx); e != nil {
		t.Fatal(e)
	}
	read, e := m.Read(ctx, source, 3)
	if e != nil {
		t.Fatal(e)
	}
	request := session.MessageRequest{RequestID: strings.Repeat("a", 32), TargetSessionID: target.ID, TargetRunID: target.RunID, SourceRunID: source.RunID, SourceRecordID: read.ID, StartLine: 2, EndLine: 3}
	queued, e := m.QueueMessage(ctx, request)
	if e != nil || queued.Body != "\u4e2d\u6587\nliteral $() ` ;" || queued.Source.StartLine != 2 || queued.Source.Complete {
		t.Fatalf("queue %+v %v", queued, e)
	}
	// Rotate the source record out; the queued payload and provenance must survive.
	for i := 0; i < 3; i++ {
		if _, e = m.Read(ctx, source, 3); e != nil {
			t.Fatal(e)
		}
	}
	retry, e := m.QueueMessage(ctx, request)
	if e != nil || retry.Body != queued.Body {
		t.Fatalf("idempotent snapshot retry after rotation: %+v %v", retry, e)
	}
	writes := 0
	written, e := m.DeliverMessage(ctx, queued.ID, target, func(ctx context.Context, body string) error {
		writes++
		before, e := db.Message(ctx, queued.ID)
		if e != nil || before.State != "unknown" {
			t.Fatal("input before durable claim")
		}
		if body != queued.Body {
			t.Fatal("payload changed")
		}
		return nil
	})
	if e != nil || written.State != "input_written" || writes != 1 {
		t.Fatalf("receipt: %+v %v writes=%d", written, e, writes)
	}
	if _, e = m.DeliverMessage(ctx, queued.ID, target, func(context.Context, string) error { writes++; return nil }); e == nil || writes != 1 {
		t.Fatal("duplicate input replayed")
	}
	if f.starts != 0 || f.stops != 0 {
		t.Fatal("messaging changed process lifetime")
	}
}

func TestMessageRejectsNewRunAndKeepsUncertainInputUnknown(t *testing.T) {
	ctx := context.Background()
	db, e := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	v := observed()
	f := &terminal{items: []session.Session{v}}
	m := manager.New(f, db)
	req := session.MessageRequest{RequestID: strings.Repeat("a", 32), TargetSessionID: v.ID, TargetRunID: v.RunID, Body: "task"}
	q, e := m.QueueMessage(ctx, req)
	if e != nil {
		t.Fatal(e)
	}
	f.items[0].RunID = strings.Repeat("9", 32)
	writes := 0
	if _, e = m.DeliverMessage(ctx, q.ID, v, func(context.Context, string) error { writes++; return nil }); e == nil || writes != 0 {
		t.Fatal("old request reached new Run")
	}
	got, _ := db.Message(ctx, q.ID)
	if got.State != "queued" {
		t.Fatal("preflight failure consumed request")
	}
	f.items[0] = v
	unknown, e := m.DeliverMessage(ctx, q.ID, v, func(context.Context, string) error { writes++; return errors.New("partial transport write") })
	if e == nil || unknown.State != "unknown" || writes != 1 {
		t.Fatalf("uncertainty lost: %+v %v", unknown, e)
	}
	if _, e = m.DeliverMessage(ctx, q.ID, v, func(context.Context, string) error { writes++; return nil }); e == nil || writes != 1 {
		t.Fatal("uncertain input replayed")
	}
}

func TestMessageWrittenBeforeReceiptCancellationStaysUnknown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, e := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	v := observed()
	m := manager.New(&terminal{items: []session.Session{v}}, db)
	q, e := m.QueueMessage(ctx, session.MessageRequest{RequestID: strings.Repeat("a", 32), TargetSessionID: v.ID, TargetRunID: v.RunID, Body: "one write"})
	if e != nil {
		t.Fatal(e)
	}
	writes := 0
	got, e := m.DeliverMessage(ctx, q.ID, v, func(context.Context, string) error { writes++; cancel(); return nil })
	if e == nil || got.State != "unknown" || writes != 1 {
		t.Fatalf("receipt falsely confirmed: %+v %v", got, e)
	}
	durable, e := db.Message(context.Background(), q.ID)
	if e != nil || durable.State != "unknown" {
		t.Fatalf("durable uncertainty: %+v %v", durable, e)
	}
}
