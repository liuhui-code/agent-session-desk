package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestF2RestoresOriginFocusAndSidebarAfterClosing(t *testing.T) {
	for _, origin := range []string{"terminal", "hidden-terminal", "management"} {
		for _, exit := range []string{"cancel", "save"} {
			t.Run(origin+"/"+exit, func(t *testing.T) {
				s := newRenameFixture()
				h := newWorkspaceHarness(t, s)
				h.until(func() bool { return h.m.workspace.conn != nil })
				if origin == "management" {
					h.send(tea.KeyPressMsg{Code: tea.KeyF6})
				}
				if origin == "hidden-terminal" {
					h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
				}
				focused, hidden := h.m.workspace.focused, h.m.workspace.sidebarHidden
				h.send(tea.KeyPressMsg{Code: tea.KeyF2})
				h.until(func() bool { return h.m.rename != nil })
				if exit == "save" {
					h.send(tea.KeyPressMsg{Code: '!', Text: "!"})
					h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
				} else {
					h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
				}
				h.until(func() bool { return h.m.rename == nil && h.m.workspace.conn != nil && !h.m.workspace.busy })
				if h.m.workspace.focused != focused || h.m.workspace.sidebarHidden != hidden {
					t.Fatalf("rename changed origin mode: focus=%v hidden=%v; want focus=%v hidden=%v", h.m.workspace.focused, h.m.workspace.sidebarHidden, focused, hidden)
				}
				if focused && h.m.workspaceFooterVisible() {
					t.Fatal("terminal return still shows management footer")
				}
				h.send(tea.KeyPressMsg{Code: 'j', Mod: tea.ModAlt})
				h.until(func() bool { return h.m.workspace.active == 1 && h.m.workspace.conn != nil && !h.m.workspace.busy })
				if h.m.workspace.focused != focused || h.m.workspace.sidebarHidden != hidden {
					t.Fatal("Alt+J changed origin mode")
				}
				h.send(tea.KeyPressMsg{Code: 'k', Mod: tea.ModAlt})
				h.until(func() bool { return h.m.workspace.active == 0 && h.m.workspace.conn != nil && !h.m.workspace.busy })
				if h.m.workspace.focused != focused || h.m.workspace.sidebarHidden != hidden {
					t.Fatal("Alt+K changed origin mode")
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				for _, stream := range s.streams {
					stream.mu.Lock()
					writes := strings.Join(stream.writes, ";")
					stream.mu.Unlock()
					if strings.Contains(writes, "text:") || strings.Contains(writes, "key:") {
						t.Fatalf("shortcut leaked to terminal: %s", writes)
					}
				}
			})
		}
	}
}

func TestAltJKPreservesCurrentFocusWithoutRename(t *testing.T) {
	for _, focused := range []bool{true, false} {
		h := newWorkspaceHarness(t, directDragFixture())
		h.until(func() bool { return h.m.workspace.conn != nil })
		if !focused {
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
		}
		for _, step := range []struct {
			key   rune
			index int
		}{{'j', 1}, {'k', 0}} {
			h.send(tea.KeyPressMsg{Code: step.key, Mod: tea.ModAlt})
			h.until(func() bool {
				return h.m.workspace.active == step.index && h.m.workspace.conn != nil && !h.m.workspace.busy
			})
			if h.m.workspace.focused != focused {
				t.Fatalf("Alt+%c changed focus %v", step.key, focused)
			}
		}
	}
}
