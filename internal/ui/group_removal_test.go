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

func TestGroupRemovalNeverConfirmsWhenTheCompleteReviewDoesNotFit(t *testing.T) {
	s, _, _, _ := newGroupRemovalFixture(t)
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '2', Text: "2"})
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "1 sessions") })
	for _, size := range [][2]int{{120, 10}, {63, 10}, {49, 8}, {27, 8}, {9, 3}, {1, 1}, {0, 0}} {
		h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		v := h.m.View()
		if size[0] == 0 {
			if v.Content != "" || v.Cursor != nil {
				t.Fatal("nonempty zero viewport")
			}
		} else {
			lines := strings.Split(ansi.Strip(v.Content), "\n")
			if len(lines) != size[1] {
				t.Fatalf("wrong height at %v", size)
			}
			for _, line := range lines {
				if ansi.StringWidth(line) != size[0] {
					t.Fatalf("wrong width at %v: %q", size, line)
				}
			}
		}
		h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
		h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
		if h.m.groupRemoval == nil || h.m.groupRemoval.busy || h.m.groupRemoval.submitted {
			t.Fatalf("unreadable impact was confirmed at %v", size)
		}
	}
	h.send(tea.WindowSizeMsg{Width: 120, Height: 32})
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.until(func() bool { return h.m.groupRemoval == nil && len(h.m.sidebarGroups()) == 2 })
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.removals) != 1 {
		t.Fatal("blocked keys were deferred/replayed")
	}
}

type groupRemovalFixture struct {
	*groupedFixture
	plan                    session.GroupRemoval
	result                  session.Workspace
	previewErr, removeErr   error
	previewGate, removeGate <-chan struct{}
	previews                []string
	removals                []session.GroupRemoval
	observed                [][]session.Session
}

func (s *groupRemovalFixture) PreviewGroupRemoval(ctx context.Context, id string) (session.GroupRemoval, error) {
	s.groupMu.Lock()
	s.previews = append(s.previews, id)
	plan, err, gate := s.plan, s.previewErr, s.previewGate
	s.groupMu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return session.GroupRemoval{}, ctx.Err()
		}
	}
	if err != nil {
		return session.GroupRemoval{}, err
	}
	if id != plan.Root.ID {
		return session.GroupRemoval{}, fmt.Errorf("unknown group")
	}
	return plan, nil
}
func (s *groupRemovalFixture) RemoveGroup(ctx context.Context, expected session.GroupRemoval, observed []session.Session) (session.Workspace, error) {
	s.groupMu.Lock()
	s.removals = append(s.removals, expected)
	s.observed = append(s.observed, append([]session.Session(nil), observed...))
	view, err, gate := s.result, s.removeErr, s.removeGate
	s.groupMu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return session.Workspace{}, ctx.Err()
		}
	}
	if err != nil {
		return session.Workspace{}, err
	}
	s.groupMu.Lock()
	s.groups = append([]session.Group(nil), view.Groups...)
	for _, v := range view.Sessions {
		s.members[v.ID] = v
	}
	s.groupMu.Unlock()
	return view, nil
}
func newGroupRemovalFixture(t *testing.T) (*groupRemovalFixture, session.Group, session.Group, session.Group) {
	t.Helper()
	s := &groupRemovalFixture{groupedFixture: newGroupedFixture()}
	root, err := s.CreateGroup(context.Background(), "", "项目", "")
	if err != nil {
		t.Fatal(err)
	}
	child, err := s.CreateGroup(context.Background(), root.ID, "审核", "")
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.CreateGroup(context.Background(), "", "保留", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.MoveToGroup(context.Background(), sampleSessions()[0], child.ID)
	if err != nil {
		t.Fatal(err)
	}
	items := sampleSessions()
	items[0].GroupRevision = 2
	s.plan = session.GroupRemoval{Root: root, Destination: s.groups[0], Groups: []session.Group{root, child}, MemberCount: 1, Stamp: strings.Repeat("a", 64)}
	s.result = session.Workspace{Groups: []session.Group{s.groups[0], other}, Sessions: items}
	return s, root, child, other
}

func TestGroupRemovalDefaultsToCancelThenCollectsMembersWithoutChangingOriginalTabsOrInput(t *testing.T) {
	s, root, _, other := newGroupRemovalFixture(t)
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	s.mu.Lock()
	original := s.streams["first"]
	s.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "original-draft"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '2', Text: "2"})
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	h.until(func() bool {
		return strings.Contains(h.m.View().Content, "2 groups") && strings.Contains(h.m.View().Content, "1 sessions")
	})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		return !strings.Contains(h.m.View().Content, "Delete group?") && h.m.workspace.conn != nil
	})
	s.groupMu.Lock()
	cancelled := len(s.removals)
	s.groupMu.Unlock()
	if cancelled != 0 || h.m.groupByID(root.ID).ID == "" {
		t.Fatal("initial Enter removed a group instead of cancelling")
	}
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "1 sessions") })
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.until(func() bool {
		return len(h.m.sidebarGroups()) == 2 && h.m.workspace.conn != nil && h.m.sessions[0].GroupRevision == 2
	})
	if h.m.groupByID(root.ID).ID != "" || h.m.groupByID(other.ID) != other || h.m.selectedGroupID != session.DefaultGroupID || h.m.cursor != -1 || len(h.m.workspace.tabs) != 2 || !sameRun(h.m.workspace.tabs[0].session, sampleSessions()[0]) || h.m.workspace.focused {
		t.Fatal("removal changed unrelated group/selection/original Tab or assigned terminal focus")
	}
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.removals) != 1 || s.removals[0].Root.ID != root.ID || len(s.observed[0]) != 2 {
		t.Fatal("confirmation targeted a different container/projection")
	}
	original.mu.Lock()
	defer original.mu.Unlock()
	writes := strings.Join(original.writes, ",")
	if strings.Count(writes, "text:original-draft") != 1 || strings.Contains(writes, "text:y") || strings.Contains(writes, "text:d,") {
		t.Fatalf("draft replayed or management keys leaked: %s", writes)
	}
}
