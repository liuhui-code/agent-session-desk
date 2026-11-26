package ui

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type visibleCopyFixture struct {
	*workspaceFixture
	captures   []session.Session
	capture    session.Preview
	captureErr error
	gate       <-chan struct{}
}

func (s *visibleCopyFixture) CaptureVisible(ctx context.Context, target session.Session) (session.Preview, error) {
	s.mu.Lock()
	s.captures = append(s.captures, target)
	p, e, gate := s.capture, s.captureErr, s.gate
	s.mu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return session.Preview{}, ctx.Err()
		}
	}
	return p, e
}
func (s *visibleCopyFixture) captureCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.captures)
}
func newVisibleCopyFixture() *visibleCopyFixture {
	return &visibleCopyFixture{workspaceFixture: &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}, capture: session.Preview{Text: "\x1b[31mhttps://example.test\x1b[0m  \r\n  中文\ttext\t \n\n"}}
}

func TestVisibleCopyTargetsSelectedSessionAndNeverWritesTerminalInput(t *testing.T) {
	s := newVisibleCopyFixture()
	h := newWorkspaceHarness(t, s)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: 'V', Text: "V"})
	h.until(func() bool {
		return s.captureCount() == 1 && h.m.clipboard.inflight != nil && h.m.workspace.visibleCopy == nil
	})
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.captures[0].ID != "second" || s.captures[0].RunID != sampleSessions()[1].RunID || h.m.workspace.active != 0 {
		t.Fatal("visible copy targeted active Tab or changed identity")
	}
	stream := s.streams["first"]
	stream.mu.Lock()
	defer stream.mu.Unlock()
	for _, v := range stream.writes {
		if strings.HasPrefix(v, "text:") || strings.HasPrefix(v, "key:") {
			t.Fatal("V wrote terminal input")
		}
	}
}

func TestVisibleCopyUsesActualTeaOSC52OutputWithoutClipboardReads(t *testing.T) {
	s := newVisibleCopyFixture()
	m := New(context.Background(), s)
	defer m.Close()
	out := &clipboardOutput{wake: make(chan struct{}, 1)}
	program := tea.NewProgram(m, tea.WithInput(nil), tea.WithOutput(out), tea.WithWindowSize(120, 32), tea.WithEnvironment([]string{"TERM=xterm-256color"}), tea.WithoutSignalHandler())
	done := make(chan error, 1)
	go func() { _, e := program.Run(); done <- e }()
	defer func() { program.Kill(); <-done }()
	await := func(marker string) {
		t.Helper()
		timer := time.NewTimer(3 * time.Second)
		defer timer.Stop()
		for {
			out.Lock()
			ok := strings.Contains(out.buffer.String(), marker)
			out.Unlock()
			if ok {
				return
			}
			select {
			case <-out.wake:
			case <-timer.C:
				t.Fatalf("Tea output missing %q", marker)
			}
		}
	}
	await("shell marker first")
	program.Send(tea.KeyPressMsg{Code: tea.KeyF6})
	program.Send(tea.KeyPressMsg{Code: tea.KeyDown})
	program.Send(tea.KeyPressMsg{Code: 'V', Text: "V"})
	payload := "https://example.test\n  中文\ttext"
	encoded := base64.StdEncoding.EncodeToString([]byte(payload))
	await("]52;c;" + encoded)
	await("Copy request sent (")
	out.Lock()
	captured := out.buffer.String()
	out.Unlock()
	if strings.Contains(captured, "]52;c;?") {
		t.Fatal("explicit copy queried the user's clipboard")
	}
	if strings.Count(captured, "]52;c;"+encoded) != 1 {
		t.Fatal("explicit page copy emitted more than once")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.captures) != 1 || s.captures[0].ID != "second" {
		t.Fatal("actual event loop copied wrong target")
	}
}
