package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestQuickNewStartsWithoutFormAndRecordsSuccessfulConfiguration(t *testing.T) {
	exe, _ := os.Executable()
	saved := make(chan session.Launch, 1)
	s := directDragFixture()
	h := newWorkspaceHarness(t, s, Options{LoadQuickLaunch: func(context.Context) (session.Launch, error) {
		return session.Launch{Executable: exe, Dir: t.TempDir(), Args: []string{"literal value"}}, nil
	}, RememberLaunch: func(_ context.Context, l session.Launch) error { saved <- l; return nil }})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker created-1") })
	if h.m.launch != nil {
		t.Fatal("quick new opened a configuration form")
	}
	select {
	case l := <-saved:
		if len(l.Args) != 1 || l.Args[0] != "literal value" {
			t.Fatal("launch arguments changed")
		}
	default:
		t.Fatal("successful configuration not recorded")
	}
}

func TestControlSTogglesSidebarWithoutSendingSaveToTerminal(t *testing.T) {
	h := newWorkspaceHarness(t, directDragFixture())
	h.until(func() bool { return h.m.workspace.conn != nil })
	original := h.m.workspaceGeometry()
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if h.m.workspaceGeometry().left != 0 || h.m.workspaceGeometry().width != h.m.width-2*h.m.tokens.workspace.sideInset {
		t.Fatal("Ctrl+S did not collapse sidebar")
	}
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if h.m.workspaceGeometry().left != original.left {
		t.Fatal("Ctrl+S did not restore split")
	}
	if got := directDragWrites(h.m.service.(*workspaceFixture), "first"); strings.Contains(got, "key:ctrl+s") {
		t.Fatal("sidebar shortcut leaked to terminal")
	}
}

func TestNavigationFocusChangesSelectedActiveCardAppearance(t *testing.T) {
	h := newWorkspaceHarness(t, directDragFixture())
	h.until(func() bool { return h.m.workspace.conn != nil })
	terminal := h.m.workspaceSidebarCard(0, 40, 2, 0)[0]
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	navigation := h.m.workspaceSidebarCard(0, 40, 2, 0)[0]
	if testBackgroundAt(terminal, 5) == testBackgroundAt(navigation, 5) {
		t.Fatal("focus mode has no visible selection change")
	}
	if h.m.View().Cursor != nil {
		t.Fatal("navigation mode retains terminal cursor")
	}
}

func TestQuickNewSuppressesDuplicateCreation(t *testing.T) {
	exe, _ := os.Executable()
	gate := make(chan struct{})
	s := directDragFixture()
	s.startGate = gate
	h := newWorkspaceHarness(t, s, Options{LoadQuickLaunch: func(context.Context) (session.Launch, error) {
		return session.Launch{Executable: exe, Dir: t.TempDir()}, nil
	}})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { s.mu.Lock(); defer s.mu.Unlock(); return len(s.starts) == 1 })
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	s.mu.Lock()
	count := len(s.starts)
	s.mu.Unlock()
	if count != 1 {
		t.Fatal("duplicate start while pending")
	}
	close(gate)
	h.until(func() bool { return h.m.quickCreate == nil })
}

func TestQuickNewDoesNotRetryUncertainCreation(t *testing.T) {
	exe, _ := os.Executable()
	s := directDragFixture()
	s.startErr = fmt.Errorf("uncertain")
	s.startResult = session.Session{ID: "created-uncertain", RunID: "unknown"}
	saved := 0
	h := newWorkspaceHarness(t, s, Options{LoadQuickLaunch: func(context.Context) (session.Launch, error) {
		return session.Launch{Executable: exe, Dir: t.TempDir()}, nil
	}, RememberLaunch: func(context.Context, session.Launch) error { saved++; return nil }})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { return h.m.quickCreate == nil && h.m.quickUncertain })
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.starts) != 1 || saved != 0 {
		t.Fatal("uncertain start was retried or remembered")
	}
}
