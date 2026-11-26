package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestWorkspaceOuterInsetsRemainWhenSidebarIsHidden(t *testing.T) {
	h := newWorkspaceHarness(t, directDragFixture())
	h.until(func() bool { return h.m.workspace.conn != nil })
	for _, width := range []int{120, 63, 49, 12} {
		h.send(tea.WindowSizeMsg{Width: width, Height: 12})
		h.m.workspace.sidebarHidden = true
		g := h.m.workspaceGeometry()
		if g.x != 2 || g.y != 1 || g.width != width-4 || g.height != 11 {
			t.Errorf("hidden insets %d: %+v", width, g)
		}
		lines := strings.Split(ansi.Strip(h.m.View().Content), "\n")
		if strings.TrimSpace(lines[0]) != "" {
			t.Fatal("top breathing row contains content")
		}
		for _, line := range lines {
			if ansi.StringWidth(line) != width {
				t.Fatal("frame exceeds width")
			}
		}
	}
}

func TestNavigationFooterIsAlwaysVisibleAndReservesItsOwnRow(t *testing.T) {
	h := newWorkspaceHarness(t, directDragFixture())
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	g := h.m.workspaceGeometry()
	if !h.m.workspaceFooterVisible() || g.y+g.height != h.m.height-1 {
		t.Fatalf("management footer overlaps content: %+v", g)
	}
	last := strings.Split(ansi.Strip(h.m.View().Content), "\n")[h.m.height-1]
	if !strings.Contains(last, "F6") || !strings.Contains(last, "Alt+J/K") || !strings.Contains(last, "F2") {
		t.Fatal("management shortcuts are hidden")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	if h.m.workspaceFooterVisible() {
		t.Fatal("ordinary terminal mode shows management footer")
	}
}

func TestF2RenamesActiveTerminalInsteadOfUnfocusedSelection(t *testing.T) {
	s := newRenameFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.m.cursor = 1
	h.send(tea.KeyPressMsg{Code: tea.KeyF2})
	h.until(func() bool { return h.m.rename != nil })
	if h.m.rename.target.ID != "first" {
		t.Fatal("F2 renamed an inactive selection")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
}

func TestAltJKSwitchesSessionsWhileTypingWithoutLeakingKeys(t *testing.T) {
	s := directDragFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	s.mu.Lock()
	first := s.streams["first"]
	s.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "draft-A"})
	h.send(tea.KeyPressMsg{Code: 'j', Mod: tea.ModAlt})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	s.mu.Lock()
	second := s.streams["second"]
	s.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "draft-B"})
	h.send(tea.KeyPressMsg{Code: 'k', Mod: tea.ModAlt})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	for index, stream := range []*workspaceStream{first, second} {
		stream.mu.Lock()
		got := strings.Join(stream.writes, ";")
		stream.mu.Unlock()
		want := []string{"text:draft-A", "text:draft-B"}[index]
		if strings.Contains(got, "alt+") || strings.Count(got, want) != 1 {
			t.Fatalf("original input owner changed: %q", got)
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.starts) != 0 {
		t.Fatal("session switching created a process")
	}
}

func TestInsetListFirstRowClickUsesRenderedCoordinates(t *testing.T) {
	m := New(context.Background(), &workspaceFixture{})
	defer m.Close()
	m.width, m.height = 120, 24
	m.sessions, m.cursor = sampleSessions(), 1
	m.sidebarMode = "flat"
	m.workspace.focused = false
	m.workspaceClick(tea.MouseClickMsg{X: 3, Y: 0, Button: tea.MouseLeft})
	if m.cursor != 1 {
		t.Fatal("top margin selected a hidden row")
	}
	m.workspaceClick(tea.MouseClickMsg{X: 3, Y: 1, Button: tea.MouseLeft})
	if m.cursor != 0 {
		t.Fatal("inset first row did not select first session")
	}
}

func TestF2RenamesSelectedGroupWithoutTerminalInput(t *testing.T) {
	s := &groupRenameFixture{groupedFixture: newGroupedFixture()}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyUp})
	h.send(tea.KeyPressMsg{Code: tea.KeyF2})
	h.until(func() bool { return h.m.groupDialog != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if s.renameCalls != 0 {
		t.Fatal("cancelled F2 editor changed group")
	}
}
