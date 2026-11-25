package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestGroupDialogsFitTinyViewportsAndKeepTheFocusedFieldVisible(t *testing.T) {
	s := newGroupedFixture()
	for i := 0; i < 20; i++ {
		_, _ = s.CreateGroup(context.Background(), "", fmt.Sprintf("group-%02d", i), "")
	}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'g', Text: "g"})
	h.until(func() bool { return h.m.groupDialog != nil })
	if plain := ansi.Strip(h.m.View().Content); !strings.Contains(plain, "Group name") || !strings.Contains(plain, "Optional working directory") {
		t.Fatal("source form placeholders are missing")
	}
	h.send(tea.PasteMsg{Content: "项目"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: `C:\项目`})
	for _, size := range [][2]int{{120, 32}, {63, 25}, {49, 25}, {27, 8}, {21, 6}, {9, 3}, {1, 1}} {
		h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		view := h.m.View()
		plain := ansi.Strip(view.Content)
		lines := strings.Split(plain, "\n")
		if len(lines) != size[1] {
			t.Fatalf("wrong height %v: %d", size, len(lines))
		}
		for _, line := range lines {
			if ansi.StringWidth(line) != size[0] {
				t.Fatalf("wrong width %v: %q", size, line)
			}
		}
		if view.Cursor != nil && (view.Cursor.X < 0 || view.Cursor.X >= size[0] || view.Cursor.Y < 0 || view.Cursor.Y >= size[1]) {
			t.Fatalf("cursor outside viewport %v: %+v", size, view.Cursor)
		}
		if size[0] >= 27 && !strings.Contains(plain, "Default directory:") {
			t.Fatalf("focused directory hidden %v:\n%s", size, plain)
		}
	}
	h.send(tea.WindowSizeMsg{Width: 120, Height: 32})
	h.send(tea.KeyPressMsg{Code: tea.KeyEsc})
	h.until(func() bool { return h.m.groupDialog == nil && h.m.workspace.conn != nil })
	h.send(tea.WindowSizeMsg{Width: 49, Height: 12})
	h.send(tea.KeyPressMsg{Code: 'M', Text: "M"})
	h.until(func() bool { return h.m.groupDialog != nil })
	for i := 0; i < 20; i++ {
		h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	if !strings.Contains(ansi.Strip(h.m.View().Content), "group-19") {
		t.Fatal("move list selected row scrolled out of view")
	}
}

func TestGroupMoveRejectsAnOriginalRunReplacedWhileTheDialogWasOpen(t *testing.T) {
	s := newGroupedFixture()
	root, _ := s.CreateGroup(context.Background(), "", "project", "")
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'M', Text: "M"})
	h.until(func() bool { return h.m.groupDialog != nil })
	replaced := sampleSessions()
	replaced[0].RunID = "replacement-run"
	h.send(sessionsMsg{generation: h.m.listGeneration, sessions: replaced, groups: s.groups})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !strings.Contains(h.m.View().Content, "原 Session/Run 已变化") {
		t.Fatal("replacement was not rejected")
	}
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.members) != 0 || h.m.groupDialog.target.GroupID == root.ID {
		t.Fatal("old editor moved a replacement Run")
	}
}

func TestGroupMoveSuccessRejectsPreCommitDiscoveryAndDoesNotDuplicateTheTab(t *testing.T) {
	s := newGroupedFixture()
	root, _ := s.CreateGroup(context.Background(), "", "project", "")
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	oldGeneration := h.m.listGeneration
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'M', Text: "M"})
	h.until(func() bool { return h.m.groupDialog != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return h.m.groupDialog == nil && h.m.sessions[0].GroupID == root.ID })
	h.send(sessionsMsg{generation: oldGeneration, sessions: sampleSessions(), groups: []session.Group{{ID: session.DefaultGroupID, Name: "My Sessions", Revision: 1}}})
	if h.m.sessions[0].GroupID != root.ID || len(h.m.sidebarGroups()) != 2 || len(h.m.workspace.tabs) != 2 {
		t.Fatal("late snapshot restored old group or duplicated Tab")
	}
}
