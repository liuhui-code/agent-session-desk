//go:build windows

package manager_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type attentionDisconnected struct{ manager.Terminal }

func (b attentionDisconnected) List(context.Context) ([]session.Session, error) {
	return nil, fmt.Errorf("injected transport unavailable")
}

// Uses real native processes/SQLite, with List failure injected only at the
// public boundary. Not a vendor-agent or Windows Terminal desktop test.
func TestWindowsNativeAttentionSimulatedFaults(t *testing.T) {
	for _, abrupt := range []bool{false, true} {
		t.Run(fmt.Sprint("abrupt=", abrupt), func(t *testing.T) { attentionFaultFlow(t, abrupt) })
	}
}
func attentionFaultFlow(t *testing.T, abrupt bool) {
	bdir := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE")
	fixture := os.Getenv("ASD_NATIVE_MESSAGING_FIXTURE")
	if bdir == "" || fixture == "" {
		t.Skip("explicit private native bundle required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	root := t.TempDir()
	id, _ := session.NewID()
	opts := backend.Options{TerminalBundle: bdir, RuntimeDir: filepath.Join(root, "r"), Namespace: "asd-af-" + id[:12]}
	b, err := backend.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	kind, loc, _ := backend.Location(opts)
	path := filepath.Join(root, "state", "metadata.sqlite3")
	db, err := store.Open(ctx, path, strings.ToLower(kind+":"+loc))
	if err != nil {
		t.Fatal(err)
	}
	m := manager.New(b, db)
	journal := filepath.Join(root, "fixture")
	v, err := m.Start(ctx, session.Launch{Name: "Attention faults", Executable: fixture, Dir: root, Env: []string{"ASD_FIXTURE_REPORT_PATH=" + journal, "ASD_FIXTURE_EXIT_CODE=73"}})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	defer func() {
		c, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if e := m.Stop(c, v); e != nil {
			t.Error(e)
		}
		db.Close()
	}()
	wait := func(check func() bool) {
		for !check() {
			if ctx.Err() != nil {
				t.Fatal("native fault deadline")
			}
			time.Sleep(30 * time.Millisecond)
		}
	}
	wait(func() bool { data, _ := os.ReadFile(journal); return bytes.Contains(data, []byte(`"event":"started"`)) })
	cred, _ := session.NewAgentCredential(v.ID, v.RunID)
	if _, err = m.BindAgent(ctx, cred); err != nil {
		t.Fatal(err)
	}
	e := session.AttentionEvent{ID: id, Version: 1, Kind: "question", State: "pending", Title: "Synthetic decision"}
	a, err := m.ReportAttention(ctx, cred, e)
	if err != nil {
		t.Fatal(err)
	}
	a, err = m.SnoozeAttention(ctx, a, 30*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 64; i++ {
		x, e2 := m.ReportAttention(ctx, cred, e)
		if e2 != nil || x.ReminderRevision != a.ReminderRevision || x.SnoozedUntil == nil {
			t.Fatal("duplicate reset reminder", e2)
		}
	}
	e.Version = 3
	e.Title = "Changed choice"
	a, err = m.ReportAttention(ctx, cred, e)
	if err != nil || a.SnoozedUntil != nil {
		t.Fatal("new prompt kept old snooze", err)
	}
	old := e
	old.Version = 2
	old.State = "resolved"
	a, err = m.ReportAttention(ctx, cred, old)
	if err != nil || a.State != "pending" || a.Version != 3 {
		t.Fatal("old resolve won", err)
	}
	offline := manager.New(attentionDisconnected{b}, db)
	if _, err = offline.OpenAttention(ctx, id, 3); err == nil {
		t.Fatal("unavailable target opened")
	}
	a, err = db.Attention(ctx, id)
	if err != nil || a.State != "pending" || a.ViewedAt != nil {
		t.Fatal("disconnect resolved/viewed", err)
	}
	if _, err = m.OpenAttention(ctx, id, 3); err != nil {
		t.Fatal("reconnect", err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, path, strings.ToLower(kind+":"+loc))
	if err != nil {
		t.Fatal(err)
	}
	m = manager.New(b, db)
	a, err = db.Attention(ctx, id)
	if err != nil || a.State != "pending" || a.ViewedAt == nil {
		t.Fatal("reopen lost state", err)
	}
	// Nonzero exit is explicit fixture failure, not an inferred agent status.
	if abrupt {
		data, _ := os.ReadFile(journal)
		var start struct{ PID int }
		if json.Unmarshal(bytes.Split(data, []byte{'\n'})[0], &start) != nil || start.PID <= 0 {
			t.Fatal("own process receipt missing")
		}
		process, e := os.FindProcess(start.PID)
		if e != nil {
			t.Fatal(e)
		}
		if e = process.Kill(); e != nil {
			t.Fatal(e)
		}
		process.Release()
	} else {
		c, err := b.OpenTerminal(ctx, v, 80, 24)
		if err != nil {
			t.Fatal(err)
		}
		if err = c.SendText(ctx, "exit"); err != nil {
			t.Fatal(err)
		}
		if err = c.SendKey(ctx, "Enter"); err != nil {
			t.Fatal(err)
		}
		c.Close()
	}
	if !abrupt {
		wait(func() bool { d, _ := os.ReadFile(journal); return bytes.Contains(d, []byte(`"code":73`)) })
	}
	wait(func() bool {
		if _, e := m.Workspace(ctx); e != nil {
			return false
		}
		x, e := db.Attention(ctx, id)
		return e == nil && x.State == "invalidated"
	})
	if _, err = m.OpenAttention(ctx, id, 3); err == nil {
		t.Fatal("ended prompt opened")
	}
	e.Version = 4
	if _, err = m.ReportAttention(ctx, cred, e); err == nil {
		t.Fatal("dead producer accepted")
	}
	data, _ := os.ReadFile(journal)
	count := 0
	for _, l := range bytes.Split(data, []byte{'\n'}) {
		var q struct{ Event, Text string }
		if json.Unmarshal(l, &q) == nil && q.Event == "input" {
			count++
			if q.Text != "exit" {
				t.Fatal("automatic answer")
			}
		}
	}
	want := 1
	if abrupt {
		want = 0
	}
	if count != want {
		t.Fatal("input count", count)
	}
	t.Log("64 duplicates; changed/old event; injected disconnect; reopen; native exit73; invalidation; no auto-answer")
}
