package terminal_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestNativeScreenRenderingKeepsCellStylesAndViewport(t *testing.T) {
	screen := terminal.Screen{Width: 8, Height: 2, Rows: []terminal.Row{{Runs: []terminal.Run{{Text: "中文A", Width: 5, Foreground: "rgb:122,162,247", Background: "idx:0", Flags: 2}}}}}
	rendered := screen.Render(8, 2)
	rows := strings.Split(rendered, "\n")
	if len(rows) != 2 {
		t.Fatalf("wrong viewport rows: %q", rendered)
	}
	for _, row := range rows {
		if ansi.StringWidth(row) != 8 {
			t.Fatalf("wrong viewport width: %q", row)
		}
	}
	if ansi.Strip(rows[0]) != "中文A   " || !strings.Contains(rows[0], "38;2;122;162;247") || !strings.Contains(rows[0], "48;5;0") {
		t.Fatalf("native cells/styles lost: %q", rendered)
	}
}
