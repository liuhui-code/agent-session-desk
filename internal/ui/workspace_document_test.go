package ui

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type documentFixture struct {
	workspaceFixture
	gate    <-chan struct{}
	docMu   sync.Mutex
	targets []session.Session
	refs    []string
}

func (s *documentFixture) ResolveMarkdown(ctx context.Context, target session.Session, ref string) (string, error) {
	s.docMu.Lock()
	s.targets = append(s.targets, target)
	s.refs = append(s.refs, ref)
	s.docMu.Unlock()
	if s.gate != nil {
		select {
		case <-s.gate:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return "/fixture/plan.md", nil
}
func documentHarness(t *testing.T, gate <-chan struct{}) (*workspaceHarness, *documentFixture, *atomic.Int32) {
	s := &documentFixture{workspaceFixture: workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}, gate: gate}
	s.frame = func(_ session.Session, w, h int) terminal.Screen {
		text := "中文 docs/plan.md"
		return terminal.Screen{Width: w, Height: h, Rows: []terminal.Row{{Runs: []terminal.Run{{Text: text + strings.Repeat(" ", max(0, w-ansi.StringWidth(text))), Width: w}}}}}
	}
	opened := new(atomic.Int32)
	h := newWorkspaceHarness(t, s, Options{OpenDocument: func(context.Context, string) error { opened.Add(1); return nil }})
	h.until(func() bool { return h.m.workspace.conn != nil && len(h.m.workspace.screen.Rows) > 0 })
	h.m.workspace.focused = true
	return h, s, opened
}
func TestMarkdownCtrlClickOpensOriginalFileWithoutTerminalInput(t *testing.T) {
	h, s, opened := documentHarness(t, nil)
	g := h.m.workspaceGeometry()
	click := tea.MouseClickMsg{X: g.x + ansi.StringWidth("中文 docs/"), Y: g.y, Button: tea.MouseLeft, Mod: tea.ModCtrl}
	h.send(click)
	h.until(func() bool { return opened.Load() == 1 && h.m.workspace.document == nil })
	h.send(tea.MouseReleaseMsg{X: click.X, Y: click.Y, Button: tea.MouseLeft, Mod: tea.ModCtrl})
	s.docMu.Lock()
	if len(s.targets) != 1 || s.targets[0].ID != sampleSessions()[0].ID || s.refs[0] != "docs/plan.md" {
		t.Fatal("wrong origin")
	}
	s.docMu.Unlock()
	s.mu.Lock()
	stream := s.streams[sampleSessions()[0].ID]
	s.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if len(stream.writes) != 0 || h.m.workspace.copyPointer != nil || h.m.workspace.programPointer != nil {
		t.Fatal("click sent terminal input or started selection")
	}
}
func TestMarkdownLateResolutionDoesNotOpenAfterSessionSwitch(t *testing.T) {
	gate := make(chan struct{})
	h, _, opened := documentHarness(t, gate)
	g := h.m.workspaceGeometry()
	h.send(tea.MouseClickMsg{X: g.x + 6, Y: g.y, Button: tea.MouseLeft, Mod: tea.ModCtrl})
	original := h.m.workspace.document
	h.run(h.m.activateWorkspace(1))
	close(gate)
	h.until(func() bool { return h.m.workspace.active == 1 && h.m.workspace.conn != nil && !h.m.workspace.busy })
	h.send(documentResolved{operation: original, err: errors.New("old reply")})
	if opened.Load() != 0 {
		t.Fatal("stale resolution opened a document")
	}
}
