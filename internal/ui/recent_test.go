package ui

import (
	tea "charm.land/bubbletea/v2"
	"testing"
)

func TestPreviousSession(t *testing.T) {
	h, _ := attentionUIFixture(t)
	h.m.rememberWorkspace()
	h.send(tea.KeyPressMsg{Code: 'j', Mod: tea.ModAlt})
	h.until(func() bool { return h.m.workspace.active == 1 && h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: 'p', Mod: tea.ModAlt})
	h.until(func() bool { return h.m.workspace.active == 0 && h.m.workspace.conn != nil })
}
