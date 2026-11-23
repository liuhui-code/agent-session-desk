package ui

import (
	"context"
	"fmt"
	"image/color"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func colorHex(c color.Color) string {
	if c == nil {
		return "nil"
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("%02x%02x%02x", r>>8, g>>8, b>>8)
}

func TestThemeSettingsSaveLightPaletteAndRestoreOnNewModelWithoutTerminalInput(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ui.toml")
	s := directDragFixture()
	frame := s.frame
	s.frame = func(v session.Session, w, h int) terminal.Screen {
		f := frame(v, w, h)
		f.Rows[0].Runs[0].Foreground = "idx:2"
		f.Rows[0].Runs[0].Background = "idx:4"
		f.Rows[0].Runs[0].Flags = 1
		return f
	}
	h := newWorkspaceHarness(t, s, Options{SaveUISettings: func(_ context.Context, v config.UISettings) error { return config.SaveUI(p, v) }})
	h.until(func() bool { return h.m.workspace.conn != nil && len(h.m.workspace.screen.Rows) > 0 })
	beforeRows := h.m.workspace.screen.Rows
	beforeSession := h.m.workspace.tabs[h.m.workspace.active].session
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	h.until(func() bool { return h.m.densitySettings != nil })
	if !strings.Contains(h.m.View().Content, "主题") {
		t.Fatal("settings have no theme row")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	h.until(func() bool { return colorHex(h.m.View().BackgroundColor) == "d5d6db" })
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.until(func() bool {
		return h.m.densitySettings == nil && h.m.workspace.conn != nil && len(h.m.workspace.screen.Rows) > 0
	})
	if !reflect.DeepEqual(beforeRows, h.m.workspace.screen.Rows) {
		t.Fatal("theme rewrote original terminal character attributes")
	}
	if !sameRun(beforeSession, h.m.workspace.tabs[h.m.workspace.active].session) {
		t.Fatal("theme changed original Session/Run")
	}
	settings, e := config.LoadUI(p)
	if e != nil || settings.Theme != "light" || settings.SidebarDensity != "compact" {
		t.Fatalf("theme did not persist without changing density: %+v %v", settings, e)
	}
	reopened := NewWithOptions(context.Background(), s, Options{Theme: settings.Theme})
	defer reopened.Close()
	if colorHex(reopened.View().BackgroundColor) != "d5d6db" {
		t.Fatal("new model did not restore saved light palette")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range s.opened {
		if id != "first" {
			t.Fatal("theme switched original terminal")
		}
	}
	for _, stream := range s.streams {
		stream.mu.Lock()
		for _, v := range stream.writes {
			if strings.HasPrefix(v, "text:") || strings.HasPrefix(v, "key:") {
				t.Fatalf("theme setting leaked input: %s", v)
			}
		}
		stream.mu.Unlock()
	}
}
