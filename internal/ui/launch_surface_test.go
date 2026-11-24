package ui

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestLaunchDialogSurfaceFillsTrailingLabelCells(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }}}
	h := newWorkspaceHarness(t, service)
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	width := h.m.launchWidth() - h.m.styles.dialog.GetHorizontalFrameSize()
	r, g, b, _ := h.m.tokens.colors.surface.RGBA()
	want := fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8)
	found := false
	for _, line := range strings.Split(h.m.View().Content, "\n") {
		plain := ansi.Strip(line)
		idx := strings.Index(plain, "名称（可空）")
		if idx < 0 {
			continue
		}
		found = true
		x := ansi.StringWidth(plain[:idx]) + width - 1
		if got := testBackgroundAt(line, x); got != want {
			t.Fatalf("dialog label trailing cell shows outside background: got %s want %s", got, want)
		}
	}
	if !found {
		t.Fatal("creation form label not visible")
	}
}

// Track generated SGR in one row; this checks visible fill after child styles
// reset, not whether a particular Style constructor has a Background property.
func testBackgroundAt(row string, x int) string {
	return testColorAt(row, x, 48, 49)
}

func testForegroundAt(row string, x int) string {
	return testColorAt(row, x, 38, 39)
}

func testColorAt(row string, x, role, reset int) string {
	matches := regexp.MustCompile(`\x1b\[([0-9;]*)m`).FindAllStringSubmatchIndex(row, -1)
	bg := "default"
	cell, offset := 0, 0
	for _, m := range matches {
		width := ansi.StringWidth(row[offset:m[0]])
		if x < cell+width {
			return bg
		}
		cell += width
		raw := row[m[2]:m[3]]
		if raw == "" {
			raw = "0"
		}
		var codes []int
		for _, v := range strings.Split(raw, ";") {
			n, _ := strconv.Atoi(v)
			codes = append(codes, n)
		}
		for i := 0; i < len(codes); i++ {
			switch codes[i] {
			case 0, reset:
				bg = "default"
			case 30, 31, 32, 33, 34, 35, 36, 37:
				if role == 38 {
					bg = fmt.Sprintf("idx:%d", codes[i]-30)
				}
			case 40, 41, 42, 43, 44, 45, 46, 47:
				if role == 48 {
					bg = fmt.Sprintf("idx:%d", codes[i]-40)
				}
			case 90, 91, 92, 93, 94, 95, 96, 97:
				if role == 38 {
					bg = fmt.Sprintf("idx:%d", codes[i]-90+8)
				}
			case 100, 101, 102, 103, 104, 105, 106, 107:
				if role == 48 {
					bg = fmt.Sprintf("idx:%d", codes[i]-100+8)
				}
			case 38, 48:
				kind := codes[i]
				if i+4 < len(codes) && codes[i+1] == 2 {
					if kind == role {
						bg = fmt.Sprintf("%d,%d,%d", codes[i+2], codes[i+3], codes[i+4])
					}
					i += 4
				} else if i+2 < len(codes) && codes[i+1] == 5 {
					if kind == role {
						bg = fmt.Sprintf("idx:%d", codes[i+2])
					}
					i += 2
				}
			}
		}
		offset = m[1]
	}
	return bg
}
