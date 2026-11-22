package store_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func exited() session.Session {
	v := sample()
	v.TerminalState = session.TerminalExited
	v.PaneExit = &session.PaneExit{PaneID: "%7", Source: "psmux.pane_dead", ObservedAt: time.Now().UTC()}
	return v
}

func TestPaneExitObservationSurvivesReopenAndDoesNotClaimCurrentState(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	v := exited()
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.History(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].PaneExit == nil {
		t.Fatalf("exit observation lost: %#v", rows)
	}
	e := rows[0].PaneExit
	if e.PaneID != "%7" || e.Source != "psmux.pane_dead" || !e.ObservedAt.Equal(v.PaneExit.ObservedAt) || e.ExitCode != nil || rows[0].Session.TerminalState != session.TerminalUnknown || rows[0].Session.PaneExit != nil {
		t.Fatalf("observation became a stronger claim: %#v", rows)
	}
}

func TestRepeatedPaneExitPreservesFirstObservationWithoutWriterLock(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := exited()
	first := v.PaneExit.ObservedAt
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if _, err = raw.Exec("BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	defer raw.Exec("ROLLBACK")
	v.PaneExit.ObservedAt = first.Add(time.Second)
	if err = db.Save(ctx, v); err != nil {
		t.Fatalf("duplicate observation requested writer lock: %v", err)
	}
	rows, err := db.History(ctx, 100)
	if err != nil || !rows[0].PaneExit.ObservedAt.Equal(first) {
		t.Fatalf("observation time replaced: %#v %v", rows, err)
	}
}

func TestExitEvidenceCannotClaimMoreThanBackendProves(t *testing.T) {
	for _, kind := range []string{"live", "code", "source", "pane", "time"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "scope-a")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			v := exited()
			switch kind {
			case "live":
				v.TerminalState = session.TerminalAvailable
			case "code":
				code := 0
				v.PaneExit.ExitCode = &code
			case "source":
				v.PaneExit.Source = "unverified"
			case "pane":
				v.PaneExit.PaneID = "wrong"
			case "time":
				v.PaneExit.ObservedAt = time.Time{}
			}
			if err = db.Save(ctx, v); err == nil {
				t.Fatal("unproven exit evidence accepted")
			}
			rows, err := db.History(ctx, 100)
			if err != nil || len(rows) != 0 {
				t.Fatalf("partial record escaped failed save: %#v %v", rows, err)
			}
		})
	}
}
