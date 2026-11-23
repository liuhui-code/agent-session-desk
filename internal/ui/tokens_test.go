package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestDesignTokensReachVisibleListAndHelp(t *testing.T) {
	m := testModel(t, fakeService{})
	m.width, m.height = 100, 12
	m.tokens.marks.below = "v"
	m.tokens.marks.menuDot = ":"
	m.tokens.layout.menuGap = 3
	m.styles = newStyles(m.tokens, m.colorProfile)
	m.sessions = append(m.sessions, m.sessions...)
	m.sessions = append(m.sessions, m.sessions...)
	m.sessions = append(m.sessions, m.sessions...)
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "v 下方") || strings.Contains(view, "⋮ 下方") {
		t.Fatalf("list indicator ignored its token: %s", view)
	}
	if !strings.Contains(view, "↑↓   :   选择") {
		t.Fatalf("menu spacing ignored its token: %s", view)
	}
}

func TestTokenLayoutChangesKeepViewportAndModelsIndependent(t *testing.T) {
	a := testModel(t, fakeService{})
	b := testModel(t, fakeService{})
	a.tokens.layout.frameGap = 2
	a.tokens.layout.dividerPadding = 2
	a.tokens.layout.listMin = 28
	a.tokens.layout.listMax = 45
	a.tokens.layout.singleBelow = 60
	a.tokens.layout.stackedBelow = 90
	a.tokens.layout.captionGap = 2
	for _, width := range []int{1, 59, 60, 89, 90, 120} {
		for _, height := range []int{1, 5, 10, 24} {
			a.width, a.height = width, height
			rows := strings.Split(a.View().Content, "\n")
			if len(rows) != height {
				t.Fatalf("%dx%d: %d rows", width, height, len(rows))
			}
			for _, row := range rows {
				if lipgloss.Width(row) != width {
					t.Fatalf("%dx%d overflow", width, height)
				}
			}
		}
	}
	if b.tokens.layout.singleBelow != 50 || b.tokens.layout.frameGap != 1 {
		t.Fatal("tokens leaked to another model")
	}
	a.width, a.height = 120, 24
	a.preview = "VISIBLE_PREVIEW"
	a.tokens.layout.singleBelow = 200
	if strings.Contains(a.View().Content, "VISIBLE_PREVIEW") {
		t.Fatal("preview ignored breakpoint token")
	}
}
