package manager_test

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type terminal struct {
	items         []session.Session
	err           error
	starts, stops int
	afterStart    func()
}

func (f *terminal) Start(context.Context, session.Launch) (session.Session, error) {
	f.starts++
	if f.afterStart != nil {
		f.afterStart()
	}
	return f.items[0], nil
}
func (f *terminal) List(context.Context) ([]session.Session, error) { return f.items, f.err }
func (f *terminal) Stop(context.Context, session.Session) error     { f.stops++; return f.err }
func (f *terminal) Preview(context.Context, session.Session, int) (session.Preview, error) {
	return session.Preview{}, nil
}
func (f *terminal) Attach(context.Context, session.Session) (*exec.Cmd, error) {
	return nil, errors.New("not used")
}
func observed() session.Session {
	return session.Session{ID: "11111111111111111111111111111111", RunID: "22222222222222222222222222222222", BackendID: "asd_observed", Title: "开发", Command: "agent.exe", TerminalState: session.TerminalAvailable}
}

func TestStartPersistsBeforeReportingSuccess(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	f := &terminal{items: []session.Session{observed()}}
	got, err := manager.New(f, db).Start(ctx, session.Launch{})
	if err != nil || got.ID != observed().ID {
		t.Fatalf("start: %#v %v", got, err)
	}
	db.Close()
	db, err = store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.History(ctx, 100)
	if err != nil || len(rows) != 1 || rows[0].Session.RunID != got.RunID {
		t.Fatalf("missing persisted run: %#v %v", rows, err)
	}
}

func TestMetadataFailureBeforeAndAfterExternalStartIsExplicit(t *testing.T) {
	for _, after := range []bool{false, true} {
		t.Run(fmt.Sprint(after), func(t *testing.T) {
			ctx := context.Background()
			db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			f := &terminal{items: []session.Session{observed()}}
			if after {
				f.afterStart = func() { db.Close() }
			} else {
				db.Close()
			}
			_, err = manager.New(f, db).Start(ctx, session.Launch{})
			if err == nil {
				t.Fatal("reported durable success")
			}
			if !after && f.starts != 0 {
				t.Fatal("created terminal with unavailable metadata")
			}
			if after && (!strings.Contains(err.Error(), observed().ID) || !strings.Contains(err.Error(), "do not retry start") || f.stops != 0) {
				t.Fatalf("ambiguous creation hidden or process killed: %v", err)
			}
		})
	}
}

func TestDiscoveryPersistsWithoutResolvingDisconnectOrMissingRun(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := &terminal{items: []session.Session{observed()}}
	m := manager.New(f, db)
	if _, err = m.List(ctx); err != nil {
		t.Fatal(err)
	}
	f.err = errors.New("disconnected")
	if _, err = m.List(ctx); err == nil {
		t.Fatal("disconnect masked")
	}
	f.err = nil
	f.items = nil
	if rows, err := m.List(ctx); err != nil || len(rows) != 0 {
		t.Fatalf("historical row became live: %#v %v", rows, err)
	}
	rows, err := db.History(ctx, 100)
	if err != nil || len(rows) != 1 || rows[0].Session.TerminalState != session.TerminalUnknown || rows[0].StopConfirmedAt != nil {
		t.Fatalf("disconnect/missing invented exit: %#v %v", rows, err)
	}
	if err = m.Stop(ctx, observed()); err != nil {
		t.Fatal(err)
	}
	rows, err = db.History(ctx, 100)
	if err != nil || rows[0].StopConfirmedAt == nil {
		t.Fatalf("explicit stop not recorded: %#v %v", rows, err)
	}
}

func TestPaneExitObservationSurvivesMissingBackendWithoutClaimingAgentOutcome(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	f := &terminal{items: []session.Session{observed()}}
	m := manager.New(f, db)
	if _, err = m.List(ctx); err != nil {
		t.Fatal(err)
	}
	f.items[0].TerminalState = session.TerminalExited
	f.items[0].PaneExit = &session.PaneExit{PaneID: "%7", Source: "psmux.pane_dead", ObservedAt: time.Now().UTC()}
	if _, err = m.List(ctx); err != nil {
		t.Fatal(err)
	}
	f.err = errors.New("disconnected")
	if _, err = m.List(ctx); err == nil {
		t.Fatal("disconnect masked")
	}
	f.err = nil
	f.items = nil
	if _, err = m.List(ctx); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.History(ctx, 100)
	if err != nil || len(rows) != 1 {
		t.Fatalf("history lost: %#v %v", rows, err)
	}
	if rows[0].PaneExit == nil || rows[0].PaneExit.ExitCode != nil || rows[0].Session.TerminalState != session.TerminalUnknown || rows[0].StopConfirmedAt != nil || f.starts != 0 {
		t.Fatalf("outcome exaggerated: %#v", rows)
	}
}
