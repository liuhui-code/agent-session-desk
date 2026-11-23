package ui

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestGroupOrderFailureCancelsQuitAndRequiresAFreshSnapshotBeforeAnotherEdit(t *testing.T) {
	s, _, _, _ := newGroupOrderFixture(t)
	s.orderErr = fmt.Errorf("synthetic uncertain commit")
	writeGate, readGate := make(chan struct{}), make(chan struct{})
	for _, gate := range []chan struct{}{writeGate, readGate} {
		t.Cleanup(func() {
			select {
			case <-gate:
			default:
				close(gate)
			}
		})
	}
	s.gate = writeGate
	var blockReads atomic.Bool
	s.fakeService.list = func(ctx context.Context) ([]session.Session, error) {
		if blockReads.Load() {
			select {
			case <-readGate:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return sampleSessions(), nil
	}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	blockReads.Store(true)
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	close(writeGate)
	h.until(func() bool {
		return h.m.groupOrder == nil && strings.Contains(h.m.View().Content, "分组顺序未确认")
	})
	if h.m.closed {
		t.Fatal("uncertain commit silently quit")
	}
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	if h.m.groupOrder != nil {
		t.Fatal("ambiguous edit was accepted again before fresh metadata")
	}
	close(readGate)
	h.until(func() bool { return !h.m.listBusy })
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.calls) != 1 {
		t.Fatalf("failed operation retried: %d", len(s.calls))
	}
}

type groupOrderCall struct {
	expected session.GroupOrder
	delta    int
}
type groupOrderFixture struct {
	*groupedFixture
	result   []session.Group
	orderErr error
	gate     <-chan struct{}
	calls    []groupOrderCall
}

func (s *groupOrderFixture) ReorderGroup(ctx context.Context, expected session.GroupOrder, delta int) ([]session.Group, error) {
	s.groupMu.Lock()
	expected.Siblings = append([]session.Group(nil), expected.Siblings...)
	s.calls = append(s.calls, groupOrderCall{expected, delta})
	result, err, gate := append([]session.Group(nil), s.result...), s.orderErr, s.gate
	s.groupMu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if err != nil {
		return nil, err
	}
	s.groupMu.Lock()
	s.groups = append([]session.Group(nil), result...)
	s.groupMu.Unlock()
	return result, nil
}
func newGroupOrderFixture(t *testing.T) (*groupOrderFixture, session.Group, session.Group, session.Group) {
	t.Helper()
	s := &groupOrderFixture{groupedFixture: newGroupedFixture()}
	first, err := s.CreateGroup(context.Background(), "", "first-root", "")
	if err != nil {
		t.Fatal(err)
	}
	child, err := s.CreateGroup(context.Background(), first.ID, "child", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.CreateGroup(context.Background(), "", "second-root", "")
	if err != nil {
		t.Fatal(err)
	}
	a, b := first, second
	a.Order, a.Revision = 2, 2
	b.Order, b.Revision = 1, 2
	s.result = []session.Group{s.groups[0], b, a, child}
	return s, first, child, second
}

func TestGroupOrderMovesAmongSiblingsAndKeepsFoldedGroupTabAndTerminalDraft(t *testing.T) {
	s, first, child, second := newGroupOrderFixture(t)
	h := newWorkspaceHarness(t, s, Options{CollapsedGroups: first.ID})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	connection := h.m.workspace.conn
	h.send(tea.KeyPressMsg{Code: 'd', Text: "terminal-draft"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	if h.m.selectedGroupID != second.ID {
		t.Fatal("did not select last root header")
	}
	oldGeneration := h.m.listGeneration
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.until(func() bool {
		view := ansi.Strip(h.m.View().Content)
		return strings.Index(view, "second-root") < strings.Index(view, "first-root")
	})
	if h.m.selectedGroupID != second.ID || h.m.cursor != -1 || !h.m.groupCollapsed(first.ID) || h.m.groupByID(child.ID) != child || h.m.workspace.conn != connection || len(h.m.workspace.tabs) != 2 || !sameRun(h.m.workspace.tabs[0].session, sampleSessions()[0]) {
		t.Fatal("reorder changed selected identity, fold, subtree or original terminal")
	}
	h.send(sessionsMsg{generation: oldGeneration, sessions: sampleSessions(), groups: []session.Group{s.groups[0], first, child, second}})
	if h.m.groupByID(second.ID).Order != 1 {
		t.Fatal("late discovery restored old order")
	}
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.calls) != 1 || s.calls[0].delta != -1 || s.calls[0].expected.TargetID != second.ID || len(s.calls[0].expected.Siblings) != 3 {
		t.Fatalf("wrong stable target/sibling snapshot: %+v", s.calls)
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	writes := strings.Join(stream.writes, ",")
	if strings.Count(writes, "text:terminal-draft") != 1 || strings.Contains(writes, "text:K") {
		t.Fatalf("draft lost/replayed or sorting key leaked: %s", writes)
	}
}

func TestGroupOrderSourceAliasesOnlyReorderSelectedGroupHeaders(t *testing.T) {
	for _, test := range []struct {
		name  string
		key   tea.KeyPressMsg
		delta int
	}{
		{"K", tea.KeyPressMsg{Code: 'K', Text: "K"}, -1},
		{"plus", tea.KeyPressMsg{Code: '+', Text: "+"}, -1},
		{"ctrl-up", tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModCtrl}, -1},
		{"shift-up", tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModShift}, -1},
		{"J", tea.KeyPressMsg{Code: 'J', Text: "J"}, 1},
		{"minus", tea.KeyPressMsg{Code: '-', Text: "-"}, 1},
		{"ctrl-down", tea.KeyPressMsg{Code: tea.KeyDown, Mod: tea.ModCtrl}, 1},
		{"shift-down", tea.KeyPressMsg{Code: tea.KeyDown, Mod: tea.ModShift}, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			s, first, _, second := newGroupOrderFixture(t)
			h := newWorkspaceHarness(t, s, Options{})
			h.until(func() bool { return h.m.workspace.conn != nil })
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			// The reserved default and Session rows are not reorder targets.
			h.send(tea.KeyPressMsg{Code: tea.KeyHome})
			h.send(test.key)
			h.send(tea.KeyPressMsg{Code: tea.KeyDown})
			h.send(test.key)
			if h.m.groupOrder != nil {
				t.Fatal("default/Session row accepted a group reorder")
			}
			target := second.ID
			if test.delta < 0 {
				h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
			} else {
				target = first.ID
				h.send(tea.KeyPressMsg{Code: '2', Text: "2"})
			}
			h.send(test.key)
			h.until(func() bool { return h.m.groupOrder == nil && h.m.groupByID(second.ID).Order == 1 })
			s.groupMu.Lock()
			defer s.groupMu.Unlock()
			if len(s.calls) != 1 || s.calls[0].delta != test.delta || s.calls[0].expected.TargetID != target {
				t.Fatalf("alias changed wrong target: %+v", s.calls)
			}
			if h.m.selectedGroupID != target {
				t.Fatal("selection followed position instead of stable group ID")
			}
		})
	}
}

func TestGroupOrderTerminalAliasesRemainOriginalInput(t *testing.T) {
	s, _, _, _ := newGroupOrderFixture(t)
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	for _, key := range []tea.KeyPressMsg{{Code: 'K', Text: "K"}, {Code: 'J', Text: "J"}, {Code: '+', Text: "+"}, {Code: '-', Text: "-"}, {Code: tea.KeyUp, Mod: tea.ModCtrl}, {Code: tea.KeyDown, Mod: tea.ModCtrl}, {Code: tea.KeyUp, Mod: tea.ModShift}, {Code: tea.KeyDown, Mod: tea.ModShift}} {
		h.send(key)
	}
	h.until(func() bool { return h.m.workspace.pending == 0 })
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.calls) != 0 {
		t.Fatal("terminal key changed group metadata")
	}
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	joined := strings.Join(stream.writes, ",")
	for _, input := range []string{"text:K", "text:J", "text:+", "text:-", "key:C-Up", "key:C-Down", "key:S-Up", "key:S-Down"} {
		if !strings.Contains(joined, input) {
			t.Fatalf("terminal input lost: %s, writes=%s", input, joined)
		}
	}
}

func TestGroupOrderPendingOperationDoesNotQueueStaleEditsOrStealNewSelection(t *testing.T) {
	s, first, _, _ := newGroupOrderFixture(t)
	gate := make(chan struct{})
	t.Cleanup(func() { releaseGroupOrderGate(gate) })
	s.gate = gate
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.send(tea.KeyPressMsg{Code: 'g', Text: "g"})
	h.send(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if h.m.groupDialog != nil || h.m.groupIntent != nil {
		t.Fatal("pending order opened an editor with a stale group snapshot")
	}
	h.send(tea.KeyPressMsg{Code: '2', Text: "2"})
	if h.m.selectedGroupID != first.ID {
		t.Fatal("pending order blocked ordinary stable selection")
	}
	close(gate)
	h.until(func() bool { return h.m.groupOrder == nil })
	if h.m.selectedGroupID != first.ID {
		t.Fatal("acknowledgement stole newer selection")
	}
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.calls) != 1 {
		t.Fatalf("pending key repeats queued stale writes: %d", len(s.calls))
	}
}

func TestGroupOrderAcknowledgementKeepsNewerObservedMetadataAndDropsOldOperation(t *testing.T) {
	s, first, _, second := newGroupOrderFixture(t)
	gate := make(chan struct{})
	t.Cleanup(func() { releaseGroupOrderGate(gate) })
	s.gate = gate
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	operation := h.m.groupOrder
	newer := append([]session.Group(nil), s.result...)
	for i, g := range newer {
		if g.ID == first.ID {
			newer[i].Name = "newer observed name"
			newer[i].Revision++
		}
	}
	h.send(sessionsMsg{generation: h.m.listGeneration, sessions: sampleSessions(), groups: newer})
	h.send(groupOrderResult{operation: operation, groups: s.result})
	if h.m.groupByID(first.ID).Name != "newer observed name" || h.m.groupByID(second.ID).Order != 1 {
		t.Fatal("earlier order receipt rolled back newer metadata")
	}
	h.send(groupOrderResult{operation: operation, groups: []session.Group{first}})
	if h.m.groupByID(first.ID).Name != "newer observed name" {
		t.Fatal("old operation result was applied twice")
	}
}

func TestGroupOrderAcknowledgementDoesNotRecreateAnObservedRemovedSubtree(t *testing.T) {
	s, first, child, second := newGroupOrderFixture(t)
	gate := make(chan struct{})
	t.Cleanup(func() { releaseGroupOrderGate(gate) })
	s.gate = gate
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	operation := h.m.groupOrder
	s.groupMu.Lock()
	retained := []session.Group{s.groups[0], second}
	s.groups = append([]session.Group(nil), retained...)
	s.groupMu.Unlock()
	h.send(sessionsMsg{generation: h.m.listGeneration, sessions: sampleSessions(), groups: retained})
	h.send(groupOrderResult{operation: operation, groups: s.result})
	if h.m.groupByID(first.ID).ID != "" || h.m.groupByID(child.ID).ID != "" {
		t.Fatal("old order receipt recreated a subtree whose absence was already observed")
	}
}
