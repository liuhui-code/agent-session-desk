package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestEmptyChineseFormPlaceholdersNeverEmitNULAndKeepTheirRealCursor(t *testing.T) {
	for _, mode := range []string{"group", "launch", "search"} {
		t.Run(mode, func(t *testing.T) {
			s := newGroupedFixture()
			h := newWorkspaceHarness(t, s, Options{})
			h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			key := 'g'
			if mode == "launch" {
				key = 'n'
			} else if mode == "search" {
				key = '/'
			}
			h.send(tea.KeyPressMsg{Code: key, Text: string(key)})
			h.until(func() bool { return h.m.groupDialog != nil || h.m.launch != nil || h.m.search != nil })
			for _, size := range [][2]int{{120, 32}, {63, 25}, {49, 25}, {27, 8}} {
				h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
				view := h.m.View()
				if strings.ContainsRune(view.Content, 0) {
					t.Fatalf("%s placeholder emits NUL at %v", mode, size)
				}
				for _, line := range strings.Split(view.Content, "\n") {
					if ansi.StringWidth(line) != size[0] {
						t.Fatalf("%s overflows %v", mode, size)
					}
				}
				if size[0] >= 49 && view.Cursor == nil {
					t.Fatalf("%s lost the focused real cursor at %v", mode, size)
				}
			}
		})
	}
}
