package ui

import (
	"context"
	"reflect"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/pathsuggest"
)

func TestExactDirectoryAsyncTabKeepsFollowingArgumentsEnvironmentAndCreate(t *testing.T) {
	s := newGroupedFixture()
	h := newWorkspaceHarness(t, s)
	h.run(h.m.openLaunchDialog())
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: "synthetic-cli"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.until(func() bool { return h.m.launch.paths.ready })
	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	t.Cleanup(release)
	started := make(chan struct{}, 1)
	h.m.workspace.directories.scan = func(ctx context.Context, query string) pathsuggest.Completion {
		started <- struct{}{}
		<-gate
		return pathsuggest.Complete(ctx, query)
	}
	dir := t.TempDir()
	h.send(tea.PasteMsg{Content: dir})
	<-started
	p := h.m.launch.paths
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: `["literal argument"]`})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: `["ASD_EXACT_TAB=retained"]`})
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	s.mu.Lock()
	before := len(s.starts)
	s.mu.Unlock()
	if before != 0 {
		t.Fatal("created before explicit Tab completed")
	}
	release()
	h.until(func() bool { return p.ready && !p.tabPending })
	if h.m.launch != nil && !h.m.launch.busy {
		t.Fatal("exact-directory focus change discarded queued args/env/create events")
	}
	h.until(func() bool { return h.m.launch == nil })
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.starts) != 1 {
		t.Fatal("queued create repeated")
	}
	got := s.starts[0]
	if got.Dir != dir || !reflect.DeepEqual(got.Args, []string{"literal argument"}) || !reflect.DeepEqual(got.Env, []string{"ASD_EXACT_TAB=retained"}) {
		t.Fatalf("queued field boundaries changed: %+v", got)
	}
}
