package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestTerminalHandoffDoesNotWaitForHostWindowLifetime(t *testing.T) {
	if path := os.Getenv("ASD_TERMINAL_HANDOFF_FIXTURE"); path != "" {
		time.Sleep(800 * time.Millisecond)
		_ = os.WriteFile(path, []byte("done"), 0600)
		return
	}
	exe, _ := os.Executable()
	done := filepath.Join(t.TempDir(), "host-ended")
	cmd := exec.Command(exe, "-test.run=TestTerminalHandoffDoesNotWaitForHostWindowLifetime")
	cmd.Env = append(os.Environ(), "ASD_TERMINAL_HANDOFF_FIXTURE="+done)
	before := time.Now()
	err := startTerminalHandoff(cmd)
	elapsed := time.Since(before)
	// The owned fixture exits by itself; retain no resident process or handle.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, e := os.Stat(done); e == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if _, e := os.Stat(done); e != nil {
		t.Fatal("owned host fixture did not end")
	}
	if err != nil {
		t.Fatal(err)
	}
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("handoff waited for host lifetime: %v", elapsed)
	}
}
