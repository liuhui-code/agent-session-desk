package ui

import "testing"

func TestCompanionWorkbenchClickRestoresWindowWithoutChangingNavigation(t *testing.T) {
	h, s := attentionUIFixture(t)
	h.send(sessionsMsg{sessions: sampleSessions()})
	calls := 0
	h.m.attentionFocus = func() bool { calls++; return true }
	active, focused := h.m.workspace.active, h.m.workspace.focused
	h.m.help = true
	h.send(WorkbenchActivate{})
	if calls != 1 || !h.m.help || h.m.workspace.active != active || h.m.workspace.focused != focused || h.m.attention.busy {
		t.Fatal("window-only click changed current view or missed foreground request")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.visits != 0 {
		t.Fatal("idle window restore opened an attention item")
	}
	for _, stream := range s.streams {
		stream.mu.Lock()
		n := len(stream.writes)
		stream.mu.Unlock()
		if n != 0 {
			t.Fatal("window restore wrote terminal input")
		}
	}
}

func TestCompanionWorkbenchRestoreFailureIsVisibleWithoutSwitchingRun(t *testing.T) {
	h, _ := attentionUIFixture(t)
	active := h.m.workspace.active
	h.m.attentionFocus = func() bool { return false }
	h.send(WorkbenchActivate{})
	if h.m.status != "Bring the original workbench window to the foreground" || h.m.workspace.active != active {
		t.Fatal("failed foreground request silently succeeded or switched Run")
	}
}
