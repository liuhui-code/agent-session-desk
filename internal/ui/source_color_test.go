package ui

import (
	"context"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestWorkspaceUsesPinnedSourceEmittedRGBWithoutChangingAuthoredPalette(t *testing.T) {
	for _, tc := range []struct{ theme, footer, key, wantForeground, wantBackground, authoredAccent, authoredDim string }{
		{"dark", "compact", "Enter", "0,0,0", "187,187,187", "bbbbbb", "777777"},
		{"light", "curated", "Enter", "105,109,124", "default", "34548a", "6a6d7c"},
	} {
		t.Run(tc.theme, func(t *testing.T) {
			m := NewWithOptions(context.Background(), newGroupedFixture(), Options{Theme: tc.theme, Footer: tc.footer, SidebarMode: "flat"})
			defer m.Close()
			m.width, m.height = 160, 32
			m.sessions, m.cursor, m.workspace.focused = sampleSessions(), 0, false
			line := m.renderManagementFooter()
			// Measure the key column after localized context, not its UTF-8 byte index.
			plain := ansi.Strip(line)
			start := strings.Index(plain, tc.key)
			if start < 0 {
				t.Fatal("missing key")
			}
			start = ansi.StringWidth(plain[:start])
			if fg, bg := testForegroundAt(line, start), testBackgroundAt(line, start); fg != tc.wantForeground || bg != tc.wantBackground {
				t.Fatalf("source RGB mismatch: fg=%s bg=%s", fg, bg)
			}
			if colorHex(m.tokens.colors.accent) != tc.authoredAccent || colorHex(m.tokens.colors.dim) != tc.authoredDim {
				t.Fatal("render adaptation mutated authored palette")
			}
		})
	}
}

func TestSourceColorAdaptationLeavesNativeRGBAndIndexedAttributesExact(t *testing.T) {
	for _, theme := range []string{"dark", "light"} {
		m := NewWithOptions(context.Background(), newGroupedFixture(), Options{Theme: theme, SidebarMode: "flat"})
		t.Cleanup(m.Close)
		m.width, m.height = 160, 32
		m.sessions, m.cursor = sampleSessions(), 0
		m.workspace.tabs = []workspaceTab{{viewID: 1, session: m.sessions[0]}}
		m.workspace.screen = terminal.Screen{Width: 80, Height: 24, Rows: []terminal.Row{{Runs: []terminal.Run{
			{Text: "native-rgb", Width: 10, Foreground: "rgb:122,162,247", Background: "rgb:106,109,124"},
			{Text: "native-idx", Width: 10, Foreground: "idx:2", Background: "idx:4"},
		}}}}
		view := m.View().Content
		for _, tc := range []struct{ text, fg, bg string }{{"native-rgb", "122,162,247", "106,109,124"}, {"native-idx", "idx:2", "idx:4"}} {
			found := false
			for _, line := range strings.Split(view, "\n") {
				plain := ansi.Strip(line)
				if i := strings.Index(plain, tc.text); i >= 0 {
					found = true
					x := ansi.StringWidth(plain[:i])
					if testForegroundAt(line, x) != tc.fg || testBackgroundAt(line, x) != tc.bg {
						t.Fatalf("%s remapped original terminal attributes for %s", theme, tc.text)
					}
				}
			}
			if !found {
				t.Fatalf("%s native terminal text missing", theme)
			}
		}
	}
}
