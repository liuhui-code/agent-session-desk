package terminal_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestNativeSelectionPaintsWholeUnicodeGraphemesAndKeepsAttributes(t *testing.T) {
	screen := terminal.Screen{Width: 8, Height: 1, CopyMode: true,
		Rows:      []terminal.Row{{Runs: []terminal.Run{{Text: "A中文Z", Width: 6, Flags: 2}}}},
		Selection: &terminal.Selection{StartRow: 0, StartCol: 1, EndRow: 0, EndCol: 4, Mode: "char"}}
	styles := terminal.CopyStyles{Selection: terminal.Run{Foreground: "idx:0", Background: "idx:3"}}
	got := screen.RenderWithCopyStyles(8, 1, styles)
	if ansi.Strip(got) != "A中文Z  " || !strings.Contains(got, "\x1b[1;38;5;0;48;5;3m中文\x1b[0m") {
		t.Fatalf("selection is invisible, splits wide glyphs or drops bold: %q", got)
	}
}

func TestCopyHighlightsRespectSelectionPriorityAndGraphemeBoundaries(t *testing.T) {
	for _, text := range []string{"中文", "éx", "👩‍💻"} {
		t.Run(text, func(t *testing.T) {
			width := ansi.StringWidth(text)
			screen := terminal.Screen{Width: width, Height: 1, CopyMode: true,
				Rows:           []terminal.Row{{Runs: []terminal.Run{{Text: text, Width: width}}}},
				CopyHighlights: []terminal.CopyHighlight{{Row: 0, Start: 0, End: width - 1, Kind: 0}, {Row: 0, Start: 0, End: width - 1, Kind: 1}}}
			styles := terminal.CopyStyles{Selection: terminal.Run{Background: "idx:3"}, Highlights: [4]terminal.Run{{Background: "idx:6"}, {Background: "idx:5"}}}
			got := screen.RenderWithCopyStyles(width, 1, styles)
			if ansi.Strip(got) != text || !strings.Contains(got, "\x1b[48;5;5m"+text+"\x1b[0m") {
				t.Fatal("current search match lost priority or split a grapheme")
			}
			screen.Selection = &terminal.Selection{StartRow: 0, EndRow: 0, Mode: "line"}
			got = screen.RenderWithCopyStyles(width, 1, styles)
			if !strings.Contains(got, "\x1b[48;5;3m"+text+"\x1b[0m") {
				t.Fatal("selection did not paint over search highlights")
			}
		})
	}
}
