package ui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestWorkspaceHelpPasteClosesWithoutInterpretingOrForwardingPayload(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	var streams []*workspaceStream
	for _, all := range []bool{false, true} {
		for _, text := range []string{"n", "q", "j", "\t", "line one\nline two"} {
			service.mu.Lock()
			streams = append(streams, service.streams["first"])
			service.mu.Unlock()
			h.send(tea.KeyPressMsg{Code: '?', Text: "?"})
			h.until(func() bool { return h.m.help && h.m.workspace.conn == nil })
			if all {
				h.send(tea.KeyPressMsg{Code: tea.KeyTab})
			}
			h.send(tea.PasteMsg{Content: text})
			if h.m.help || h.m.closed || h.m.launch != nil || h.m.workspace.active != 0 || h.m.workspace.focused {
				t.Fatalf("help did not consume paste as close only: all=%t payload=%q", all, text)
			}
			h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
		}
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.starts) != 0 {
		t.Fatal("paste triggered a launch")
	}
	for _, id := range service.opened {
		if id != "first" {
			t.Fatal("paste switched the target")
		}
	}
	streams = append(streams, service.streams["first"])
	for _, stream := range streams {
		stream.mu.Lock()
		writes := append([]string(nil), stream.writes...)
		stream.mu.Unlock()
		for _, write := range writes {
			if strings.HasPrefix(write, "text:") || strings.HasPrefix(write, "key:") || strings.HasPrefix(write, "paste:") {
				t.Fatalf("paste leaked to CLI: %q", write)
			}
		}
	}
}
