package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

func TestDirectoryPopupUsesMeasuredSourceDisplayBackground(t *testing.T) {
	tokens := defaultDesignTokens()
	if tokens.paths.background != lipgloss.Color("#222222") {
		t.Fatal("authored dark color changed")
	}
	for _, c := range []struct {
		profile     colorprofile.Profile
		theme, want string
	}{
		{colorprofile.TrueColor, "dark", "#222222"}, {colorprofile.ANSI, "dark", "0"}, {colorprofile.ANSI256, "dark", "235"},
		{colorprofile.TrueColor, "light", "#dcdde2"}, {colorprofile.ANSI, "light", "7"}, {colorprofile.ANSI256, "light", "188"},
	} {
		if got := newStyles(tokens.withTheme(c.theme), c.profile).pathBox.GetBackground(); got != lipgloss.Color(c.want) {
			t.Fatalf("%v/%s background: %v want %s", c.profile, c.theme, got, c.want)
		}
	}
}
func TestDirectoryPopupViewportFocusAndThemeMatrix(t *testing.T) {
	var frames []map[string]any
	for _, mode := range []string{"launch", "group"} {
		for _, theme := range []string{"dark", "light"} {
			for _, profile := range []colorprofile.Profile{colorprofile.TrueColor, colorprofile.ANSI256, colorprofile.ANSI, colorprofile.ASCII} {
				for _, size := range [][2]int{{120, 40}, {100, 30}, {80, 24}, {50, 16}, {32, 12}, {31, 11}, {16, 8}, {4, 4}} {
					name := fmt.Sprint(mode, "/", theme, "/", profile, "/", size)
					t.Run(name, func(t *testing.T) {
						m := NewWithOptions(context.Background(), newGroupedFixture(), Options{Theme: theme})
						defer m.Close()
						m.width, m.height = size[0], size[1]
						m.colorProfileEvent(tea.ColorProfileMsg{Profile: profile})
						var p *pathPicker
						if mode == "launch" {
							m.openLaunchDialog()
							m.focusLaunchField(2)
							p = m.launch.paths
							m.launch.fields[2].SetValue("directory-draft")
						} else {
							m.groupIntent = &groupDialog{}
							m.openGroupDialog()
							m.groupDialog.focus = 1
							m.groupDialog.inputs[0].Blur()
							m.groupDialog.inputs[1].Focus()
							m.groupDialog.inputs[1].SetValue("directory-draft")
							p = m.groupDialog.paths
						}
						p.matches = []string{"/projects/Alpha", "/projects/alpine", "/work/中文 空格", "/work/other", "/work/five", "/work/six", "/work/seven"}
						p.selected = 3
						p.active = mode == "launch"
						v := m.View()
						rows := strings.Split(v.Content, "\n")
						if len(rows) != size[1] {
							t.Fatalf("height: %d", len(rows))
						}
						for i, row := range rows {
							if lipgloss.Width(row) != size[0] {
								t.Fatalf("row %d width %d", i, lipgloss.Width(row))
							}
						}
						if v.Cursor != nil && (v.Cursor.X < 0 || v.Cursor.Y < 0 || v.Cursor.X >= size[0] || v.Cursor.Y >= size[1]) {
							t.Fatal("cursor outside viewport")
						}
						if size[0] >= 32 && size[1] >= 12 && !strings.Contains(v.Content, "Select") && !strings.Contains(v.Content, "浏览") {
							t.Fatal("operable picker not rendered")
						}
						if !m.pathPopupAvailable() {
							if _, ok := m.pathPickerKey(p, "esc", ""); ok {
								t.Fatal("invisible popup trapped Escape")
							}
						}
						frames = append(frames, map[string]any{"mode": mode, "theme": theme, "profile": fmt.Sprint(profile), "width": size[0], "height": size[1], "content": v.Content, "cursor": v.Cursor})
					})
				}
			}
		}
	}
	if out := os.Getenv("ASD_DIRECTORY_FRAMES"); out != "" {
		b, e := json.MarshalIndent(frames, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(out, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
}
