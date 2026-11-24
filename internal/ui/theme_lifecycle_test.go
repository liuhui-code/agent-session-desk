package ui

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func openThemeSettings(t *testing.T, h *workspaceHarness) {
	t.Helper()
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'S', Text: "S"})
	h.until(func() bool { return h.m.densitySettings != nil })
}

func TestThemeAndDensitySerializeLatestChoicesBeforeClosing(t *testing.T) {
	started := make(chan config.UISettings, 4)
	release := make(chan struct{}, 2)
	path := filepath.Join(t.TempDir(), "ui.toml")
	h := newWorkspaceHarness(t, directDragFixture(), Options{SaveUISettings: func(ctx context.Context, s config.UISettings) error {
		started <- s
		select {
		case <-release:
			return config.SaveUI(path, s)
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	openThemeSettings(t, h)
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	select {
	case s := <-started:
		if s.Theme != "light" || s.SidebarDensity != "compact" {
			t.Fatalf("first snapshot: %+v", s)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("theme write did not start")
	}
	if colorHex(h.m.View().BackgroundColor) != "000000" {
		t.Fatal("applied unsaved theme")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	h.send(tea.KeyPressMsg{Code: tea.KeyUp})
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	select {
	case s := <-started:
		t.Fatalf("concurrent preference writer: %+v", s)
	default:
	}
	if !strings.Contains(h.m.View().Content, "正在保存") {
		t.Fatal("close discarded accepted choices")
	}
	release <- struct{}{}
	h.until(func() bool { return colorHex(h.m.View().BackgroundColor) == "d5d6db" })
	select {
	case s := <-started:
		if s.Theme != "dark" || s.SidebarDensity != "auto" {
			t.Fatalf("latest combined snapshot: %+v", s)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("latest choices not serialized")
	}
	old := h.m.preferences.generation - 1
	release <- struct{}{}
	h.until(func() bool { return h.m.densitySettings == nil && h.m.workspace.conn != nil })
	h.send(uiPreferencesSaved{generation: old, settings: config.UISettings{Theme: "light", SidebarDensity: "compact"}})
	s, e := config.LoadUI(path)
	if e != nil || s.Theme != "dark" || s.SidebarDensity != "auto" || colorHex(h.m.View().BackgroundColor) != "000000" {
		t.Fatalf("latest persisted/applied choice lost: %+v %v", s, e)
	}
}

func TestThemeFailedSavePreservesAppliedPaletteAndCancelsWaitingClose(t *testing.T) {
	h := newWorkspaceHarness(t, directDragFixture(), Options{Theme: "light", SaveUISettings: func(context.Context, config.UISettings) error { return errors.New("synthetic theme write refusal") }})
	openThemeSettings(t, h)
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.until(func() bool {
		return h.m.densitySettings != nil && !h.m.densitySettings.busy && h.m.densitySettings.err != nil
	})
	h.send(tea.WindowSizeMsg{Width: 27, Height: 8})
	if colorHex(h.m.View().BackgroundColor) != "d5d6db" || !strings.Contains(h.m.View().Content, "未保存") || h.m.densitySettings.closing {
		t.Fatal("failed save applied dark, hid error or closed dialog")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.until(func() bool { return h.m.densitySettings == nil })
}

func TestThemeNeedsWholeDocumentSaveCapability(t *testing.T) {
	h := newWorkspaceHarness(t, directDragFixture(), Options{SaveSidebarDensity: func(context.Context, string) error { t.Error("theme fell back to density writer"); return nil }})
	openThemeSettings(t, h)
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	h.until(func() bool { return h.m.densitySettings.err != nil })
	if colorHex(h.m.View().BackgroundColor) != "000000" || !strings.Contains(h.m.View().Content, "不支持主题") {
		t.Fatal("unsupported writer falsely applied theme")
	}
}
