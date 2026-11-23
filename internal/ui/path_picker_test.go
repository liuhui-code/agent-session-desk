package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"strings"
	"testing"
)

func TestGroupPathPickerAcceptsThenDismissesBeforeSubmitting(t *testing.T) {
	s := newGroupedFixture()
	s.groups = append(s.groups, session.Group{ID: "directory-group", Name: "已有组", DefaultDir: "/demo-work", Revision: 1})
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return len(h.m.groups) > 1 })
	h.m.groupIntent = &groupDialog{}
	h.run(h.m.openGroupDialog())
	h.send(tea.PasteMsg{Content: "新分组"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "/demo-work") })
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.groupDialog == nil || h.m.groupDialog.inputs[1].Value() != "/demo-work" || h.m.groupDialog.submitted {
		t.Fatal("accepting candidate submitted or closed the group form")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if h.m.groupDialog != nil {
		t.Fatal("dismissed picker did not permit cancel")
	}
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.groups) != 2 {
		t.Fatal("directory navigation created a group")
	}
}

func TestAsyncRecentsKeepAnExplicitlyNavigatedCandidateIdentity(t *testing.T) {
	gate := make(chan struct{})
	s := newGroupedFixture()
	s.groups = append(s.groups, session.Group{ID: "a", Name: "A", DefaultDir: "/group-a"}, session.Group{ID: "b", Name: "B", DefaultDir: "/group-b"})
	h := newWorkspaceHarness(t, s, Options{LoadRecentDirs: func(ctx context.Context) ([]string, error) {
		select {
		case <-gate:
			return []string{"/recent-first"}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}})
	h.until(func() bool { return len(h.m.groups) == 3 })
	h.m.groupIntent = &groupDialog{}
	h.run(h.m.openGroupDialog())
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	close(gate)
	h.until(func() bool { return h.m.groupDialog.paths.loaded })
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.groupDialog.inputs[1].Value() != "/group-b" {
		t.Fatal("late corpus changed the explicitly highlighted directory")
	}
}
