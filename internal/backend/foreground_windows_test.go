//go:build windows

package backend

import (
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Renamed fixtures test process observation, not vendor-agent compatibility.
func TestWindowsNativeForegroundProbe(t *testing.T) {
	h := newNativeHarness(t)
	shell := os.Getenv("ASD_SHELL_TEST_BINARY")
	if shell == "" {
		t.Fatal("explicit shell required")
	}
	data, e := os.ReadFile(h.fixture)
	if e != nil {
		t.Fatal(e)
	}
	s := h.start(session.Launch{Name: "Foreground probe", Executable: shell, Args: []string{"-NoLogo", "-NoProfile"}, Dir: h.dir})
	query := func() session.Session {
		rows, e := h.b.List(h.ctx)
		if e != nil {
			t.Fatal(e)
		}
		for _, v := range rows {
			if v.ID == s.ID && v.RunID == s.RunID {
				return v
			}
		}
		t.Fatal("original Run missing")
		return session.Session{}
	}
	wait := func(want string) {
		until := time.Now().Add(8 * time.Second)
		for {
			v := query()
			if v.Command != s.Command {
				t.Fatal("launch changed")
			}
			if v.CurrentProgram == want && v.TerminalState == session.TerminalAvailable {
				return
			}
			if time.Now().After(until) {
				t.Fatalf("current=%q want=%q", v.CurrentProgram, want)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	wait("")
	for _, name := range []string{"codex", "claude", "opencode"} {
		child := filepath.Join(h.dir, name+".exe")
		if e = os.WriteFile(child, data, 0700); e != nil {
			t.Fatal(e)
		}
		h.sendLine(s, "& '"+strings.ReplaceAll(child, "'", "''")+"'")
		wait(name)
		h.sendLine(s, "exit")
		wait("")
	}
	t.Log("three synthetic child names under original PowerShell; start/exit observed, launch unchanged")
}
