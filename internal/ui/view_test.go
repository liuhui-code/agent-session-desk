package ui

import (
	"fmt"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// Unit evidence for the layout boundaries in TC-UI-003/004/010, with ANSI,
// combining characters, CJK, emoji and keycaps. Actual terminal cell rendering
// and Windows input still require the catalog's native/manual cases.
func TestFramesStayInsideViewport(t *testing.T) {
	m := testModel(t, fakeService{})
	m.sessions[0].Title = strings.Repeat("中文👩🏽‍💻é#️⃣", 20)
	m.preview = strings.Repeat("界面 👩🏽‍💻 é 1️⃣ ", 30) + "\n第二行\nthird line"
	for _, width := range []int{1, 2, 10, 22, 49, 50, 79, 80, 100, 160} {
		for _, height := range []int{1, 4, 5, 8, 10, 24, 40} {
			for _, help := range []bool{false, true} {
				t.Run(fmt.Sprintf("%dx%d/help=%t", width, height, help), func(t *testing.T) {
					m.width, m.height, m.help = width, height, help
					view := m.View()
					rows := strings.Split(view.Content, "\n")
					if !view.AltScreen || len(rows) != height {
						t.Fatalf("view has %d rows, want %d, alt=%t", len(rows), height, view.AltScreen)
					}
					for i, row := range rows {
						if cells := lipgloss.Width(row); cells != width {
							t.Fatalf("row %d has %d cells, want %d: %q", i, cells, width, row)
						}
					}
				})
			}
		}
	}
}

func TestSelectedRowVisibleAtListBoundaries(t *testing.T) {
	m := testModel(t, fakeService{})
	m.sessions = nil
	for i := range 100 {
		m.sessions = append(m.sessions, session.Session{ID: fmt.Sprint(i), Title: fmt.Sprintf("task-%03d", i)})
	}
	for _, height := range []int{1, 2, 3, 4, 8, 30} {
		for _, cursor := range []int{0, 1, 49, 98, 99} {
			m.cursor = cursor
			rows := m.renderSessionList(40, height)
			if !strings.Contains(ansi.Strip(rows), fmt.Sprintf("› task-%03d", cursor)) {
				t.Fatalf("selected row invisible: cursor=%d height=%d view=%q", cursor, height, rows)
			}
			if count := len(strings.Split(rows, "\n")); count > height {
				t.Fatalf("list has %d rows for height %d", count, height)
			}
		}
	}
}

func TestPreviewControlsCannotReachManagerTerminal(t *testing.T) {
	source := "\x1b]52;c;c2VjcmV0\a" + "before\x1b[2Jafter\x1b[?1049h" +
		"\x1b]0;untrusted-title\x1b\\" + "\x1bPignored\x1b\\" +
		"\x00\x03\a\r\u009b\u009d\n中文\tend"
	clean, _ := boundedPreview(source)
	if strings.Contains(clean, "secret") || strings.Contains(clean, "untrusted-title") || strings.Contains(clean, "ignored") {
		t.Fatalf("control string payload leaked: %q", clean)
	}
	for _, r := range clean {
		if unicode.IsControl(r) && r != '\n' {
			t.Fatalf("untrusted terminal control U+%04X retained", r)
		}
	}
	if !strings.Contains(clean, "beforeafter") || !strings.Contains(clean, "中文    end") {
		t.Fatalf("ordinary text was not preserved: %q", clean)
	}
	if got := label("hello\n\x1b]0;bad\a\tworld"); got != "hello  world" {
		t.Fatalf("unsafe label: %q", got)
	}
}

func TestPreviewRetentionIsBoundedAndKeepsTail(t *testing.T) {
	for _, input := range []string{
		strings.Repeat("中文输出", maxPreviewBytes) + "tail",
		strings.Repeat("\t", maxPreviewBytes) + "tail",
		strings.Repeat("one line\n", 1000) + "tail",
	} {
		got, truncated := boundedPreview(input)
		if !truncated || len(got) > maxPreviewBytes || len(strings.Split(got, "\n")) > maxPreviewLines || !utf8.ValidString(got) {
			t.Fatalf("retention bounds failed: bytes=%d lines=%d truncated=%t valid=%t", len(got), len(strings.Split(got, "\n")), truncated, utf8.ValidString(got))
		}
		if !strings.HasSuffix(got, "tail") {
			t.Fatal("newest captured text was lost")
		}
	}
}
