package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type groupedFixture struct {
	workspaceFixture
	groupMu sync.Mutex
	groups  []session.Group
	members map[string]session.Session
}

func newGroupedFixture() *groupedFixture {
	return &groupedFixture{workspaceFixture: workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}, groups: []session.Group{{ID: session.DefaultGroupID, Name: "My Sessions", Revision: 1}}, members: map[string]session.Session{}}
}

func (s *groupedFixture) Workspace(ctx context.Context) (session.Workspace, error) {
	items, err := s.List(ctx)
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	for i, v := range items {
		if saved, ok := s.members[v.ID]; ok {
			items[i].GroupID, items[i].GroupRevision = saved.GroupID, saved.GroupRevision
		}
	}
	return session.Workspace{Groups: append([]session.Group(nil), s.groups...), Sessions: items}, err
}
func (s *groupedFixture) CreateGroup(ctx context.Context, parent, name, dir string) (session.Group, error) {
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	name, err := session.NormalizeGroupName(name)
	if err != nil {
		return session.Group{}, err
	}
	g := session.Group{ID: fmt.Sprintf("%032x", len(s.groups)), ParentID: parent, Name: name, DefaultDir: dir, Revision: 1, Order: int64(len(s.groups))}
	s.groups = append(s.groups, g)
	return g, nil
}
func (s *groupedFixture) MoveToGroup(ctx context.Context, v session.Session, target string) (session.Session, error) {
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	current := s.members[v.ID]
	if current.GroupRevision != v.GroupRevision {
		return session.Session{}, fmt.Errorf("stale membership")
	}
	v.GroupID, v.GroupRevision = target, v.GroupRevision+1
	if target == session.DefaultGroupID {
		v.GroupID = ""
	}
	s.members[v.ID] = v
	return v, nil
}

func TestGroupedTreeRendersNestedEmptyGroupsAndUsesTheirOwnMouseRows(t *testing.T) {
	s := newGroupedFixture()
	root, _ := s.CreateGroup(context.Background(), "", "项目", "")
	child, _ := s.CreateGroup(context.Background(), root.ID, "审核", "")
	_, _ = s.MoveToGroup(context.Background(), sampleSessions()[1], child.ID)
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	view := ansi.Strip(h.m.View().Content)
	if !strings.Contains(view, "▾ 项目 (0)") || !strings.Contains(view, "▾ 审核 (1)") || !strings.Contains(view, "▾ My Sessions (1)") {
		t.Fatalf("tree/members missing:\n%s", view)
	}
	for y, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "▾ 项目") {
			h.send(tea.MouseClickMsg{Button: tea.MouseLeft, X: 2, Y: y})
			break
		}
	}
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "▸ 项目") })
	if strings.Contains(ansi.Strip(h.m.View().Content), "审核") {
		t.Fatal("collapsed parent left nested header visible")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "▾ 审核") })
}

func TestGroupCreateAndMoveDialogsKeepOriginalTabAndInput(t *testing.T) {
	s := newGroupedFixture()
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	s.mu.Lock()
	original := s.streams["first"]
	s.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "draft-preserved"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'g', Text: "g"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Create group") })
	h.send(tea.PasteMsg{Content: "设计"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: `C:\work\项目`})
	h.send(tea.KeyPressMsg{Code: tea.KeyUp})
	if !strings.Contains(h.m.View().Content, "Create subgroup") {
		t.Fatal("Up did not switch root/subgroup context")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "▾ 设计 (0)") })
	s.groupMu.Lock()
	created := s.groups[1]
	s.groupMu.Unlock()
	if created.ParentID != session.DefaultGroupID || created.DefaultDir != `C:\work\项目` {
		t.Fatalf("lost create context: %+v", created)
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyHome})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: 'M', Text: "M"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Move to group") })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		return strings.Contains(ansi.Strip(h.m.View().Content), "▾ 设计 (1)") && strings.Contains(h.m.View().Content, "shell marker first")
	})
	if len(h.m.workspace.tabs) != 2 || !sameRun(h.m.workspace.tabs[0].session, sampleSessions()[0]) {
		t.Fatal("move replaced original Tab/Run")
	}
	original.mu.Lock()
	writes := strings.Join(original.writes, ",")
	original.mu.Unlock()
	if strings.Count(writes, "text:draft-preserved") != 1 || strings.Contains(writes, "设计") || strings.Contains(writes, "text:g") || strings.Contains(writes, "text:M") {
		t.Fatalf("modal leaked or lost input: %s", writes)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.starts) != 0 {
		t.Fatal("group operation started a terminal")
	}
}

func TestNewTerminalUsesSelectedGroupsDirectoryAndStableIdentity(t *testing.T) {
	s := newGroupedFixture()
	g, err := s.CreateGroup(context.Background(), "", "project", `C:\work\project`)
	if err != nil {
		t.Fatal(err)
	}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '2', Text: "2"})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Create terminal") })
	if !strings.Contains(h.m.View().Content, g.DefaultDir) {
		t.Fatal("selected group directory was not prefilled")
	}
	h.send(tea.PasteMsg{Content: "shell in project"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: "pwsh.exe"})
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.until(func() bool { s.mu.Lock(); defer s.mu.Unlock(); return len(s.starts) == 1 })
	s.mu.Lock()
	launch := s.starts[0]
	s.mu.Unlock()
	if launch.GroupID != g.ID || launch.Dir != g.DefaultDir {
		t.Fatalf("lost selected launch context: %+v", launch)
	}
}
