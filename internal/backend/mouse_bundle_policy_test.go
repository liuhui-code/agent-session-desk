package backend

import (
	"strings"
	"testing"
)

func TestSingleMousePipeIsRequestedOnlyForTheVerifiedBundle(t *testing.T) {
	t.Setenv("PSMUX_CONPTY_PIPE_MOUSE_ONLY", "inherited-untrusted")
	t.Setenv("PSMUX_CONPTY_DIR", "inherited-untrusted")
	plain := (&Backend{bin: "synthetic-backend", kind: "psmux"}).command()
	for _, entry := range plain.Env {
		if strings.HasPrefix(entry, "PSMUX_CONPTY_") {
			t.Fatal("default backend inherited explicit bundle transport")
		}
	}
	// bundle is private state produced only by prepareTerminalBundle validation.
	selected := (&Backend{bin: "synthetic-backend", kind: "psmux", bundle: t.TempDir()}).command()
	got := ""
	for _, entry := range selected.Env {
		if strings.HasPrefix(entry, "PSMUX_CONPTY_PIPE_MOUSE_ONLY=") {
			got = strings.TrimPrefix(entry, "PSMUX_CONPTY_PIPE_MOUSE_ONLY=")
		}
	}
	if got != "1" {
		t.Fatal("verified bundle did not select one native mouse channel")
	}
}
