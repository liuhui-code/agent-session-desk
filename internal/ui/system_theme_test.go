package ui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/appearance"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"strings"
)

func TestSystemThemeCanBeSelectedAndPersistsChoice(t *testing.T) {
	t.Setenv("COLORFGBG", "0;15")
	var saved config.UISettings
	h := newWorkspaceHarness(t, directDragFixture(), Options{SaveUISettings: func(_ context.Context, s config.UISettings) error { saved = s; return nil }})
	openThemeSettings(t, h)
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	h.until(func() bool { return !h.m.preferences.busy })
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	h.until(func() bool { return !h.m.preferences.busy })
	if saved.Theme != "system" {
		t.Fatalf("System is not selectable: %+v", saved)
	}
	h.until(func() bool { return colorHex(h.m.View().BackgroundColor) == "d5d6db" })
}

func TestSystemThemeObservationsNeverSavePreferenceOrWriteTerminalInput(t *testing.T) {
	a := newControlledAppearance()
	s := directDragFixture()
	h := newWorkspaceHarness(t, s, Options{Theme: "system", OpenSystemTheme: func(context.Context) (appearance.Connection, error) { return a, nil }, SaveUISettings: func(context.Context, config.UISettings) error { t.Error("observation wrote preferences"); return nil }})
	h.until(func() bool {
		return h.m.systemTheme.op != nil && h.m.systemTheme.op.reading && h.m.workspace.conn != nil
	})
	a.values <- appearance.Observation{Theme: "light", Source: "windows"}
	h.until(func() bool { return colorHex(h.m.View().BackgroundColor) == "d5d6db" })
	if h.m.preferences.saved.Theme != "system" || h.m.preferences.busy {
		t.Fatal("observation replaced persisted choice")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, stream := range s.streams {
		stream.mu.Lock()
		for _, v := range stream.writes {
			if strings.HasPrefix(v, "text:") || strings.HasPrefix(v, "key:") {
				t.Error("observation wrote terminal input")
			}
		}
		stream.mu.Unlock()
	}
}
