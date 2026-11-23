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

type renameFixture struct {
	workspaceFixture
	names     map[string]session.Session
	renamed   []session.Session
	renameErr error
}

func TestSessionRenameCancelKeepsPrefilledNameAndRealCursorAtEnd(t *testing.T) {
	service := newRenameFixture()
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "重命名会话") })
	h.send(tea.KeyPressMsg{Code: '!', Text: "!"})
	if !strings.Contains(ansi.Strip(h.m.View().Content), "> 开发!") {
		t.Fatal("existing name was not prefilled with cursor at end")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.until(func() bool { return h.m.workspace.conn != nil })
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.renamed) != 0 || h.m.workspace.focused {
		t.Fatal("cancel saved or assigned terminal focus")
	}
}

func TestSessionRenameWaitsForAcceptedInputAndRejectsReplacedRun(t *testing.T) {
	service := newRenameFixture()
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	gate := make(chan struct{})
	stream.mu.Lock()
	stream.gate = gate
	stream.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "pending-draft"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if strings.Contains(h.m.View().Content, "重命名会话") {
		t.Fatal("dialog cancelled an accepted write")
	}
	close(gate)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "重命名会话") })
	items := sampleSessions()
	items[0].RunID = "replacement"
	h.send(sessionsMsg{generation: h.m.listGeneration, sessions: items})
	h.send(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	h.send(tea.PasteMsg{Content: "new name"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !strings.Contains(ansi.Strip(h.m.View().Content), "原 Session/Run") {
		t.Fatal("replaced Run was not rejected")
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.renamed) != 0 {
		t.Fatal("replacement retargeted rename")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if strings.Count(strings.Join(stream.writes, ","), "text:pending-draft") != 1 {
		t.Fatal("accepted draft lost or replayed")
	}
}

func TestSessionRenameErrorRequiresReconcileInsteadOfRepeatedSubmit(t *testing.T) {
	service := newRenameFixture()
	service.renameErr = fmt.Errorf("concurrent editor changed name")
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "重命名会话") })
	h.send(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	h.send(tea.PasteMsg{Content: "edited"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "请返回") })
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.renamed) != 1 || h.m.workspace.tabs[h.m.workspace.active].session.ID != "first" {
		t.Fatal("error retried or switched terminal")
	}
}

func TestDuplicateSessionNamesExposeDistinctIdentitiesInTabsListAndSearch(t *testing.T) {
	service := newRenameFixture()
	service.fakeService.list = func(context.Context) ([]session.Session, error) {
		items := sampleSessions()
		items[0].Title = "duplicate"
		items[1].Title = "duplicate"
		return items, nil
	}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	view := ansi.Strip(h.m.View().Content)
	for _, marker := range []string{"[first]", "[second]"} {
		if strings.Count(view, marker) < 2 {
			t.Errorf("Tab/list lacks identity cue %s", marker)
		}
	}
	h.send(tea.KeyPressMsg{Code: '/', Text: "/"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "本地会话搜索") })
	view = ansi.Strip(h.m.View().Content)
	for _, marker := range []string{"[first]", "[second]"} {
		if !strings.Contains(view, marker) {
			t.Errorf("search lacks identity cue %s", marker)
		}
	}
}

func TestSessionRenameFitsSourceGeometryAndBoundedWindows(t *testing.T) {
	service := newRenameFixture()
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "重命名会话") })
	for _, size := range [][2]int{{120, 32}, {63, 25}, {49, 25}, {27, 10}, {10, 4}, {1, 1}, {0, 0}} {
		h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		v := h.m.View()
		if size[0] == 0 {
			if v.Content != "" || v.Cursor != nil {
				t.Fatal("nonempty zero viewport")
			}
			continue
		}
		rows := strings.Split(ansi.Strip(v.Content), "\n")
		if len(rows) != size[1] {
			t.Fatalf("rows %v: %d", size, len(rows))
		}
		for _, r := range rows {
			if ansi.StringWidth(r) != size[0] {
				t.Fatalf("columns %v: %q", size, r)
			}
		}
		if size[0] >= 49 {
			wantWidth, wantTop := 46, 7
			if size[0] == 120 {
				wantTop = 11
			}
			if size[0] == 49 {
				wantWidth = 41
			}
			border := "╭" + strings.Repeat("─", wantWidth-2) + "╮"
			if !strings.Contains(rows[wantTop], border) || !strings.Contains(rows[wantTop+4], "> 开发") || !strings.Contains(rows[wantTop+9], "╰") {
				t.Fatalf("source geometry changed at %v", size)
			}
		}
		if v.Cursor != nil && (v.Cursor.X < 0 || v.Cursor.X >= size[0] || v.Cursor.Y < 0 || v.Cursor.Y >= size[1]) {
			t.Fatal("cursor outside viewport")
		}
	}
}

func TestSessionRenameRejectsMultilinePasteAndKeepsTerminalShortcutsLiteral(t *testing.T) {
	service := newRenameFixture()
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.send(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	h.until(func() bool {
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "text:r,key:C-r")
	})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "重命名会话") })
	h.send(tea.PasteMsg{Content: "should\nnot\nsubmit"})
	if strings.Contains(ansi.Strip(h.m.View().Content), "> should") || !strings.Contains(h.m.View().Content, "单行") {
		t.Fatal("multiline paste was sanitized into a different accepted name")
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.renamed) != 0 {
		t.Fatal("paste submitted rename")
	}
}

func TestManagementRefreshReloadsNamesChangedByAnotherEditor(t *testing.T) {
	service := newRenameFixture()
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	// Suppress the already scheduled periodic poll so it cannot mask a broken
	// explicit refresh by eventually discovering the external rename itself.
	h.m.pollGeneration++
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	if _, err := service.Rename(context.Background(), sampleSessions()[1], "外部修改"); err != nil {
		t.Fatal(err)
	}
	h.send(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "外部修改") })
	if h.m.rename != nil || h.m.workspace.focused {
		t.Fatal("refresh opened rename or changed focus")
	}
}

func TestSessionRenamePreCommitDiscoveryCannotRestoreTheOldDisplayName(t *testing.T) {
	service := newRenameFixture()
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.m.pollGeneration++
	oldGeneration, oldRows := h.m.listGeneration, sampleSessions()
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "重命名会话") })
	h.send(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	h.send(tea.PasteMsg{Content: "已保存新名"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		return strings.Contains(h.m.View().Content, "管理模式") && strings.Contains(h.m.View().Content, "已保存新名")
	})
	// This is a discovery snapshot requested before the name transaction, but
	// delivered after its success. A fresh read is allowed; a visible rollback is not.
	h.send(sessionsMsg{generation: oldGeneration, sessions: oldRows})
	if !strings.Contains(h.m.View().Content, "已保存新名") {
		t.Fatal("late discovery restored the old name")
	}
	h.until(func() bool { return !h.m.listBusy })
	if !strings.Contains(h.m.View().Content, "已保存新名") {
		t.Fatal("fresh discovery lost the saved name")
	}
}

func (s *renameFixture) List(ctx context.Context) ([]session.Session, error) {
	items, err := s.workspaceFixture.List(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, v := range items {
		if named, ok := s.names[v.ID]; ok {
			items[i].Title, items[i].TitleRevision = named.Title, named.TitleRevision
		}
	}
	return items, err
}
func (s *renameFixture) Rename(ctx context.Context, v session.Session, name string) (session.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.renamed = append(s.renamed, v)
	if s.renameErr != nil {
		return session.Session{}, s.renameErr
	}
	if current, ok := s.names[v.ID]; ok && current.TitleRevision != v.TitleRevision {
		return session.Session{}, fmt.Errorf("name changed")
	}
	v.Title, v.TitleRevision = name, v.TitleRevision+1
	if s.names == nil {
		s.names = map[string]session.Session{}
	}
	s.names[v.ID] = v
	return v, nil
}

func newRenameFixture() *renameFixture {
	return &renameFixture{workspaceFixture: workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}}
}

func TestSessionRenameChangesSelectedBackgroundLabelWithoutSwitchingOrTerminalInput(t *testing.T) {
	service := newRenameFixture()
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	first := service.streams["first"]
	service.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "draft-preserved"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "重命名会话") })
	if !strings.Contains(h.m.View().Content, "评审") || h.m.View().Cursor == nil {
		t.Fatal("rename lost current name or editable cursor")
	}
	h.send(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	h.send(tea.PasteMsg{Content: "中文新名 ; 'literal'"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		return strings.Contains(h.m.View().Content, "shell marker first") && strings.Contains(h.m.View().Content, "中文新名")
	})
	if !strings.Contains(ansi.Strip(h.m.View().Content), "管理模式") {
		t.Fatal("rename success hid management focus and reachable navigation hints")
	}
	if h.m.workspace.focused || h.m.workspace.tabs[h.m.workspace.active].session.ID != "first" {
		t.Fatal("renaming selected background session activated it")
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.renamed) != 1 || service.renamed[0].ID != "second" || len(service.starts) != 0 {
		t.Fatalf("wrong rename target: %+v", service.renamed)
	}
	first.mu.Lock()
	defer first.mu.Unlock()
	writes := strings.Join(first.writes, ",")
	if strings.Count(writes, "text:draft-preserved") != 1 || strings.Contains(writes, "literal") || strings.Contains(writes, "key:Enter") {
		t.Fatalf("rename leaked/lost terminal input: %s", writes)
	}
}
