package manager_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type outputTerminal struct {
	terminal
	captures   int
	capture    session.Preview
	captureErr error
}

func (f *outputTerminal) Preview(context.Context, session.Session, int) (session.Preview, error) {
	f.captures++
	return f.capture, f.captureErr
}

func TestExplicitOutputCaptureUsesBackendAndPersistenceWithoutControlEffects(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := observed()
	f := &outputTerminal{terminal: terminal{items: []session.Session{v}}, capture: session.Preview{Text: "captured original Run"}}
	m := manager.New(f, db)
	if _, err = m.List(ctx); err != nil {
		t.Fatal(err)
	}
	saved, err := m.RecordOutput(ctx, v)
	if err != nil || saved.RunID != v.RunID || saved.Text != f.capture.Text || saved.Complete {
		t.Fatalf("capture/persist boundary: %#v %v", saved, err)
	}
	// Offline reading must not call any terminal method, even with no Terminal.
	rows, err := manager.New(nil, db).Output(ctx, v.RunID)
	if err != nil || len(rows) != 1 || rows[0].ID != saved.ID || f.captures != 1 || f.starts != 0 || f.stops != 0 {
		t.Fatalf("output caused control effects: %#v %v", rows, err)
	}
}
