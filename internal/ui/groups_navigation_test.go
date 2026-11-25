package ui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestTabRevealOfNestedGroupDoesNotQueueAPreferenceWriteBehindAnAcceptedFold(t *testing.T) {
	s := newGroupedFixture()
	root, _ := s.CreateGroup(context.Background(), "", "project", "")
	child, _ := s.CreateGroup(context.Background(), root.ID, "review", "")
	_, _ = s.MoveToGroup(context.Background(), sampleSessions()[1], child.ID)
	started := make(chan config.UISettings, 4)
	release := make(chan struct{}, 4)
	h := newWorkspaceHarness(t, s, Options{SaveUISettings: func(ctx context.Context, v config.UISettings) error {
		started <- v
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '2', Text: "2"})
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("fold save not started")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	release <- struct{}{}
	h.until(func() bool { return !h.m.preferences.busy })
	select {
	case v := <-started:
		t.Fatalf("Tab reveal queued another write: %+v", v)
	default:
	}
	if h.m.preferences.saved.CollapsedGroups != root.ID || strings.Contains(h.m.preferences.wanted.CollapsedGroups, root.ID) {
		t.Fatal("temporary reveal changed accepted durable fold")
	}
	if !strings.Contains(ansi.Strip(h.m.View().Content), "▾ review (1)") {
		t.Fatal("Tab did not reveal nested ancestors")
	}
}

func TestEndSelectsLastEmptyGroupAndNestedMetadataClickKeepsExactRun(t *testing.T) {
	s := newGroupedFixture()
	root, _ := s.CreateGroup(context.Background(), "", "project", "")
	child, _ := s.CreateGroup(context.Background(), root.ID, "review", "")
	_, _ = s.MoveToGroup(context.Background(), sampleSessions()[1], child.ID)
	for i := 0; i < 8; i++ {
		_, _ = s.CreateGroup(context.Background(), "", strings.Repeat("x", i+1), "")
	}
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	if !strings.Contains(ansi.Strip(h.m.View().Content), "  ▾ xxxxxxxx (0)") {
		t.Fatal("End skipped empty groups")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyHome})
	clickWorkspaceListRow(t, h, "评审")
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
}

func TestTerminalGroupLettersRemainLiteralInput(t *testing.T) {
	s := newGroupedFixture()
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: 'g', Text: "g"})
	h.send(tea.KeyPressMsg{Code: 'M', Text: "M"})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	writes := strings.Join(stream.writes, ",")
	if !strings.Contains(writes, "text:g") || !strings.Contains(writes, "text:M") || h.m.groupDialog != nil {
		t.Fatalf("terminal keys intercepted: %s", writes)
	}
}

func TestDiscoveryOfAnExplicitMoveRevealsTheSelectedRunWithoutSavingTheFold(t *testing.T) {
	s := newGroupedFixture()
	root, _ := s.CreateGroup(context.Background(), "", "project", "")
	h := newWorkspaceHarness(t, s, Options{CollapsedGroups: root.ID})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	_, _ = s.MoveToGroup(context.Background(), sampleSessions()[0], root.ID)
	view, err := s.Workspace(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	h.send(sessionsMsg{generation: h.m.listGeneration, sessions: view.Sessions, groups: view.Groups})
	selected, ok := h.m.selected()
	if !ok || !sameRun(selected, sampleSessions()[0]) || !strings.Contains(ansi.Strip(h.m.View().Content), "▾ project (1)") {
		t.Fatal("selected Run became invisible after observed move")
	}
	if h.m.preferences.saved.CollapsedGroups != root.ID || h.m.preferences.busy {
		t.Fatal("observed move saved presentation change")
	}
}

var _ groupService = (*groupedFixture)(nil)
