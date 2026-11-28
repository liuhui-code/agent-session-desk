package backend

import (
	"context"
	"strings"
	"testing"
)

func TestDiscoveryWaitsForNewSessionPublicationWithoutRetry(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	var published string
	editDiscoveryState(t, func(state *fakeState) {
		published = state.Metadata
		state.Metadata = ""
	})
	rows, err := b.List(context.Background())
	if err != nil || len(rows) != 0 {
		t.Fatalf("unpublished session must not fail or appear live: %+v %v", rows, err)
	}
	editDiscoveryState(t, func(state *fakeState) { state.Metadata = published })
	rows, err = b.List(context.Background())
	if err != nil || len(rows) != 1 || rows[0].ID != s.ID || rows[0].RunID != s.RunID {
		t.Fatalf("published original identity missing: %+v %v", rows, err)
	}
	// The first discovery made one metadata read, never an error retry.
	reads := 0
	for _, args := range commandLog(t) {
		if strings.Contains(strings.Join(args, " "), "show-options") {
			reads++
		}
	}
	if reads != 2 {
		t.Fatalf("unexpected discovery read/retry count: %d", reads)
	}
}

func TestDiscoveryStillRejectsNonemptyCorruptMetadata(t *testing.T) {
	for _, value := range []string{" ", "@@@@", "e30"} {
		t.Run(value, func(t *testing.T) {
			b := newFake(t)
			startFake(t, b)
			editDiscoveryState(t, func(state *fakeState) { state.Metadata = value })
			if _, err := b.List(context.Background()); err == nil {
				t.Fatal("nonempty corrupted marker was hidden as construction")
			}
		})
	}
}

func TestPreviouslyPublishedMissingMarkerRemainsUncertainAndUnusable(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	if _, err := b.List(context.Background()); err != nil {
		t.Fatal(err)
	}
	editDiscoveryState(t, func(state *fakeState) { state.Metadata = "" })
	before := len(commandLog(t))
	if _, err := b.List(context.Background()); err == nil {
		t.Fatal("known original with a missing marker reported certain")
	}
	if _, err := b.Preview(context.Background(), s, 20); err == nil {
		t.Fatal("unpublished target permitted capture")
	}
	if err := b.Stop(context.Background(), s); err == nil {
		t.Fatal("unpublished target permitted stop")
	}
	for _, args := range commandLog(t)[before:] {
		for _, arg := range args {
			if arg == "capture-pane" || arg == "kill-session" || arg == "send-keys" {
				t.Fatalf("uncertain target caused side effect: %q", args)
			}
		}
	}
}
