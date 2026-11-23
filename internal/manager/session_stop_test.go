package manager_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type reviewedStopper interface {
	StopReviewed(context.Context, session.Session) error
}

func TestReviewedStopRejectsChangedTargetBeforeTerminalEffect(t *testing.T) {
	for _, change := range []string{"same", "missing", "run", "backend", "command", "name", "unknown", "duplicate", "exited"} {
		t.Run(change, func(t *testing.T) {
			ctx := context.Background()
			db, err := store.Open(ctx, filepath.Join(t.TempDir(), "stop.sqlite3"), "scope")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			f := &terminal{items: []session.Session{observed()}}
			m := manager.New(f, db)
			view, err := m.Workspace(ctx)
			if err != nil {
				t.Fatal(err)
			}
			expected := view.Sessions[0]
			switch change {
			case "missing":
				f.items = nil
			case "run":
				f.items[0].RunID = "33333333333333333333333333333333"
			case "backend":
				f.items[0].BackendID = "replacement"
			case "command":
				f.items[0].Command = "replacement.exe"
			case "name":
				if _, err = db.Rename(ctx, expected, "changed"); err != nil {
					t.Fatal(err)
				}
			case "unknown":
				f.items[0].TerminalState = session.TerminalUnknown
			case "duplicate":
				f.items = append(f.items, f.items[0])
			case "exited":
				f.items[0].TerminalState = session.TerminalExited
			}
			stopper, ok := any(m).(reviewedStopper)
			if !ok {
				t.Fatal("manager has no reviewed Session stop")
			}
			err = stopper.StopReviewed(ctx, expected)
			accepted := change == "same" || change == "exited"
			if accepted != (err == nil) || f.starts != 0 || (accepted && f.stops != 1) || (!accepted && f.stops != 0) {
				t.Fatalf("stop effects: starts=%d stops=%d err=%v", f.starts, f.stops, err)
			}
			if accepted {
				rows, e := db.History(ctx, 100)
				if e != nil || len(rows) != 1 || rows[0].StopConfirmedAt == nil || rows[0].Session.RunID != expected.RunID {
					t.Fatalf("stop lost original history: %+v %v", rows, e)
				}
			}
		})
	}
}

func TestReviewedStopNameABAMustBeReviewedAgain(t *testing.T) {
	ctx := context.Background()
	db, e := store.Open(ctx, filepath.Join(t.TempDir(), "aba.sqlite3"), "scope")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	f := &terminal{items: []session.Session{observed()}}
	m := manager.New(f, db)
	view, e := m.Workspace(ctx)
	if e != nil {
		t.Fatal(e)
	}
	original := view.Sessions[0]
	changed, e := db.Rename(ctx, original, "temporary")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.Rename(ctx, changed, original.Title); e != nil {
		t.Fatal(e)
	}
	if e = m.StopReviewed(ctx, original); e == nil || f.stops != 0 {
		t.Fatal("name ABA was accepted")
	}
}

type stopHookTerminal struct {
	terminal
	afterStop func()
}

func (f *stopHookTerminal) Stop(ctx context.Context, v session.Session) error {
	f.stops++
	if f.afterStop != nil {
		f.afterStop()
	}
	return f.err
}

func TestReviewedStopSeparatesBackendEffectFromMetadataFailure(t *testing.T) {
	for _, failure := range []string{"discovery", "backend", "metadata-after-stop"} {
		t.Run(failure, func(t *testing.T) {
			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "partial.sqlite3")
			db, e := store.Open(ctx, path, "scope")
			if e != nil {
				t.Fatal(e)
			}
			defer db.Close()
			f := &stopHookTerminal{terminal: terminal{items: []session.Session{observed()}}}
			m := manager.New(f, db)
			view, e := m.Workspace(ctx)
			if e != nil {
				t.Fatal(e)
			}
			switch failure {
			case "discovery":
				db.Close()
			case "backend":
				f.afterStop = func() { f.err = fmt.Errorf("uncertain backend result") }
			case "metadata-after-stop":
				f.afterStop = func() { db.Close() }
			}
			e = m.StopReviewed(ctx, view.Sessions[0])
			if e == nil || f.starts != 0 || (failure == "discovery" && f.stops != 0) || (failure != "discovery" && f.stops != 1) {
				t.Fatalf("effect/result mismatch: stops%d %v", f.stops, e)
			}
			if failure == "metadata-after-stop" && !strings.Contains(e.Error(), "backend session stopped") {
				t.Fatal("known stop effect hidden")
			}
			db.Close()
			db, e = store.Open(ctx, path, "scope")
			if e != nil {
				t.Fatal(e)
			}
			defer db.Close()
			rows, e := db.History(ctx, 100)
			if e != nil || len(rows) != 1 || rows[0].StopConfirmedAt != nil {
				t.Fatal("failed stop invented durable confirmation")
			}
		})
	}
}
