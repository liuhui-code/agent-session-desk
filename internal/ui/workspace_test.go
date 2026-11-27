package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type workspaceFixture struct {
	openGate    <-chan struct{}
	startGate   <-chan struct{}
	startErr    error
	startResult session.Session
	starts      []session.Launch
	created     []session.Session
	frame       func(session.Session, int, int) terminal.Screen
	fakeService
	mu      sync.Mutex
	opened  []string
	streams map[string]*workspaceStream
}

func (s *workspaceFixture) List(ctx context.Context) ([]session.Session, error) {
	rows, err := s.fakeService.List(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	return append(append([]session.Session(nil), rows...), s.created...), err
}
func (s *workspaceFixture) Start(ctx context.Context, l session.Launch) (session.Session, error) {
	s.mu.Lock()
	s.starts = append(s.starts, l)
	if s.startErr != nil {
		v, err := s.startResult, s.startErr
		s.mu.Unlock()
		return v, err
	}
	v := session.Session{ID: fmt.Sprintf("created-%d", len(s.starts)), RunID: fmt.Sprintf("created-run-%d", len(s.starts)), Title: l.Name, Command: l.Executable, TerminalState: session.TerminalAvailable}
	s.created = append(s.created, v)
	gate := s.startGate
	s.mu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return v, ctx.Err()
		}
	}
	return v, nil
}

func (s *workspaceFixture) OpenTerminal(ctx context.Context, v session.Session, w, h int) (terminal.Connection, error) {
	if s.openGate != nil {
		select {
		case <-s.openGate:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	stream := &workspaceStream{closed: make(chan struct{}), frames: make(chan terminal.Screen, 1)}
	if s.frame != nil {
		stream.frames <- s.frame(v, w, h)
	} else {
		stream.frames <- terminal.Screen{Width: w, Height: h, Cursor: terminal.Cursor{X: 2, Y: 0, Visible: true}, Rows: []terminal.Row{{Runs: []terminal.Run{{Text: "shell marker " + v.ID, Width: len("shell marker " + v.ID)}}}}}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.streams == nil {
		s.streams = map[string]*workspaceStream{}
	}
	s.opened = append(s.opened, v.ID)
	s.streams[v.ID] = stream
	return stream, nil
}

type workspaceStream struct {
	gate   <-chan struct{}
	mu     sync.Mutex
	closed chan struct{}
	once   sync.Once
	frames chan terminal.Screen
	writes []string
}

func (s *workspaceStream) Next(ctx context.Context) (terminal.Screen, error) {
	select {
	case f := <-s.frames:
		return f, nil
	case <-ctx.Done():
		return terminal.Screen{}, ctx.Err()
	case <-s.closed:
		return terminal.Screen{}, fmt.Errorf("closed")
	}
}
func (s *workspaceStream) Close() error { s.once.Do(func() { close(s.closed) }); return nil }
func (s *workspaceStream) record(ctx context.Context, text string) error {
	s.mu.Lock()
	gate := s.gate
	s.mu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writes = append(s.writes, text)
	return nil
}
func (s *workspaceStream) SendText(ctx context.Context, text string) error {
	return s.record(ctx, "text:"+text)
}
func (s *workspaceStream) SendKey(ctx context.Context, key string) error {
	return s.record(ctx, "key:"+key)
}
func (s *workspaceStream) Paste(ctx context.Context, text string) error {
	return s.record(ctx, "paste:"+text)
}

func (s *workspaceStream) Resize(ctx context.Context, w, h int) error {
	return s.record(ctx, fmt.Sprintf("size:%d,%d", w, h))
}
func (s *workspaceStream) Refresh(ctx context.Context) error { return s.record(ctx, "refresh") }
func (s *workspaceStream) Scroll(ctx context.Context, x, y int, up bool) error {
	return s.record(ctx, fmt.Sprintf("scroll:%t,%d,%d", up, x, y))
}

type workspaceHarness struct {
	t        *testing.T
	m        *Model
	messages chan tea.Msg
	ctx      context.Context
	cancel   context.CancelFunc
}

func newWorkspaceHarness(t *testing.T, service Service, options ...Options) *workspaceHarness {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	settings := Options{SidebarMode: "flat"} // Explicit flat variant; grouped tests use defaults.
	if len(options) > 0 {
		settings = options[0]
	}
	h := &workspaceHarness{t: t, m: NewWithOptions(ctx, service, settings), messages: make(chan tea.Msg, 32), ctx: ctx, cancel: cancel}
	t.Cleanup(func() { h.m.Close(); cancel() })
	h.send(tea.WindowSizeMsg{Width: 120, Height: 32})
	h.send(tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	h.run(h.m.Init())
	return h
}
func (h *workspaceHarness) run(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	go func() {
		msg := cmd()
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, child := range batch {
				h.run(child)
			}
			return
		}
		select {
		case h.messages <- msg:
		case <-h.ctx.Done():
		}
	}()
}
func (h *workspaceHarness) send(msg tea.Msg) {
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, cmd := range batch {
			h.run(cmd)
		}
		return
	}
	_, cmd := h.m.Update(msg)
	h.run(cmd)
}
func (h *workspaceHarness) until(check func() bool) {
	h.t.Helper()
	timeout := time.NewTimer(3 * time.Second)
	defer timeout.Stop()
	for !check() {
		select {
		case msg := <-h.messages:
			h.send(msg)
		case <-timeout.C:
			h.t.Fatalf("workspace did not reach expected state:\n%s", h.m.View().Content)
		}
	}
}

func TestWorkspaceShowsNativeTerminalWithTabNavigationAndRealCursor(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	view := h.m.View()
	if !strings.Contains(view.Content, "开发") || !strings.Contains(view.Content, "评审") {
		t.Fatal("session Tabs are not available alongside terminal")
	}
	if view.Cursor == nil {
		t.Fatal("native terminal cursor is hidden")
	}
}

func TestWorkspaceTerminalKeysDoNotTriggerManagerQuit(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	h.send(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Join(stream.writes, ",") == "text:q,key:C-c"
	})
	if h.m.closed {
		t.Fatal("terminal keys quit the manager")
	}
}

func TestWorkspaceNativeViewportAndCursorReachLastCell(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}, frame: func(_ session.Session, w, h int) terminal.Screen {
		rows := make([]terminal.Row, h)
		rows[h-1] = terminal.Row{Runs: []terminal.Run{{Text: strings.Repeat(" ", w-1) + "Z", Width: w}}}
		return terminal.Screen{Width: w, Height: h, Cursor: terminal.Cursor{X: w - 1, Y: h - 1, Visible: true}, Rows: rows}
	}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.View().Cursor != nil })
	view := h.m.View()
	lines := strings.Split(ansi.Strip(view.Content), "\n")
	if view.Cursor.Y >= len(lines) {
		t.Fatal("cursor outside visible viewport")
	}
	line := lines[view.Cursor.Y]
	idx := strings.Index(line, "Z")
	if idx < 0 || ansi.StringWidth(line[:idx]) != view.Cursor.X {
		t.Fatalf("native last cell clipped or cursor offset wrong: cursor=%+v line=%q", view.Cursor, line)
	}
}

func TestWorkspaceSwitchesTabsByStableRunWithoutExtraInput(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	h.send(tea.KeyPressMsg{Code: tea.KeyPgUp, Mod: tea.ModCtrl})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	defer service.mu.Unlock()
	if strings.Join(service.opened, ",") != "first,second,first" {
		t.Fatalf("wrong Tab target identity: %v", service.opened)
	}
	for _, stream := range service.streams {
		stream.mu.Lock()
		writes := append([]string(nil), stream.writes...)
		stream.mu.Unlock()
		if len(writes) != 0 {
			t.Fatalf("Tab switch injected input: %v", writes)
		}
	}
}

func TestWorkspaceSwitchDrainsOldInputAndKeepsNewInputOnNewTab(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	first := service.streams["first"]
	service.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'a', Text: "a"})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: 'b', Text: "b"})
	close(gate)
	h.until(func() bool {
		service.mu.Lock()
		second := service.streams["second"]
		service.mu.Unlock()
		if second == nil {
			return false
		}
		first.mu.Lock()
		a := strings.Join(first.writes, ",")
		first.mu.Unlock()
		second.mu.Lock()
		b := strings.Join(second.writes, ",")
		second.mu.Unlock()
		return a == "text:a" && b == "text:b"
	})
}

func TestWorkspaceOversizedInputIsRefusedBeforeQueueing(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: 'a', Text: strings.Repeat("a", terminal.MaxInputBytes+1)})
	if !strings.Contains(h.m.View().Content, "limit") {
		t.Fatal("oversized input was not refused at the UI boundary")
	}
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if len(stream.writes) != 0 {
		t.Fatal("oversized input reached terminal writer")
	}
}

func TestWorkspaceManagementFocusDoesNotSendKeysToTerminal(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	if h.m.View().Cursor != nil || h.m.workspace.focused {
		t.Fatal("management focus is not visible")
	}
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	h.until(func() bool { return h.m.closed })
	service.mu.Lock()
	stream := service.streams["first"]
	service.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	for _, write := range stream.writes {
		if strings.HasPrefix(write, "key:") || strings.HasPrefix(write, "text:") {
			t.Fatalf("manager key reached terminal: %q", write)
		}
	}
}

func TestWorkspaceQuitDrainsPreviouslyTypedInput(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	first := service.streams["first"]
	service.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'a', Text: "a"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if h.m.closed {
		close(gate)
		t.Fatal("quit cancelled previously typed input before it was written")
	}
	close(gate)
	h.until(func() bool { return h.m.closed })
	first.mu.Lock()
	defer first.mu.Unlock()
	if len(first.writes) == 0 || first.writes[0] != "text:a" {
		t.Fatalf("typed input lost during graceful quit: %v", first.writes)
	}
}

func TestWorkspaceClosingAndReopeningTabKeepsOriginalRun(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	h.until(func() bool {
		lines := strings.Split(ansi.Strip(h.m.View().Content), "\n")
		return len(lines) > 1 && len(h.m.workspace.tabs) == 1 && h.m.workspace.tabs[0].session.ID == "second" && strings.Contains(h.m.View().Content, "shell marker second")
	})
	h.send(tea.KeyPressMsg{Code: tea.KeyUp})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	defer service.mu.Unlock()
	if strings.Join(service.opened, ",") != "first,second,first" {
		t.Fatalf("closed Tab was not reopened by original identity: %v", service.opened)
	}
}

func TestWorkspaceCloseTabDrainsPreviouslyTypedInput(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	service.mu.Lock()
	first := service.streams["first"]
	service.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'a', Text: "a"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	close(gate)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	first.mu.Lock()
	defer first.mu.Unlock()
	if len(first.writes) == 0 || first.writes[0] != "text:a" {
		t.Fatalf("closing Tab lost previously typed input: %v", first.writes)
	}
}
