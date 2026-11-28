package ui

import (
	"context"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/snippets"
)

func snippetHarness(t *testing.T) (*workspaceHarness, *workspaceFixture) {
	s := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, s, Options{LoadSnippets: func(context.Context, session.Session) (snippets.Library, error) {
		return snippets.Library{ProjectDir: "/synthetic", Items: []snippets.Item{{ID: "review", Title: "Review changes", Body: "审查 ${HOME}\nleave draft", Source: "Project"}}}, nil
	}})
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	return h, s
}
func TestSnippetsPreviewThenPasteOnlyToOriginalTerminalWithoutEnter(t *testing.T) {
	h, s := snippetHarness(t)
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Review changes") })
	h.send(tea.PasteMsg{Content: "review"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !strings.Contains(h.m.View().Content, "Multiline paste") {
		t.Fatal("missing explicit multiline preview")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		if h.m.snippets.dialog != nil || h.m.workspace.pending != 0 {
			return false
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		p := s.streams["first"]
		p.mu.Lock()
		defer p.mu.Unlock()
		return strings.Contains(strings.Join(p.writes, ","), "paste:审查 ${HOME}\nleave draft")
	})
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.streams["first"]
	p.mu.Lock()
	defer p.mu.Unlock()
	all := strings.Join(p.writes, ",")
	if strings.Contains(all, "key:Enter") || strings.Contains(all, "text:review") || strings.Contains(all, "key:M-s") || len(s.opened) != 1 {
		t.Fatalf("picker leaked input or reconnected: %q %v", all, s.opened)
	}
}

func TestSnippetsKeepGlobalActivationInsideExistingModal(t *testing.T) {
	h, s := attentionUIFixture(t)
	h.m.snippets.load = func(context.Context, session.Session) (snippets.Library, error) { return snippets.Library{}, nil }
	h.m.attentionFocus = func() bool { return true }
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	h.until(func() bool { return h.m.snippets.dialog != nil && !h.m.snippets.dialog.loading })
	h.send(GlobalActivate{})
	if h.m.attention.busy || h.m.workspace.transition != workspaceIdle || h.m.workspace.active != 0 {
		t.Fatal("global activation escaped snippets and navigated to another Run")
	}
}

func TestSnippetsTinyWindowCannotPasteUntilPreviewVisible(t *testing.T) {
	h, _ := snippetHarness(t)
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	h.until(func() bool { return h.m.snippets.dialog != nil && !h.m.snippets.dialog.loading })
	h.send(tea.WindowSizeMsg{Width: 12, Height: 3})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.snippets.dialog == nil {
		t.Fatal("tiny invisible preview allowed a paste")
	}
}

func TestSnippetsCancelRestoresFocusAndRejectsReplacementRun(t *testing.T) {
	h, s := snippetHarness(t)
	h.m.workspace.sidebarHidden = true
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	h.until(func() bool { return h.m.snippets.dialog != nil && !h.m.snippets.dialog.loading })
	h.send(tea.PasteMsg{Content: "review"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if h.m.snippets.dialog != nil || !h.m.workspace.focused || !h.m.workspace.sidebarHidden {
		t.Fatal("cancel did not restore focus/sidebar")
	}
	h.until(func() bool { return h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	h.until(func() bool { return h.m.snippets.dialog != nil && !h.m.snippets.dialog.loading })
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.m.workspace.tabs[0].session.RunID = "replacement"
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.snippets.dialog == nil || !strings.Contains(h.m.snippets.dialog.notice, "Original terminal changed") {
		t.Fatal("replacement Run was permitted")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.streams {
		p.mu.Lock()
		for _, v := range p.writes {
			if strings.HasPrefix(v, "paste:") {
				t.Errorf("unexpected terminal input %q", v)
			}
		}
		p.mu.Unlock()
	}
}

func TestSnippetsDiscardLateLibraryAndRespectInputBarrier(t *testing.T) {
	h, _ := snippetHarness(t)
	h.m.workspace.pending = 1
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	if h.m.snippets.dialog != nil {
		t.Fatal("accepted input interrupted")
	}
	h.m.workspace.pending = 0
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	old := h.m.snippets.dialog
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.send(snippetsLoaded{dialog: old, library: snippets.Library{Items: []snippets.Item{{ID: "late", Body: "no"}}}})
	if h.m.snippets.dialog != nil {
		t.Fatal("late loading result reopened cancelled modal")
	}
}

func TestSnippetsFramesAndPreviewStayWithinViewport(t *testing.T) {
	h, _ := snippetHarness(t)
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	h.until(func() bool { return h.m.snippets.dialog != nil && !h.m.snippets.dialog.loading })
	for _, size := range [][2]int{{120, 32}, {63, 20}, {49, 12}, {32, 10}, {12, 3}, {0, 0}} {
		h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		for _, preview := range []bool{false, true} {
			h.m.snippets.dialog.preview = preview
			v := h.m.View()
			if size[0] == 0 || size[1] == 0 {
				if v.Content != "" {
					t.Fatal("empty viewport rendered")
				}
				continue
			}
			lines := strings.Split(v.Content, "\n")
			if len(lines) != size[1] {
				t.Fatalf("rows %v: %d", size, len(lines))
			}
			for _, line := range lines {
				if ansi.StringWidth(line) != size[0] {
					t.Fatalf("width %v: %q", size, line)
				}
			}
			if v.Cursor != nil && (v.Cursor.X < 0 || v.Cursor.X >= size[0] || v.Cursor.Y < 0 || v.Cursor.Y >= size[1]) {
				t.Fatal("cursor outside frame")
			}
		}
	}
}

func TestSnippetsOnlyOneLoadCanRemainOutstandingAfterCancel(t *testing.T) {
	h, _ := snippetHarness(t)
	gate := make(chan struct{})
	started := make(chan struct{}, 1)
	h.m.snippets.load = func(context.Context, session.Session) (snippets.Library, error) {
		started <- struct{}{}
		<-gate
		return snippets.Library{}, nil
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	<-started
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	close(gate)
	if h.m.snippets.dialog != nil {
		t.Fatal("cancelled slow file read allowed a second outstanding load")
	}
}

func TestSnippetsMinimumFrameKeepsSelectionBodyAndActionsVisible(t *testing.T) {
	h, _ := snippetHarness(t)
	h.send(tea.KeyPressMsg{Code: tea.KeyF7})
	h.until(func() bool { return h.m.snippets.dialog != nil && !h.m.snippets.dialog.loading })
	h.send(tea.WindowSizeMsg{Width: 32, Height: 10})
	view := h.m.View().Content
	if !strings.Contains(view, "Project") || !strings.Contains(view, "Preview") {
		t.Fatal("minimum list hides selection or action")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	view = h.m.View().Content
	if !strings.Contains(view, "审查 ${HOME}") || !strings.Contains(view, "Enter Paste") || !strings.Contains(view, "Esc Back") {
		t.Fatal("minimum preview hides body or confirmation/back actions")
	}
}

func TestSnippetsDefaultFunctionKeyAcceptsLockFlags(t *testing.T) {
	h, _ := snippetHarness(t)
	h.send(tea.KeyPressMsg{Code: tea.KeyF7, Mod: tea.ModCapsLock})
	if h.m.snippets.dialog == nil {
		t.Fatal("F7 did not open snippets before terminal forwarding")
	}
}
