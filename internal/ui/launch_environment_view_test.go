package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func TestEnvironmentFieldKeepsValueFocusAndBoundsAcrossThemeProfileAndResize(t *testing.T) {
	var frames []map[string]any
	for _, theme := range []string{"dark", "light"} {
		for _, profile := range []colorprofile.Profile{colorprofile.TrueColor, colorprofile.ANSI256, colorprofile.ANSI, colorprofile.ASCII} {
			for _, size := range [][2]int{{120, 40}, {100, 30}, {80, 24}, {63, 16}, {49, 12}, {27, 12}, {10, 4}, {1, 1}} {
				t.Run(fmt.Sprint(theme, "/", profile, "/", size), func(t *testing.T) {
					m := NewWithOptions(context.Background(), &workspaceFixture{}, Options{Theme: theme})
					defer m.Close()
					m.width, m.height = 120, 40
					m.openLaunchDialog()
					m.focusLaunchField(4)
					field := &m.launch.fields[4]
					value := `["ASD_ENV_MARKER=中文 é 1️⃣ ` + strings.Repeat("long-value/", 5) + `Q"]`
					field.SetValue(value)
					field.CursorEnd()
					position := field.Position()
					m.Update(tea.ColorProfileMsg{Profile: profile})
					m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
					v := m.View()
					if field.Value() != value || field.Position() != position || !field.Focused() {
						t.Fatal("display update changed environment draft or focus")
					}
					rows := strings.Split(v.Content, "\n")
					if len(rows) != size[1] {
						t.Fatalf("height %d", len(rows))
					}
					for _, row := range rows {
						if ansi.StringWidth(row) != size[0] {
							t.Fatal("environment form exceeded viewport")
						}
					}
					if size[0] >= 27 && size[1] >= 12 && v.Cursor == nil {
						t.Fatal("usable environment field lost cursor")
					}
					if v.Cursor != nil && (v.Cursor.X < 0 || v.Cursor.X >= size[0] || v.Cursor.Y < 0 || v.Cursor.Y >= size[1]) {
						t.Fatal("environment cursor outside viewport")
					}
					frames = append(frames, map[string]any{"theme": theme, "profile": fmt.Sprint(profile), "width": size[0], "height": size[1], "content": v.Content, "cursor": v.Cursor})
				})
			}
		}
	}
	if path := os.Getenv("ASD_LAUNCH_ENV_FRAMES"); path != "" {
		b, e := json.MarshalIndent(frames, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
}
