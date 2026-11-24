package ui

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type workspaceCopyFixture struct {
	workspaceFixture
	clipboard chan string
	readCalls chan string
}
type workspaceCopyStream struct {
	*workspaceStream
	clipboard <-chan string
	readCalls chan<- string
	target    string
}

func (s *workspaceCopyFixture) OpenTerminal(ctx context.Context, target session.Session, width, height int) (terminal.Connection, error) {
	conn, err := s.workspaceFixture.OpenTerminal(ctx, target, width, height)
	if err != nil {
		return nil, err
	}
	return &workspaceCopyStream{conn.(*workspaceStream), s.clipboard, s.readCalls, target.ID}, nil
}

func (s *workspaceCopyStream) EnterCopyMode(ctx context.Context) error {
	return s.record(ctx, "copy-mode")
}

func (s *workspaceCopyStream) NextClipboard(ctx context.Context) (string, error) {
	if s.readCalls != nil {
		s.readCalls <- s.target
	}
	select {
	case text := <-s.clipboard:
		return text, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func TestWorkspaceFinishesIssuedClipboardBeforeReadingNewTabCopies(t *testing.T) {
	service := &workspaceCopyFixture{workspaceFixture: workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}, readCalls: make(chan string, 4)}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	select {
	case <-service.readCalls:
	case <-time.After(time.Second):
		t.Fatal("initial clipboard reader did not start")
	}
	old := h.m.workspace.generation
	// Hold the returned renderer command to represent an issued copy whose
	// output has not completed. Navigation can proceed while it is pending.
	_, issued := h.m.Update(workspaceClipboard{generation: old, text: "copy A"})
	if issued == nil {
		t.Fatal("copy A was not submitted")
	}
	issuedToken := h.m.clipboard.inflight.token
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	select {
	case target := <-service.readCalls:
		t.Fatalf("new Tab clipboard started before A submission completed: %s", target)
	case <-time.After(30 * time.Millisecond):
	}
	_, release := h.m.Update(workspaceClipboardSent{token: issuedToken})
	if release == nil {
		t.Fatal("A completion did not release the current clipboard reader")
	}
	h.run(release)
	select {
	case target := <-service.readCalls:
		if target != "second" {
			t.Fatal("completed copy resumed the old Run reader")
		}
	case <-time.After(time.Second):
		t.Fatal("current Tab clipboard reader did not resume")
	}
	_, duplicate := h.m.Update(workspaceClipboardSent{token: issuedToken})
	if duplicate != nil {
		t.Fatal("duplicate completion started a second reader")
	}
}

type clipboardOutput struct {
	sync.Mutex
	buffer bytes.Buffer
	wake   chan struct{}
}

func (out *clipboardOutput) Write(data []byte) (int, error) {
	out.Lock()
	n, err := out.buffer.Write(data)
	out.Unlock()
	select {
	case out.wake <- struct{}{}:
	default:
	}
	return n, err
}

func TestWorkspaceForwardsClipboardThroughActualTeaProgramInOrder(t *testing.T) {
	service := &workspaceCopyFixture{workspaceFixture: workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}, clipboard: make(chan string, 1)}
	m := New(context.Background(), service)
	defer m.Close()
	out := &clipboardOutput{wake: make(chan struct{}, 1)}
	program := tea.NewProgram(m, tea.WithInput(nil), tea.WithOutput(out), tea.WithWindowSize(120, 32), tea.WithEnvironment([]string{"TERM=xterm-256color"}), tea.WithoutSignalHandler())
	done := make(chan error, 1)
	go func() { _, err := program.Run(); done <- err }()
	defer func() { program.Kill(); <-done }()
	for _, text := range []string{"copy one", "中文 copy two"} {
		service.clipboard <- text
		encoded := base64.StdEncoding.EncodeToString([]byte(text))
		deadline := time.NewTimer(3 * time.Second)
		for {
			out.Lock()
			found := strings.Contains(out.buffer.String(), "]52;c;"+encoded)
			out.Unlock()
			if found {
				break
			}
			select {
			case <-out.wake:
			case <-deadline.C:
				select {
				case err := <-done:
					done <- err
					t.Fatalf("Tea program exited before clipboard forwarding: %v", err)
				default:
				}
				out.Lock()
				captured := out.buffer.String()
				out.Unlock()
				service.mu.Lock()
				opened := append([]string(nil), service.opened...)
				service.mu.Unlock()
				t.Fatalf("clipboard not observed; opened=%v output=%q", opened, captured)
			}
		}
		deadline.Stop()
	}
}

func TestWorkspaceCopyEntryIsSemanticAndTerminalVStaysLiteral(t *testing.T) {
	service := &workspaceCopyFixture{workspaceFixture: workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: 'v', Text: "v"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'v', Text: "v"})
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return strings.Contains(strings.Join(stream.writes, ","), "copy-mode")
	})
	h.send(tea.KeyPressMsg{Code: '/', Text: "/needle"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	h.until(func() bool {
		service.mu.Lock()
		stream := service.streams["first"]
		service.mu.Unlock()
		stream.mu.Lock()
		defer stream.mu.Unlock()
		writes := strings.Join(stream.writes, ",")
		return strings.Contains(writes, "text:/needle,key:Enter")
	})
	service.mu.Lock()
	defer service.mu.Unlock()
	stream := service.streams["first"]
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if strings.Count(strings.Join(stream.writes, ","), "text:v") != 1 || strings.Join(service.opened, ",") != "first" {
		t.Fatal("copy entry typed a command or replaced the original terminal")
	}
}

func TestWorkspaceNativeSearchPromptKeepsCursorAndUnicodeInputVisible(t *testing.T) {
	for _, width := range []int{120, 27} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			status := "(search) " + strings.Repeat("中文", 20) + "needle"
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}, frame: func(_ session.Session, w, h int) terminal.Screen {
				return terminal.Screen{Width: w, Height: h, CopyMode: true, Status: status, CopyPrompt: &terminal.CopyPrompt{Text: status, LabelBytes: len("(search) "), CursorBytes: len(status)}}
			}}
			h := newWorkspaceHarness(t, service)
			h.until(func() bool { return strings.Contains(h.m.View().Content, "(search)") })
			h.send(tea.WindowSizeMsg{Width: width, Height: 32})
			view := h.m.View()
			cursor := view.Cursor
			if cursor == nil || cursor.X != min(95, width-1) || cursor.Y != 31 {
				t.Fatalf("native search cursor is not at the visible prompt: %+v", cursor)
			}
			footer := ansi.Strip(strings.Split(view.Content, "\n")[31])
			if !strings.Contains(footer, "(search) ") || !strings.Contains(footer, "needle") {
				t.Fatal("prompt label or input at cursor was clipped")
			}
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			if h.m.View().Cursor != nil {
				t.Fatal("hidden native prompt retained keyboard focus")
			}
		})
	}
}

func TestWorkspaceOldClipboardEventCannotReachNewTab(t *testing.T) {
	service := &workspaceCopyFixture{workspaceFixture: workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	old := h.m.workspace.generation
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker second") })
	_, cmd := h.m.Update(workspaceClipboard{generation: old, text: "old selection"})
	if cmd != nil {
		t.Fatal("old terminal event emitted a host clipboard command")
	}
	_, cmd = h.m.Update(workspaceClipboardSent{token: old})
	if cmd != nil {
		t.Fatal("old copy acknowledgement restarted a reader on the new Tab")
	}
}

func TestWorkspaceNativeCopyCursorIsVisibleWithoutRestoringLiveCursor(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}, frame: func(_ session.Session, width, height int) terminal.Screen {
		return terminal.Screen{Width: width, Height: height, CopyMode: true, CopyCursor: &terminal.Cursor{X: 3, Y: 2, Visible: true}}
	}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.screen.CopyMode })
	cursor := h.m.View().Cursor
	if cursor == nil || cursor.X != h.m.workspaceGeometry().x+3 || cursor.Y != h.m.workspaceGeometry().y+2 || cursor.Blink {
		t.Fatalf("copy cursor not projected independently of hidden live cursor: %+v", cursor)
	}
}
