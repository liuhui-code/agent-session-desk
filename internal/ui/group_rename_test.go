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

type groupRenameFixture struct {
	*groupedFixture
	renameCalls int
	renameErr   error
}

func TestGroupRenameAcknowledgementDoesNotRecreateAnObservedRemovedGroup(t *testing.T) {
	s := &groupRenameFixture{groupedFixture: newGroupedFixture()}
	root, err := s.CreateGroup(context.Background(), "", "project", "")
	if err != nil {
		t.Fatal(err)
	}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '2', Text: "2"})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return h.m.groupDialog != nil })
	dialog := h.m.groupDialog
	named := root
	named.Name = "renamed"
	named.Revision++
	s.groupMu.Lock()
	retained := []session.Group{s.groups[0]}
	s.groups = append([]session.Group(nil), retained...)
	s.groupMu.Unlock()
	h.send(sessionsMsg{generation: h.m.listGeneration, sessions: sampleSessions(), groups: retained})
	h.send(groupResult{dialog: dialog, group: named})
	if h.m.groupByID(root.ID).ID != "" {
		t.Fatal("late rename receipt recreated an observed removed group")
	}
}

func (s *groupRenameFixture) RenameGroup(ctx context.Context, expected session.Group, name string) (session.Group, error) {
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	s.renameCalls++
	if s.renameErr != nil {
		return session.Group{}, s.renameErr
	}
	for i, g := range s.groups {
		if g.ID == expected.ID && g == expected {
			g.Name, g.Revision = name, g.Revision+1
			s.groups[i] = g
			return g, nil
		}
	}
	return session.Group{}, fmt.Errorf("stale group")
}

func TestGroupRenameUsesNameOnlyAndKeepsFoldedGroupOriginalTabAndDraft(t *testing.T) {
	s := &groupRenameFixture{groupedFixture: newGroupedFixture()}
	root, err := s.CreateGroup(context.Background(), "", "项目", `C:\work`)
	if err != nil {
		t.Fatal(err)
	}
	child, err := s.CreateGroup(context.Background(), root.ID, "评审", "")
	if err != nil {
		t.Fatal(err)
	}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "draft-preserved"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyHome})
	for i := 0; i < 3; i++ {
		h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	if h.m.selectedGroupID != root.ID || h.m.cursor != -1 {
		t.Fatal("did not select root header")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.until(func() bool { return h.m.groupCollapsed(root.ID) })
	oldGeneration := h.m.listGeneration
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "重命名分组") })
	if strings.Contains(h.m.View().Content, "默认目录") {
		t.Fatal("rename exposed hidden directory field")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: " 新版"})
	if !strings.Contains(ansi.Strip(h.m.View().Content), "> 项目 新版") {
		t.Fatal("name not prefilled/cursor not at end or Tab changed hidden focus")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		return h.m.groupDialog == nil && h.m.workspace.conn != nil && h.m.groupByID(root.ID).Name == "项目 新版"
	})
	if !h.m.groupCollapsed(root.ID) || h.m.selectedGroupID != root.ID || h.m.groupPath(child.ID) != "项目 新版/评审" {
		t.Fatal("rename lost fold/selection/dynamic child path")
	}
	h.send(sessionsMsg{generation: oldGeneration, sessions: sampleSessions(), groups: []session.Group{{ID: session.DefaultGroupID, Name: "我的会话", Revision: 1}, root, child}})
	if h.m.groupByID(root.ID).Name != "项目 新版" || len(h.m.workspace.tabs) != 2 || !sameRun(h.m.workspace.tabs[0].session, sampleSessions()[0]) {
		t.Fatal("late discovery rolled back rename or replaced original Run")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	writes := strings.Join(stream.writes, ",")
	if strings.Count(writes, "text:draft-preserved") != 1 || strings.Contains(writes, "新版") || strings.Contains(writes, "text:r") {
		t.Fatalf("management text leaked or draft replayed: %s", writes)
	}
}

func TestGroupRenameAcknowledgementCannotRollBackANewerObservedName(t *testing.T) {
	s := &groupRenameFixture{groupedFixture: newGroupedFixture()}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyHome})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return h.m.groupDialog != nil })
	dialog := h.m.groupDialog
	old := h.m.groupByID(session.DefaultGroupID)
	committed := old
	committed.Name, committed.Revision = "first edit", old.Revision+1
	newer := committed
	newer.Name, newer.Revision = "second edit", committed.Revision+1
	s.groupMu.Lock()
	s.groups[0] = newer
	s.groupMu.Unlock()
	h.send(sessionsMsg{generation: h.m.listGeneration, sessions: sampleSessions(), groups: []session.Group{newer}})
	h.send(groupResult{dialog: dialog, group: committed})
	if h.m.groupByID(old.ID) != newer {
		t.Fatal("earlier acknowledgement rolled back newer observed metadata")
	}
}

func TestGroupRenameWaitsForAcceptedTerminalInputAndCancelNeverSaves(t *testing.T) {
	s := &groupRenameFixture{groupedFixture: newGroupedFixture()}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	gate := make(chan struct{})
	defer func() {
		select {
		case <-gate:
		default:
			close(gate)
		}
	}()
	stream.mu.Lock()
	stream.gate = gate
	stream.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r-terminal-draft"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyHome})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if h.m.groupDialog != nil {
		t.Fatal("dialog cancelled an accepted terminal write")
	}
	close(gate)
	h.until(func() bool { return h.m.groupDialog != nil })
	h.send(tea.PasteMsg{Content: "invalid\nmulti-line"})
	if !strings.Contains(h.m.View().Content, "有界单行") {
		t.Fatal("multiline paste was not rejected")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEsc})
	h.until(func() bool { return h.m.groupDialog == nil && h.m.workspace.conn != nil })
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if s.renameCalls != 0 || s.groups[0].Name != "我的会话" {
		t.Fatal("cancel changed metadata")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	var inputWrites []string
	for _, write := range stream.writes {
		if !strings.HasPrefix(write, "size:") {
			inputWrites = append(inputWrites, write)
		}
	}
	if strings.Join(inputWrites, ",") != "text:r-terminal-draft" {
		t.Fatalf("terminal input lost/replayed or modal paste leaked: %v", stream.writes)
	}
}

func TestGroupRenameFailureCannotResubmitOrSwitchTabsAndOldResultCannotCloseNewDialog(t *testing.T) {
	s := &groupRenameFixture{groupedFixture: newGroupedFixture(), renameErr: fmt.Errorf("synthetic conflicting editor")}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyHome})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return h.m.groupDialog != nil })
	old := h.m.groupDialog
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.send(tea.KeyPressMsg{Code: tea.KeyEsc})
	if h.m.groupDialog != old {
		t.Fatal("busy commit was abandoned")
	}
	h.until(func() bool { return strings.Contains(h.m.View().Content, "请返回并刷新") })
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	s.groupMu.Lock()
	calls := s.renameCalls
	s.groupMu.Unlock()
	if calls != 1 || h.m.workspace.tabs[h.m.workspace.active].session.ID != "first" {
		t.Fatal("failed submit retried or switched terminal")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEsc})
	h.until(func() bool { return h.m.groupDialog == nil && h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return h.m.groupDialog != nil })
	current := h.m.groupDialog
	h.send(groupResult{dialog: old, group: old.groupTarget})
	if h.m.groupDialog != current {
		t.Fatal("stale result closed a different editor")
	}
}

func TestGroupRenameFitsSourceGeometryAndKeepsOnlyNameFocusedAfterResize(t *testing.T) {
	s := &groupRenameFixture{groupedFixture: newGroupedFixture()}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyHome})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return h.m.groupDialog != nil })
	for _, size := range [][2]int{{120, 32}, {63, 25}, {49, 25}, {27, 10}, {10, 4}, {1, 1}, {0, 0}} {
		h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		view := h.m.View()
		if size[0] == 0 {
			if view.Content != "" || view.Cursor != nil {
				t.Fatal("nonempty zero viewport")
			}
			continue
		}
		rows := strings.Split(ansi.Strip(view.Content), "\n")
		if len(rows) != size[1] {
			t.Fatalf("wrong height at %v", size)
		}
		for _, row := range rows {
			if ansi.StringWidth(row) != size[0] {
				t.Fatalf("wrong width at %v: %q", size, row)
			}
		}
		if size[0] >= 49 {
			width, top := 46, 7
			if size[0] == 120 {
				top = 11
			}
			if size[0] == 49 {
				width = 41
			}
			if !strings.Contains(rows[top], "╭"+strings.Repeat("─", width-2)+"╮") || !strings.Contains(rows[top+4], "> 我的会话") || !strings.Contains(rows[top+9], "╰") {
				t.Fatalf("source geometry changed at %v", size)
			}
			if view.Cursor == nil {
				t.Fatalf("real name cursor missing at %v", size)
			}
		}
		if view.Cursor != nil && (view.Cursor.X < 0 || view.Cursor.X >= size[0] || view.Cursor.Y < 0 || view.Cursor.Y >= size[1]) {
			t.Fatalf("cursor outside viewport at %v", size)
		}
		if h.m.groupDialog.focus != 0 {
			t.Fatal("resize changed focused field")
		}
	}
}
