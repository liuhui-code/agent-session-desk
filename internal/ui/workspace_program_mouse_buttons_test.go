package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func (s *workspaceStream) Mouse(ctx context.Context, action, button string, modifiers uint8, x, y int) error {
	if button == "left" && modifiers == 0 {
		return s.Pointer(ctx, action, x, y)
	}
	return s.record(ctx, fmt.Sprintf("mouse:%s,%s,%d,%d,%d", action, button, modifiers, x, y))
}
func TestProgramMouseRightGestureKeepsModifiersAndOriginalTarget(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	mods := tea.ModShift | tea.ModCtrl
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseRight, Mod: mods})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseRight, Mod: mods})
	h.send(tea.MouseReleaseMsg{X: g.x + 9, Y: g.y + 2, Button: tea.MouseRight, Mod: mods})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	got := directDragWrites(s, "first")
	want := "mouse:down,right,5,3,2;mouse:drag,right,5,8,2;mouse:up,right,5,9,2"
	if got != want {
		t.Fatalf("right Shift+Ctrl gesture missing: %q", got)
	}
	if strings.Contains(got, "copy") || strings.Contains(got, "key:") || strings.Contains(got, "text:") {
		t.Fatal("program mouse became copy or keyboard input")
	}
}

func TestProgramMouseThreeButtonsAndModifierMatrix(t *testing.T) {
	for _, button := range []struct {
		key  tea.MouseButton
		name string
	}{{tea.MouseLeft, "left"}, {tea.MouseMiddle, "middle"}, {tea.MouseRight, "right"}} {
		for mask := range 8 {
			t.Run(fmt.Sprint(button.name, "/", mask), func(t *testing.T) {
				h, s := newProgramMouseHarness(t)
				g := h.m.workspaceGeometry()
				var mod tea.KeyMod
				if mask&1 != 0 {
					mod |= tea.ModShift
				}
				if mask&2 != 0 {
					mod |= tea.ModAlt
				}
				if mask&4 != 0 {
					mod |= tea.ModCtrl
				}
				mod |= tea.ModCapsLock | tea.ModNumLock | tea.ModScrollLock
				h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: button.key, Mod: mod})
				h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: button.key, Mod: mod})
				h.send(tea.MouseReleaseMsg{X: g.x + 9, Y: g.y + 2, Button: button.key, Mod: mod})
				h.until(func() bool { return h.m.workspace.pending == 0 })
				want := fmt.Sprintf("mouse:down,%s,%d,3,2;mouse:drag,%s,%d,8,2;mouse:up,%s,%d,9,2", button.name, mask, button.name, mask, button.name, mask)
				if button.name == "left" && mask == 0 {
					want = "pointer:down,3,2;pointer:drag,8,2;pointer:up,9,2"
				}
				if got := directDragWrites(s, "first"); got != want {
					t.Fatalf("gesture changed: %q", got)
				}
			})
		}
	}
}
func TestProgramMouseModifierChangeReleasesLastAcceptedOriginalStateBeforeSwitch(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseMiddle, Mod: tea.ModShift})
	h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseMiddle, Mod: tea.ModAlt})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: 'z', Text: "z"})
	h.until(func() bool {
		return h.m.workspace.active == 1 && h.m.workspace.conn != nil && h.m.workspace.pending == 0
	})
	h.send(tea.MouseReleaseMsg{X: g.x + 9, Y: g.y + 2, Button: tea.MouseMiddle, Mod: tea.ModAlt})
	if got := directDragWrites(s, "first"); got != "mouse:down,middle,1,3,2;mouse:drag,middle,2,8,2;mouse:up,middle,2,8,2" {
		t.Fatalf("old gesture release: %q", got)
	}
	if got := directDragWrites(s, "second"); got != "text:z" {
		t.Fatalf("old mouse reached new Run: %q", got)
	}
}
func TestLegacyPointerConnectionNeverDropsNewButtonOrModifierIntoLeftSelection(t *testing.T) {
	base := programMouseFixture()
	h := newWorkspaceHarness(t, withoutDirectDragService{base})
	h.until(func() bool { return h.m.workspace.screen.WantsMouse })
	g := h.m.workspaceGeometry()
	for _, button := range []tea.MouseButton{tea.MouseMiddle, tea.MouseRight} {
		h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: button})
		h.send(tea.MouseReleaseMsg{X: g.x + 3, Y: g.y + 2, Button: button})
	}
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft, Mod: tea.ModAlt})
	h.send(tea.MouseMotionMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseNone})
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseLeft})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if got := directDragWrites(base, "first"); got != "pointer:down,3,2;pointer:up,3,2" {
		t.Fatalf("legacy connection received unsupported mouse: %q", got)
	}
}

func TestProgramMousePressureReleasesLastAcceptedModifierBeforeSwitch(t *testing.T) {
	h, s := newProgramMouseHarness(t)
	g := h.m.workspaceGeometry()
	s.mu.Lock()
	first := s.streams["first"]
	s.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(tea.MouseClickMsg{X: g.x + 3, Y: g.y + 2, Button: tea.MouseRight, Mod: tea.ModShift})
	h.until(func() bool { return len(h.m.workspace.inputs) == 0 })
	for range workspaceInputCapacity - 1 {
		h.send(tea.MouseMotionMsg{X: g.x + 8, Y: g.y + 2, Button: tea.MouseRight, Mod: tea.ModAlt})
	}
	h.send(tea.MouseMotionMsg{X: g.x + 9, Y: g.y + 2, Button: tea.MouseRight, Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: 'z', Text: "z"})
	close(gate)
	h.until(func() bool {
		return h.m.workspace.active == 1 && h.m.workspace.conn != nil && h.m.workspace.pending == 0
	})
	got := directDragWrites(s, "first")
	if strings.Count(got, "mouse:") != workspaceInputCapacity+1 || !strings.HasSuffix(got, "mouse:up,right,2,8,2") || strings.Contains(got, "mouse:drag,right,4,9,2") {
		t.Fatalf("reserve lost accepted modifier/coordinates: %q", got)
	}
	if directDragWrites(s, "second") != "text:z" {
		t.Fatal("mouse gesture crossed Run")
	}
}
