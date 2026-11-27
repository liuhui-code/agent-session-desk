package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/keybindings"
	"testing"
)

func TestKeyOverrideAndOldKey(t *testing.T) {
	h, s := attentionUIFixture(t)
	h.m.workspace.bindings = keybindings.Map{Overrides: map[string]string{"workspace.next": "alt+x"}}
	h.send(tea.KeyPressMsg{Code: 'j', Mod: tea.ModAlt})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if h.m.workspace.active != 0 {
		t.Fatal("old key retained")
	}
	s.mu.Lock()
	st := s.streams["first"]
	s.mu.Unlock()
	st.mu.Lock()
	n := len(st.writes)
	st.mu.Unlock()
	if n != 1 {
		t.Fatal("unbound key wasn't native", n)
	}
	h.send(tea.KeyPressMsg{Code: 'x', Mod: tea.ModAlt})
	h.until(func() bool { return h.m.workspace.active == 1 && h.m.workspace.conn != nil })
}
func TestUnboundKeyRespectsNavigation(t *testing.T) {
	h, s := attentionUIFixture(t)
	h.m.workspace.bindings = keybindings.Map{Overrides: map[string]string{"workspace.next": ""}}
	h.m.workspace.transition = workspaceHelp
	h.send(tea.KeyPressMsg{Code: 'j', Mod: tea.ModAlt})
	s.mu.Lock()
	st := s.streams["first"]
	s.mu.Unlock()
	st.mu.Lock()
	defer st.mu.Unlock()
	if len(st.writes) > 0 || h.m.workspace.pending > 0 {
		t.Fatal("key passed navigation fence")
	}
}
