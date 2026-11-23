package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestWorkspaceHelpFillsEveryInteriorCellWithSurface(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	t.Cleanup(m.Close)
	m.help = true
	r, g, b, _ := m.tokens.colors.surface.RGBA()
	want := fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8)
	for _, size := range [][2]int{{120, 32}, {63, 25}, {49, 25}} {
		m.width, m.height = size[0], size[1]
		for _, page := range []helpPage{{}, {all: true}, {all: true, offset: 9999}} {
			m.helpPage = page
			inside := false
			x, width := 0, 0
			for row, line := range strings.Split(m.View().Content, "\n") {
				plain := ansi.Strip(line)
				if i := strings.Index(plain, "╭"); i >= 0 {
					inside = true
					x = ansi.StringWidth(plain[:i])
					width = ansi.StringWidth(strings.TrimSpace(plain))
					continue
				}
				if strings.Contains(plain, "╰") {
					inside = false
					continue
				}
				if !inside {
					continue
				}
				for cell := x + 1; cell < x+width-1; cell++ {
					if got := testBackgroundAt(line, cell); got != want {
						t.Fatalf("%dx%d all=%t offset=%d row=%d cell=%d shows %s, want surface %s", size[0], size[1], page.all, page.offset, row, cell, got, want)
					}
				}
			}
		}
	}
}
