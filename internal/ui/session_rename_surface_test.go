package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestSessionRenameSurfaceFillsEveryInteriorCellAfterChildStyleResets(t *testing.T) {
	h := newWorkspaceHarness(t, newRenameFixture())
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Rename session") })
	r, g, b, _ := h.m.tokens.colors.surface.RGBA()
	want := fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8)
	for _, width := range []int{120, 63, 49} {
		h.send(tea.WindowSizeMsg{Width: width, Height: 32})
		for _, withError := range []bool{false, true} {
			h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
			if withError {
				h.send(tea.PasteMsg{Content: "invalid\nname"})
			}
			inside, left, right := false, 0, 0
			for _, row := range strings.Split(h.m.View().Content, "\n") {
				plain := ansi.Strip(row)
				if at := strings.Index(plain, "╭"); at >= 0 {
					left = ansi.StringWidth(plain[:at])
					right = left + ansi.StringWidth(strings.TrimSpace(plain)) - 1
					inside = true
					continue
				}
				if strings.Contains(plain, "╰") {
					break
				}
				if inside {
					for x := left + 1; x < right; x++ {
						if got := testBackgroundAt(row, x); got != want {
							t.Fatalf("%d columns, error=%v, cell %d shows outside background: %s want %s", width, withError, x, got, want)
						}
					}
				}
			}
		}
	}
}
