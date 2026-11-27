package backend

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func editDiscoveryState(t *testing.T, edit func(*fakeState)) {
	t.Helper()
	p := os.Getenv("ASD_TEST_BACKEND_STATE")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var state fakeState
	if err = json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	edit(&state)
	data, err = json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(p, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoveryRevalidatesOmittedOriginal(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	ctx := context.Background()
	if _, err := b.List(ctx); err != nil {
		t.Fatal(err)
	}
	editDiscoveryState(t, func(s *fakeState) { s.OmitList = true })
	for range 3 {
		rows, err := b.List(ctx)
		if err != nil || len(rows) != 1 || rows[0].ID != s.ID || rows[0].RunID != s.RunID || rows[0].TerminalState != s.TerminalState {
			t.Fatalf("omitted original lost: rows=%+v err=%v", rows, err)
		}
	}
	// Same backend name with changed pane metadata cannot inherit the old identity.
	editDiscoveryState(t, func(s *fakeState) {
		r, err := decodeRecord(s.Name, s.Metadata)
		if err != nil {
			t.Fatal(err)
		}
		r.PaneID = "%8"
		s.Metadata, err = encodeRecord(r)
		if err != nil {
			t.Fatal(err)
		}
	})
	if rows, err := b.List(ctx); err == nil || len(rows) != 0 {
		t.Fatalf("changed target accepted: %+v %v", rows, err)
	}
	if rows, err := b.List(ctx); err != nil || len(rows) != 0 {
		t.Fatalf("failed hint retained: %+v %v", rows, err)
	}
}

func TestDiscoveryStopForgetsOriginal(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	ctx := context.Background()
	if _, err := b.List(ctx); err != nil {
		t.Fatal(err)
	}
	if err := b.Stop(ctx, s); err != nil {
		t.Fatal(err)
	}
	if rows, err := b.List(ctx); err != nil || len(rows) != 0 {
		t.Fatalf("stopped target retained: %+v %v", rows, err)
	}
}

func TestDiscoveryMissingTargetDoesNotStayCached(t *testing.T) {
	b := newFake(t)
	startFake(t, b)
	ctx := context.Background()
	if _, err := b.List(ctx); err != nil {
		t.Fatal(err)
	}
	editDiscoveryState(t, func(s *fakeState) { s.OmitList = true; s.Name = "" })
	if _, err := b.List(ctx); err == nil {
		t.Fatal("unverifiable original reported as known")
	}
	if rows, err := b.List(ctx); err != nil || len(rows) != 0 {
		t.Fatalf("missing hint retained: %+v %v", rows, err)
	}
}

func TestDiscoveryOmittedPaneStateIsObservedAgain(t *testing.T) {
	b := newFake(t)
	startFake(t, b)
	ctx := context.Background()
	if _, err := b.List(ctx); err != nil {
		t.Fatal(err)
	}
	editDiscoveryState(t, func(s *fakeState) { s.OmitList = true; s.Dead = "1" })
	rows, err := b.List(ctx)
	if err != nil || len(rows) != 1 || rows[0].TerminalState != "exited" {
		t.Fatalf("cached liveness reused: %+v %v", rows, err)
	}
}
