package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"errors"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
	"strings"
	"testing"
	"time"
)

func TestVisibleCopyNormalizesPreservesLayoutAndRejectsBadCaptures(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		truncated bool
		want      string
		err       error
	}{
		{name: "text", raw: "\x1b[31mhttps://example.test\x1b[0m  \r\n  中文\ttext\t \n\n", want: "https://example.test\n  中文\ttext"},
		{name: "internal-blank", raw: "\n  one  \n\n two\t\n\n", want: "\n  one\n\n two"},
		{name: "at-limit", raw: strings.Repeat("x", terminal.MaxClipboardBytes), want: strings.Repeat("x", terminal.MaxClipboardBytes)},
		{name: "C1-controls", raw: "\x9b31mred\x9b0m\x90synthetic\x9cvisible", want: "redvisible"},
		{name: "empty", raw: " \t\n\n"}, {name: "oversize", raw: strings.Repeat("x", terminal.MaxClipboardBytes+1)},
		{name: "truncated", raw: "partial", truncated: true}, {name: "invalid-utf8", raw: string([]byte{0xff})},
		{name: "capture-error", err: errors.New("pane gone")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newVisibleCopyFixture()
			s.capture = session.Preview{Text: tc.raw, Truncated: tc.truncated}
			s.captureErr = tc.err
			h := newWorkspaceHarness(t, s)
			h.until(func() bool { return h.m.workspace.conn != nil })
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			cmd := h.m.copyVisiblePane()
			if cmd == nil {
				t.Fatal("no capture command")
			}
			result := cmd().(visiblePaneResult)
			_, issued := h.m.Update(result)
			if tc.want == "" {
				if issued != nil || h.m.clipboard.inflight != nil || h.m.workspace.visibleCopy != nil {
					t.Fatal("invalid/empty capture issued a clipboard command")
				}
				return
			}
			if issued == nil || h.m.clipboard.inflight == nil {
				t.Fatal("valid capture not submitted")
			}
			if got := normalizeVisiblePane(tc.raw); got != tc.want {
				t.Fatalf("normalization: %q", got)
			}
		})
	}
}
func TestVisibleCopySingleFlightAndViewCancellationDiscardOldReceipt(t *testing.T) {
	s := newVisibleCopyFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	cmd := h.m.copyVisiblePane()
	op := h.m.workspace.visibleCopy
	if cmd == nil || h.m.copyVisiblePane() != nil {
		t.Fatal("duplicate V started a second capture")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.until(func() bool { return h.m.workspace.conn != nil && h.m.workspace.active == 1 })
	_, issued := h.m.Update(visiblePaneResult{operation: op, capture: session.Preview{Text: "late from old view"}})
	if issued != nil || h.m.clipboard.inflight != nil || h.m.workspace.visibleCopy != nil {
		t.Fatal("old view copied after switch")
	}
}
func TestVisibleCopyReplacementRunCannotReachClipboard(t *testing.T) {
	s := newVisibleCopyFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	cmd := h.m.copyVisiblePane()
	op := h.m.workspace.visibleCopy
	h.m.sessions[0].RunID = "replacement"
	_, issued := h.m.Update(visiblePaneResult{operation: op, capture: session.Preview{Text: "old Run"}})
	if cmd == nil || issued != nil || h.m.clipboard.inflight != nil {
		t.Fatal("replacement accepted old snapshot")
	}
}
func TestVisibleCopyTerminalVRemainsLiteralAndCaptureCanRunWithNoOpenTab(t *testing.T) {
	s := newVisibleCopyFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: 'V', Text: "V"})
	h.until(func() bool {
		s.mu.Lock()
		stream := s.streams["first"]
		s.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "text:V")
	})
	if s.captureCount() != 0 {
		t.Fatal("terminal V invoked management copy")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	h.until(func() bool { return len(h.m.workspace.tabs) == 1 && h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	h.until(func() bool { return len(h.m.workspace.tabs) == 0 })
	cmd := h.m.copyVisiblePane()
	if cmd == nil {
		t.Fatal("closed view prevented selected live pane capture")
	}
	h.send(cmd())
	if h.m.clipboard.inflight == nil {
		t.Fatal("fresh selected snapshot did not issue without an active Tab")
	}
}

func TestClipboardOldCompletionCannotReleaseANewerRequestInSameView(t *testing.T) {
	h := newWorkspaceHarness(t, newVisibleCopyFixture())
	h.until(func() bool { return h.m.workspace.conn != nil })
	generation := h.m.workspace.generation
	_, first := h.m.Update(workspaceClipboard{generation: generation, text: "first"})
	if first == nil {
		t.Fatal("first copy missing")
	}
	one := h.m.clipboard.inflight.token
	h.send(workspaceClipboard{generation: generation, text: "second"})
	h.send(workspaceClipboardSent{token: one})
	two := h.m.clipboard.inflight.token
	if one == two {
		t.Fatal("copy token reused view generation")
	}
	_, extra := h.m.Update(workspaceClipboardSent{token: one})
	if extra != nil || h.m.clipboard.inflight == nil || h.m.clipboard.inflight.token != two {
		t.Fatal("old completion released a newer copy")
	}
}
func TestVisibleCopyWaitsForNativeSubmissionAndRechecksQueuedTarget(t *testing.T) {
	h := newWorkspaceHarness(t, newVisibleCopyFixture())
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	cmd := h.m.copyVisiblePane()
	result := cmd().(visiblePaneResult)
	h.send(workspaceClipboard{generation: h.m.workspace.generation, text: "earlier native copy"})
	token := h.m.clipboard.inflight.token
	_, issued := h.m.Update(result)
	if issued != nil || h.m.workspace.visibleCopy == nil || !h.m.workspace.visibleCopy.ready {
		t.Fatal("explicit copy overlapped native submission")
	}
	// The target can change after capture, while another host submission blocks.
	h.m.sessions[0].RunID = "replacement-after-capture"
	h.send(workspaceClipboardSent{token: token})
	if h.m.clipboard.inflight != nil {
		t.Fatal("queued old Run copied after target replacement")
	}
}

func TestVisibleCopyRunningCaptureIsCancelledOnTabSwitch(t *testing.T) {
	s := newVisibleCopyFixture()
	s.gate = make(chan struct{})
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'V', Text: "V"})
	h.until(func() bool { return s.captureCount() == 1 })
	op := h.m.workspace.visibleCopy
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	received := false
	for !received {
		select {
		case msg := <-h.messages:
			if result, ok := msg.(visiblePaneResult); ok && result.operation == op {
				if !errors.Is(result.err, context.Canceled) {
					t.Fatalf("running capture not cancelled: %v", result.err)
				}
				received = true
			}
			h.send(msg)
		case <-timer.C:
			t.Fatal("running capture did not finish after view cancellation")
		}
	}
	h.until(func() bool { return h.m.workspace.conn != nil && h.m.workspace.active == 1 })
	if h.m.workspace.visibleCopy != nil || h.m.clipboard.inflight != nil || s.captureCount() != 1 {
		t.Fatal("cancelled request copied or restarted")
	}
}
