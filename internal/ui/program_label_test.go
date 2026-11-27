package ui

import (
	"strings"
	"testing"
)

func TestSidebarUsesCurrentProgramNotLauncher(t *testing.T) {
	h, _ := attentionUIFixture(t)
	h.m.sessions[0].Command = `C:\Program Files\PowerShell\pwsh.exe`
	h.m.sessions[0].CurrentProgram = "codex"
	v := strings.Join(h.m.workspaceSidebarCard(0, 40, 2, 1), "")
	if strings.Contains(v, "pwsh.exe") || !strings.Contains(v, "codex") {
		t.Fatal("launcher shown")
	}
}
