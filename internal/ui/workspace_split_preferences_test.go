package ui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestWorkspaceSplitExplicitKeysPersistAndAutoReset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.toml")
	h := newWorkspaceHarness(t, newStopFixture(), Options{SaveUISettings: func(_ context.Context, s config.UISettings) error { return config.SaveUI(path, s) }})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '<', Text: "<"})
	h.until(func() bool { b, _ := os.ReadFile(path); return strings.Contains(string(b), "sidebar_percent = 40") })
	if h.m.workspaceSidebarPercent() != 40 {
		t.Fatal("manual split not visible")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	if h.m.workspaceSidebarPercent() != 40 {
		t.Fatal("focus discarded manual ratio")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '=', Text: "="})
	h.until(func() bool { b, _ := os.ReadFile(path); return strings.Contains(string(b), "sidebar_percent = 0") })
	if h.m.workspaceSidebarPercent() != 35 {
		t.Fatal("auto reset lost management default")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	if h.m.workspaceSidebarPercent() != 18 {
		t.Fatal("auto reset lost terminal default")
	}
}

func TestWorkspaceSplitRestoresManualAcrossFileLoadAndKeepsNativeDraftIdentity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.toml")
	s := newStopFixture()
	h := newWorkspaceHarness(t, s, Options{SaveUISettings: func(_ context.Context, p config.UISettings) error { return config.SaveUI(path, p) }})
	h.until(func() bool { return h.m.workspace.conn != nil })
	s.mu.Lock()
	original := s.streams["first"]
	s.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "original-draft"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '<', Text: "<"})
	h.until(func() bool { return !h.m.preferences.busy && h.m.preferences.saved.SidebarPercent == 40 })
	settings, e := config.LoadUI(path)
	if e != nil {
		t.Fatal(e)
	}
	reopened := newWorkspaceHarness(t, newStopFixture(), Options{SidebarPercent: settings.SidebarPercent, SidebarDensity: settings.SidebarDensity, SidebarMode: settings.SidebarMode})
	reopened.until(func() bool { return reopened.m.workspace.conn != nil })
	if reopened.m.workspaceSidebarPercent() != 40 || !sameRun(reopened.m.workspace.tabs[0].session, h.m.workspace.tabs[0].session) {
		t.Fatal("reopen lost manual ratio or original identity")
	}
	original.mu.Lock()
	defer original.mu.Unlock()
	if strings.Count(strings.Join(original.writes, ","), "text:original-draft") != 1 {
		t.Fatal("original draft lost/replayed")
	}
	for _, v := range original.writes {
		if v == "text:<" {
			t.Fatal("management ratio key leaked")
		}
	}
}

func TestWorkspaceSplitDragPersistsOnlyOnLeftRelease(t *testing.T) {
	writes := make(chan config.UISettings, 8)
	h := newWorkspaceHarness(t, newStopFixture(), Options{SidebarPercent: 40, SaveUISettings: func(_ context.Context, s config.UISettings) error { writes <- s; return nil }})
	h.until(func() bool { return h.m.workspace.conn != nil })
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.left + 1, Y: g.y + 1, Button: tea.MouseLeft})
	for _, x := range []int{50, 55, 60} {
		h.send(tea.MouseMotionMsg{X: x, Y: g.y + 1, Button: tea.MouseLeft})
	}
	if h.m.workspaceSidebarPercent() != 50 || h.m.preferences.wanted.SidebarPercent != 40 || h.m.preferences.busy {
		t.Fatal("motion became an accepted preference")
	}
	h.send(tea.MouseReleaseMsg{X: 60, Y: g.y + 1, Button: tea.MouseLeft})
	h.until(func() bool { return !h.m.preferences.busy && h.m.preferences.saved.SidebarPercent == 50 })
	select {
	case v := <-writes:
		if v.SidebarPercent != 50 {
			t.Fatal("wrong final drag ratio")
		}
	default:
		t.Fatal("release did not save")
	}
	select {
	case <-writes:
		t.Fatal("motion/release created extra writes")
	default:
	}
}

func TestWorkspaceSplitEndedDragSavesLastPreviewOnce(t *testing.T) {
	for _, kind := range []string{"key", "click", "wheel", "motion", "resize", "switch", "none-release"} {
		t.Run(kind, func(t *testing.T) {
			writes := make(chan config.UISettings, 8)
			h := newWorkspaceHarness(t, newStopFixture(), Options{SidebarPercent: 40, SaveUISettings: func(_ context.Context, s config.UISettings) error { writes <- s; return nil }})
			h.until(func() bool { return h.m.workspace.conn != nil })
			g := h.m.workspaceGeometry()
			h.send(tea.MouseClickMsg{X: g.left + 1, Y: g.y + 1, Button: tea.MouseLeft})
			h.send(tea.MouseMotionMsg{X: 60, Y: g.y + 1, Button: tea.MouseLeft})
			if h.m.workspaceSidebarPercent() != 50 {
				t.Fatal("drag did not preview")
			}
			switch kind {
			case "key":
				h.send(tea.KeyPressMsg{Code: 'z', Text: "z"})
			case "click":
				h.send(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
			case "wheel":
				h.send(tea.MouseWheelMsg{X: 0, Y: 0, Button: tea.MouseWheelUp})
			case "motion":
				h.send(tea.MouseMotionMsg{X: 0, Y: 0, Button: tea.MouseNone})
			case "resize":
				h.send(tea.WindowSizeMsg{Width: 120, Height: 32})
			case "switch":
				h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
			case "none-release":
				h.send(tea.MouseReleaseMsg{X: 0, Y: g.y + 1, Button: tea.MouseNone})
			}
			h.send(tea.MouseReleaseMsg{X: 60, Y: g.y + 1, Button: tea.MouseLeft})
			h.until(func() bool { return !h.m.preferences.busy && h.m.preferences.saved.SidebarPercent == 50 })
			if h.m.workspaceSidebarPercent() != 50 || h.m.workspace.split.dragging || h.m.preferences.wanted.SidebarPercent != 50 {
				t.Fatal("drag end lost last preview")
			}
			select {
			case v := <-writes:
				if v.SidebarPercent != 50 {
					t.Fatal("wrong ended ratio")
				}
			default:
				t.Fatal("drag end not saved")
			}
			select {
			case <-writes:
				t.Fatal("drag end saved twice")
			default:
			}
		})
	}
}

func TestWorkspaceSplitSavesLatestWholeDocumentAndOldAckDoesNotUndoPreview(t *testing.T) {
	started := make(chan config.UISettings, 8)
	release := make(chan struct{}, 8)
	h := newWorkspaceHarness(t, newStopFixture(), Options{SidebarPercent: 25, SaveUISettings: func(ctx context.Context, s config.UISettings) error {
		started <- s
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '<', Text: "<"})
	h.until(func() bool { return len(started) == 1 })
	if v := <-started; v.SidebarPercent != 30 {
		t.Fatal("wrong first choice")
	}
	h.send(tea.KeyPressMsg{Code: '<', Text: "<"})
	h.send(tea.KeyPressMsg{Code: '<', Text: "<"})
	// A background whole-document save must not adopt the transient drag.
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.left + 1, Y: g.y + 1, Button: tea.MouseLeft})
	h.send(tea.MouseMotionMsg{X: 60, Y: g.y + 1, Button: tea.MouseLeft})
	h.m.preferences.wanted.SidebarDensity = "full"
	h.run(h.m.saveUIPreferences())
	if len(started) != 0 || h.m.preferences.wanted.SidebarPercent != 40 {
		t.Fatal("preview leaked to concurrent save")
	}
	release <- struct{}{}
	h.until(func() bool { return len(started) == 1 })
	if v := <-started; v.SidebarPercent != 40 || v.SidebarDensity != "full" {
		t.Fatalf("newest whole document lost fields: %+v", v)
	}
	if h.m.workspaceSidebarPercent() != 50 || !h.m.workspace.split.dragging {
		t.Fatal("old ack undid current drag preview")
	}
	release <- struct{}{}
	h.until(func() bool { return !h.m.preferences.busy })
	if h.m.workspaceSidebarPercent() != 50 {
		t.Fatal("save completion changed preview")
	}
	h.send(tea.MouseReleaseMsg{X: 60, Y: g.y + 1, Button: tea.MouseLeft})
	h.until(func() bool { return len(started) == 1 })
	if v := <-started; v.SidebarPercent != 50 || v.SidebarDensity != "full" {
		t.Fatal("release lost coalesced preferences")
	}
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if h.m.closed {
		t.Fatal("quit cancelled accepted drag release")
	}
	release <- struct{}{}
	h.until(func() bool { return h.m.closed })
}

func TestWorkspaceSplitSaveFailureRollsBackAndCancelsWaitingQuit(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{}, 1)
	h := newWorkspaceHarness(t, newStopFixture(), Options{SidebarPercent: 35, SaveUISettings: func(ctx context.Context, _ config.UISettings) error {
		started <- struct{}{}
		select {
		case <-release:
			return errors.New("synthetic preference failure")
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return h.m.workspace.conn != nil && h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '<', Text: "<"})
	h.until(func() bool { return len(started) == 1 })
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if h.m.closed || h.m.workspace.transition != workspaceQuit {
		t.Fatal("quit did not wait for preference")
	}
	close(release)
	h.until(func() bool { return !h.m.preferences.busy })
	if h.m.closed || h.m.workspace.transition != workspaceIdle || h.m.workspaceSidebarPercent() != 35 || h.m.preferences.wanted.SidebarPercent != 35 || !strings.Contains(h.m.status, "preferences not saved") {
		t.Fatal("failed preference hid error, kept new ratio, or quit")
	}
	<-started
	select {
	case <-started:
		t.Fatal("failure automatically retried")
	default:
	}
}

func TestWorkspaceSplitRollbackErrorRemainsVisibleWhenResizeCannotBeEnqueued(t *testing.T) {
	h := newWorkspaceHarness(t, newStopFixture(), Options{SidebarPercent: 35})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.m.workspace.split.override, h.m.workspace.split.percent = true, 40
	h.m.preferences.wanted.SidebarPercent = 40
	h.m.preferences.busy, h.m.preferences.generation = true, 77
	// A failed/disconnected writer cannot enqueue a restoring size event. The
	// preference failure must still remain visible, with its saved ratio restored.
	h.m.workspace.inputs = nil
	h.send(uiPreferencesSaved{generation: 77, settings: config.UISettings{SidebarPercent: 40}, err: errors.New("synthetic save failure")})
	if h.m.workspaceSidebarPercent() != 35 || !strings.Contains(h.m.status, "preferences not saved") {
		t.Fatalf("rollback masked preference error: %s", h.m.status)
	}
}

func TestWorkspaceSplitReadAndFocusDoNotSaveUntilUserChangesChoice(t *testing.T) {
	writes := make(chan config.UISettings, 8)
	h := newWorkspaceHarness(t, newStopFixture(), Options{SidebarPercent: 40, SaveUISettings: func(_ context.Context, s config.UISettings) error { writes <- s; return nil }})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	g := h.m.workspaceGeometry()
	h.send(tea.MouseMotionMsg{X: g.left + 1, Y: g.y, Button: tea.MouseNone})
	h.send(tea.MouseClickMsg{X: g.left + 1, Y: g.y, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: 0, Y: g.y, Button: tea.MouseNone})
	h.send(tea.WindowSizeMsg{Width: 63, Height: 32})
	h.send(tea.WindowSizeMsg{Width: 120, Height: 32})
	if h.m.preferences.busy || h.m.workspaceSidebarPercent() != 40 {
		t.Fatal("incidental event changed choice")
	}
	select {
	case <-writes:
		t.Fatal("startup/focus/hover/plain click/resize saved preferences")
	default:
	}
	h.send(tea.KeyPressMsg{Code: '<', Text: "<"})
	if h.m.preferences.busy {
		t.Fatal("terminal operator changed preference")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '>', Text: ">"})
	h.until(func() bool { return !h.m.preferences.busy && h.m.preferences.saved.SidebarPercent == 35 })
}

func TestWorkspaceSplitUnchangedChoiceDoesNotPersistTemporaryGroupReveal(t *testing.T) {
	for _, c := range []struct {
		percent int
		key     string
	}{{0, "="}, {10, ">"}, {90, "<"}} {
		t.Run(c.key, func(t *testing.T) {
			h := newWorkspaceHarness(t, newStopFixture(), Options{SidebarPercent: c.percent, DefaultGroupCollapsed: true})
			h.until(func() bool { return h.m.workspace.conn != nil })
			h.m.revealGroup("default")
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			h.send(tea.KeyPressMsg{Code: rune(c.key[0]), Text: c.key})
			if h.m.preferences.busy || !h.m.preferences.accepted.DefaultGroupCollapsed {
				t.Fatal("no-op ratio action saved temporary reveal")
			}
		})
	}
}
