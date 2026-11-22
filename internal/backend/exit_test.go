package backend

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestExitedOriginalPaneCarriesBoundedObservationNotExitCode(t *testing.T) {
	b := newFake(t)
	ctx := context.Background()
	created, err := b.Start(ctx, session.Launch{Executable: b.bin, Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if created.PaneExit != nil {
		t.Fatal("live terminal manufactured exit evidence")
	}
	path := os.Getenv("ASD_TEST_BACKEND_STATE")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var state fakeState
	if err = json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	state.Dead = "1"
	data, _ = json.Marshal(state)
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	before := time.Now().UTC()
	rows, err := b.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].TerminalState != session.TerminalExited {
		t.Fatalf("pane exit missing: %#v", rows)
	}
	e := rows[0].PaneExit
	if e == nil || e.PaneID != "%7" || e.Source != "psmux.pane_dead" || e.ExitCode != nil || e.ObservedAt.Before(before) {
		t.Fatalf("unreliable exit evidence: %#v", e)
	}
}

func TestMalformedOrDifferentPaneDoesNotBecomeExitEvidence(t *testing.T) {
	for _, tuple := range []string{"%9|1", "%7|bad", "%7|1\n%7|0"} {
		t.Run(tuple, func(t *testing.T) {
			b := newFake(t)
			ctx := context.Background()
			if _, err := b.Start(ctx, session.Launch{Executable: b.bin, Dir: t.TempDir()}); err != nil {
				t.Fatal(err)
			}
			t.Setenv("ASD_TEST_PANE_TUPLE", tuple)
			rows, err := b.List(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 || rows[0].TerminalState != session.TerminalUnknown || rows[0].PaneExit != nil {
				t.Fatalf("wrong pane evidence: %#v", rows)
			}
		})
	}
}
