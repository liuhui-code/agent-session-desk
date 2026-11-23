package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func removalFoldList(ids ...string) string { sort.Strings(ids); return strings.Join(ids, ",") }
func openReviewedRemoval(h *workspaceHarness) {
	h.send(tea.KeyPressMsg{Code: '2', Text: "2"})
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	h.until(func() bool { return h.m.groupRemoval != nil && !h.m.groupRemoval.loading })
}

func TestGroupRemovalCleansOnlyDeletedFoldIdsAndPreservesUnknownScopePreferences(t *testing.T) {
	s, root, child, other := newGroupRemovalFixture(t)
	unknown := strings.Repeat("e", 32)
	saves := make(chan config.UISettings, 2)
	h := newWorkspaceHarness(t, s, Options{CollapsedGroups: removalFoldList(root.ID, child.ID, other.ID, unknown), SaveUISettings: func(_ context.Context, value config.UISettings) error { saves <- value; return nil }})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	openReviewedRemoval(h)
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.until(func() bool { return h.m.groupRemoval == nil && !h.m.preferences.busy && len(h.m.sidebarGroups()) == 2 })
	value := <-saves
	if value.CollapsedGroups != removalFoldList(other.ID, unknown) || h.m.groupCollapsed(root.ID) || h.m.groupCollapsed(child.ID) || !h.m.groupCollapsed(other.ID) {
		t.Fatalf("cleanup changed unrelated folds: %+v", value)
	}
}

func TestGroupRemovalCleanupDoesNotPersistAnotherTabsTemporaryExpansion(t *testing.T) {
	s, root, _, other := newGroupRemovalFixture(t)
	v, err := s.MoveToGroup(context.Background(), sampleSessions()[1], other.ID)
	if err != nil {
		t.Fatal(err)
	}
	s.result.Sessions[1] = v
	saves := make(chan config.UISettings, 2)
	h := newWorkspaceHarness(t, s, Options{CollapsedGroups: removalFoldList(root.ID, other.ID), SaveUISettings: func(_ context.Context, value config.UISettings) error { saves <- value; return nil }})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.until(func() bool {
		return h.m.workspace.tabs[h.m.workspace.active].session.ID == "second" && h.m.workspace.conn != nil && !h.m.groupCollapsed(other.ID)
	})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	openReviewedRemoval(h)
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.until(func() bool { return h.m.groupRemoval == nil && !h.m.preferences.busy && len(h.m.sidebarGroups()) == 2 })
	value := <-saves
	if value.CollapsedGroups != other.ID {
		t.Fatalf("cleanup persisted unrelated temporary reveal: %+v", value)
	}
	if h.m.groupCollapsed(other.ID) || h.m.workspace.tabs[h.m.workspace.active].session.ID != "second" {
		t.Fatal("cleanup reversed temporary presentation or switched the active Run")
	}
}

func TestGroupRemovalPreferenceFailureDoesNotRestoreDeletedFoldsOrRetryTheRemoval(t *testing.T) {
	s, root, child, other := newGroupRemovalFixture(t)
	unknown := strings.Repeat("e", 32)
	gate := make(chan struct{})
	t.Cleanup(func() { releaseGroupOrderGate(gate) })
	var calls atomic.Int32
	saves := make(chan config.UISettings, 2)
	h := newWorkspaceHarness(t, s, Options{CollapsedGroups: removalFoldList(root.ID, child.ID, other.ID, unknown), SaveUISettings: func(ctx context.Context, value config.UISettings) error {
		if calls.Add(1) == 1 {
			select {
			case <-gate:
				return fmt.Errorf("synthetic old preference writer failure")
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		saves <- value
		return nil
	}})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '3', Text: "3"})
	h.send(tea.KeyPressMsg{Code: tea.KeyRight})
	if !h.m.preferences.busy {
		t.Fatal("old preference save not pending")
	}
	openReviewedRemoval(h)
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.until(func() bool { return h.m.groupRemoval == nil && len(h.m.sidebarGroups()) == 2 })
	close(gate)
	h.until(func() bool { return !h.m.preferences.busy && h.m.sidebarPreferenceError != nil })
	if strings.Contains(h.m.preferences.wanted.CollapsedGroups, root.ID) || strings.Contains(h.m.preferences.wanted.CollapsedGroups, child.ID) {
		t.Fatal("old writer failure restored deleted group fold IDs")
	}
	if calls.Load() != 1 {
		t.Fatal("failed preference publication retried automatically")
	}
	h.send(tea.KeyPressMsg{Code: 'v', Mod: tea.ModAlt})
	h.until(func() bool { return !h.m.preferences.busy && h.m.sidebarMode == "flat" })
	value := <-saves
	if value.CollapsedGroups != removalFoldList(other.ID, unknown) {
		t.Fatalf("later explicit preference failed to clean only deleted folds: %+v", value)
	}
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.removals) != 1 || session.GroupOf(h.m.sessions[0]) != session.DefaultGroupID {
		t.Fatal("preference failure repeated/undid metadata removal")
	}
}
