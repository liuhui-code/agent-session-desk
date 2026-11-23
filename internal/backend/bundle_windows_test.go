//go:build windows

package backend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsTerminalBundleRejectsSystemHostedExistingRun(t *testing.T) {
	dir := os.Getenv("ASD_TERMINAL_BUNDLE_TEST_DIR")
	if dir == "" {
		t.Skip("requires an explicitly prepared pinned terminal bundle")
	}
	h := newNativeHarness(t)
	// Use the very same image in both configurations. A refusal must be about
	// the actual host, not merely about two copies of a backend executable.
	h.opts.Binary = filepath.Join(dir, "psmux.exe")
	var err error
	h.b, err = New(h.opts)
	if err != nil {
		t.Fatal(err)
	}
	launch, report := h.launch("System host must not masquerade as selected bundle")
	s := h.start(launch)
	fixtureEvent(t, h.ctx, report, "started", "")
	opts := h.opts
	opts.Binary, opts.TerminalBundle = "", dir
	b, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}
	live, err := b.OpenTerminal(h.ctx, s, 80, 20)
	if live != nil {
		live.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "terminal bundle") {
		t.Fatalf("must refuse existing system-hosted Run without replacing or sending input: %v", err)
	}
	// Rejection must leave the original Run discoverable for explicit stop.
	rows, err := h.fresh().List(h.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := findNativeSession(rows, s); err != nil {
		t.Fatal(err)
	}
}

func TestWindowsTerminalBundleRejectsUnpinnedFilesBeforeRuntimeCreation(t *testing.T) {
	dir := t.TempDir()
	bundle := filepath.Join(dir, "bundle")
	if err := os.Mkdir(bundle, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"psmux.exe", "conpty.dll", "OpenConsole.exe"} {
		if err := os.WriteFile(filepath.Join(bundle, name), []byte("not the pinned artifact"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	runtimeDir := filepath.Join(dir, "runtime")
	_, err := New(Options{TerminalBundle: bundle, RuntimeDir: runtimeDir})
	if err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("unpinned bundle not rejected: %v", err)
	}
	if _, err := os.Stat(runtimeDir); !os.IsNotExist(err) {
		t.Fatalf("invalid bundle created runtime: %v", err)
	}
}
