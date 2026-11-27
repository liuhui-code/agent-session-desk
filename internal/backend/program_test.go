package backend

import (
	"context"
	"testing"
)

func TestForegroundDoesNotReplaceLaunch(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	t.Setenv("ASD_TEST_PANE_PROGRAM", "codex.exe")
	rows, e := b.List(context.Background())
	if e != nil || len(rows) != 1 || rows[0].CurrentProgram != "codex" || rows[0].Command != s.Command {
		t.Fatal("foreground/launch conflated", e)
	}
	t.Setenv("ASD_TEST_PANE_PROGRAM", "pwsh.exe")
	rows, e = b.List(context.Background())
	if e != nil || rows[0].CurrentProgram != "" {
		t.Fatal("shell label", e)
	}
}
