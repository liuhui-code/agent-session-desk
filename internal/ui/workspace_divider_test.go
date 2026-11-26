package ui

import (
	"context"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestWorkspaceDividerLeavesTopAndBottomBreathingRoom(t *testing.T) {
	for _, height := range []int{1, 2, 3, 4, 5, 12, 32} {
		m := New(context.Background(), &workspaceFixture{})
		m.width, m.height = 120, height
		g := m.workspaceGeometry()
		lines := strings.Split(m.View().Content, "\n")
		for y := range height {
			cell := ansi.Strip(ansi.Cut(lines[y], g.bodyX+g.left, g.x))
			want := "   "
			if y >= g.bodyY+2 && y < g.bodyY+g.bodyHeight-2 {
				want = " │ "
			}
			if cell != want {
				t.Errorf("height %d row %d: separator %q, want %q", height, y, cell, want)
			}
		}
		if g.height != height-g.bodyY || len(lines) != height {
			t.Errorf("divider inset changed terminal/frame height at %d", height)
		}
		m.Close()
	}
}
