package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func TestThemesKeepModelsIndependentAndRetainLayoutTokens(t *testing.T) {
	dark := New(context.Background(), directDragFixture())
	defer dark.Close()
	light := NewWithOptions(context.Background(), directDragFixture(), Options{Theme: "light"})
	defer light.Close()
	light.Update(tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	if colorHex(dark.tokens.colors.bg) != "1a1b26" || colorHex(light.tokens.colors.bg) != "d5d6db" || colorHex(dark.styles.text.GetForeground()) != "c0caf5" || colorHex(light.styles.text.GetForeground()) != "343b58" {
		t.Fatal("model palettes shared or styles stale")
	}
	if dark.tokens.layout != light.tokens.layout || dark.tokens.workspace != light.tokens.workspace {
		t.Fatal("theme altered geometry")
	}
	if colorHex(light.tokens.footer.contextColor) != "7847bd" || colorHex(light.tokens.help.keyColor) != "7847bd" || colorHex(light.tokens.rename.borderColor) != "7847bd" {
		t.Fatal("light palette retained dark purple sub-roles")
	}
	light.applyUITheme("dark")
	if colorHex(dark.tokens.colors.bg) != "1a1b26" || colorHex(light.tokens.colors.bg) != "1a1b26" {
		t.Fatal("one theme switch mutated other model")
	}
	light.applyUITheme("light")
	light.history.open = true
	if colorHex(light.View().BackgroundColor) != "d5d6db" {
		t.Fatal("history view reset light theme")
	}
}

func TestThemeSettingsKeepFocusedChoiceInsideAllBoundedFrames(t *testing.T) {
	for _, theme := range []string{"dark", "light", "system"} {
		for _, width := range []int{8, 12, 27, 49, 80, 120} {
			for _, height := range []int{1, 5, 8, 12, 24} {
				for _, row := range []int{0, 1} {
					t.Run(fmt.Sprintf("%s/%dx%d/row%d", theme, width, height, row), func(t *testing.T) {
						m := NewWithOptions(context.Background(), directDragFixture(), Options{Theme: theme, SidebarDensity: "auto"})
						defer m.Close()
						m.width, m.height = width, height
						m.openDensitySettings()
						m.densitySettings.row = row
						view := m.View().Content
						lines := strings.Split(view, "\n")
						if len(lines) != height {
							t.Fatalf("settings height %d > %d", len(lines), height)
						}
						for _, line := range lines {
							if ansi.StringWidth(line) != width {
								t.Fatalf("settings width %d != %d", ansi.StringWidth(line), width)
							}
						}
						if width == 27 && height == 8 {
							want := ">Auto"
							if row == 0 {
								want = ">Dark"
								if theme == "light" {
									want = ">Light"
								} else if theme == "system" {
									want = ">System"
								}
							}
							if !strings.Contains(ansi.Strip(view), want) {
								t.Fatalf("focused choice cropped: %s", view)
							}
						}
					})
				}
			}
		}
	}
}
