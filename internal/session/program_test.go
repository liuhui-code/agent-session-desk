package session

import "testing"

func TestProgramLabel(t *testing.T) {
	for _, s := range []string{"codex.exe", "claude.exe", "opencode.exe"} {
		if ProgramLabel(s) == "" {
			t.Fatal(s)
		}
	}
	for _, s := range []string{"pwsh.exe", "powershell.exe", "cmd.exe", "node.exe", "codex --token secret", "codex\x1b]0;bad\a"} {
		if ProgramLabel(s) != "" {
			t.Fatal("unusable display", s)
		}
	}
}
