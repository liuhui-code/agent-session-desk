package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"errors"
	"github.com/charmbracelet/colorprofile"
	"github.com/liuhui-code/agent-session-desk/internal/pathsuggest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func openPickerLaunch(h *workspaceHarness) {
	h.run(h.m.openLaunchDialog())
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
}
func TestLaunchPathPickerSeparatesCustomBrowseAndFieldAdvance(t *testing.T) {
	s := newGroupedFixture()
	h := newWorkspaceHarness(t, s, Options{LoadRecentDirs: func(context.Context) ([]string, error) { return []string{"/known-project"}, nil }})
	openPickerLaunch(h)
	h.until(func() bool { return len(h.m.launch.paths.matches) == 1 })
	p := h.m.launch.paths
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	if !p.active || p.selected != 1 {
		t.Fatal("down did not enter real candidate browse")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyUp})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.launch.focus != 2 || p.active || !p.hidden {
		t.Fatal("custom Enter must edit same path field")
	}
	h.send(tea.PasteMsg{Content: "known"})
	h.until(func() bool { return p.ready })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.ColorProfileMsg{Profile: colorprofile.ANSI256})
	h.m.applyUITheme("light")
	if h.m.launch.paths != p || !p.active || p.selected != 1 {
		t.Fatal("restyling reset browsing")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if h.m.launch == nil || p.active || p.hidden {
		t.Fatal("browse Esc must return to edit, keeping menu")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.launch.focus != 3 || h.m.launch.fields[2].Value() != "/known-project" || !p.hidden {
		t.Fatal("real Enter must accept and advance")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.starts) != 0 {
		t.Fatal("directory browse started a process")
	}
}
func TestPathTabCyclesFilesystemAndPreservesFollowingPasteOrder(t *testing.T) {
	root := t.TempDir()
	for _, n := range []string{"alpha", "alpine"} {
		if err := os.Mkdir(filepath.Join(root, n), 0700); err != nil {
			t.Fatal(err)
		}
	}
	h := newWorkspaceHarness(t, newGroupedFixture())
	openPickerLaunch(h)
	h.until(func() bool { return h.m.launch.paths.ready })
	gate := make(chan struct{})
	started := make(chan struct{}, 1)
	h.m.workspace.directories.scan = func(ctx context.Context, q string) pathsuggest.Completion {
		started <- struct{}{}
		<-gate
		return pathsuggest.Complete(ctx, q)
	}
	h.send(tea.PasteMsg{Content: root + string(os.PathSeparator) + "al"})
	<-started
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	close(gate)
	h.until(func() bool { return h.m.launch.paths.cycleIndex == 1 })
	if got := h.m.launch.fields[2].Value(); got != filepath.Join(root, "alpine") {
		t.Fatalf("cycle order: %q", got)
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.send(tea.PasteMsg{Content: `["literal argument"]`})
	if h.m.launch.focus != 3 || h.m.launch.fields[3].Value() != `["literal argument"]` {
		t.Fatal("completion changed following field input")
	}
}
func TestPathScannerSerializesUninterruptibleCallsAndDropsClosedModal(t *testing.T) {
	h := newWorkspaceHarness(t, newGroupedFixture())
	var calls atomic.Int32
	gate := make(chan struct{})
	started := make(chan struct{}, 1)
	h.m.workspace.directories.scan = func(ctx context.Context, q string) pathsuggest.Completion {
		if calls.Add(1) == 1 {
			started <- struct{}{}
			<-gate
			return pathsuggest.Completion{Paths: []string{"/stale"}}
		}
		return pathsuggest.Completion{Paths: []string{"/fresh"}}
	}
	openPickerLaunch(h)
	h.until(func() bool { return h.m.launch.paths.ready })
	h.send(tea.PasteMsg{Content: "./initial"})
	<-started
	old := h.m.launch.paths
	h.send(tea.PasteMsg{Content: "./a"})
	h.send(tea.PasteMsg{Content: "b"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	openPickerLaunch(h)
	h.send(tea.PasteMsg{Content: "./new"})
	p := h.m.launch.paths
	if old == p || calls.Load() != 1 {
		t.Fatal("parallel scan or reused modal")
	}
	close(gate)
	h.until(func() bool { return p.ready })
	if calls.Load() != 2 || strings.Contains(h.m.View().Content, "stale") || h.m.launch.fields[2].Value() != "./new" {
		t.Fatal("stale scan affected fresh modal")
	}
}
func TestSuccessfulLaunchSavesHintFailureWithoutRetryingOrBlockingTab(t *testing.T) {
	gate := make(chan struct{})
	var saved atomic.Int32
	s := newGroupedFixture()
	h := newWorkspaceHarness(t, s, Options{RememberDirs: func(ctx context.Context, dirs []string) error {
		saved.Add(1)
		select {
		case <-gate:
		case <-ctx.Done():
		}
		return errors.New("synthetic hint failure")
	}})
	h.run(h.m.openLaunchDialog())
	h.m.launch.fields[1].SetValue("synthetic-shell")
	h.m.launch.fields[2].SetValue(t.TempDir())
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.until(func() bool { return h.m.launch == nil && h.m.workspace.conn != nil })
	close(gate)
	h.until(func() bool { return strings.Contains(h.m.status, "最近目录未保存") })
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.starts) != 1 || saved.Load() != 1 {
		t.Fatal("hint failure retried launch")
	}
}

func TestTinyResizeReleasesBrowseFocusAndPreservesDirectoryDraft(t *testing.T) {
	h := newWorkspaceHarness(t, newGroupedFixture(), Options{LoadRecentDirs: func(context.Context) ([]string, error) { return []string{"/known"}, nil }})
	openPickerLaunch(h)
	h.until(func() bool { return h.m.launch.paths.ready })
	h.send(tea.PasteMsg{Content: "known"})
	h.until(func() bool { return h.m.launch.paths.ready })
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.WindowSizeMsg{Width: 20, Height: 8})
	h.send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if h.m.launch.fields[2].Value() != "knownx" || !h.m.launch.fields[2].Focused() {
		t.Fatal("hidden browse retained focus and swallowed editing")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if h.m.launch != nil {
		t.Fatal("tiny dialog trapped Escape")
	}
}

func TestExplicitTabCompletesBareRelativeDirectoryAndKeepsInvalidPathInField(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.Mkdir("project-one", 0700); err != nil {
		t.Fatal(err)
	}
	h := newWorkspaceHarness(t, newGroupedFixture())
	openPickerLaunch(h)
	h.until(func() bool { return h.m.launch.paths.ready })
	h.send(tea.PasteMsg{Content: "project-"})
	h.until(func() bool { return h.m.launch.paths.ready })
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.until(func() bool { return h.m.launch.fields[2].Value() == "project-one" })
	if h.m.launch.focus != 2 {
		t.Fatal("completion advanced field")
	}
	h.m.launch.fields[2].SetValue("./missing-directory")
	h.run(h.m.pathEdited(h.m.launch.paths, "./missing-directory"))
	h.until(func() bool { return h.m.launch.paths.ready })
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.until(func() bool { return !h.m.launch.paths.tabPending })
	if h.m.launch.focus != 2 {
		t.Fatal("invalid directory silently advanced")
	}
}

func TestLaunchControlNavigationKeepsEditingUntilExplicitAccept(t *testing.T) {
	h := newWorkspaceHarness(t, newGroupedFixture(), Options{LoadRecentDirs: func(context.Context) ([]string, error) { return []string{"/one", "/two"}, nil }})
	openPickerLaunch(h)
	h.until(func() bool { return h.m.launch.paths.ready })
	h.send(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	p := h.m.launch.paths
	if p.active || p.selected != 1 || !h.m.launch.fields[2].Focused() {
		t.Fatal("Ctrl+N should keep editing focus")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	if h.m.launch.focus != 3 || h.m.launch.fields[2].Value() != "/one" {
		t.Fatal("Tab did not accept explicitly navigated candidate")
	}
}
