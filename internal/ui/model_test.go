package ui

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type fakeService struct {
	history func(context.Context, int) ([]session.HistoryRecord, error)
	list    func(context.Context) ([]session.Session, error)
	preview func(context.Context, session.Session, int) (session.Preview, error)
	attach  func(context.Context, session.Session) (*exec.Cmd, error)
}

func (s fakeService) List(ctx context.Context) ([]session.Session, error) {
	if s.list != nil {
		return s.list(ctx)
	}
	return nil, nil
}

func (s fakeService) Preview(ctx context.Context, item session.Session, lines int) (session.Preview, error) {
	if s.preview != nil {
		return s.preview(ctx, item, lines)
	}
	return session.Preview{}, nil
}

func (s fakeService) Attach(ctx context.Context, item session.Session) (*exec.Cmd, error) {
	if s.attach != nil {
		return s.attach(ctx, item)
	}
	return nil, errors.New("fake attach is unavailable")
}

func sampleSessions() []session.Session {
	return []session.Session{
		{ID: "first", RunID: "first-run", BackendID: "terminal-a", Title: "开发", TerminalState: session.TerminalAvailable},
		{ID: "second", RunID: "second-run", BackendID: "terminal-b", Title: "评审", TerminalState: session.TerminalAvailable},
	}
}

func testModel(t *testing.T, service Service) *Model {
	t.Helper()
	m := New(context.Background(), service)
	m.width, m.height = 100, 24
	m.sessions = sampleSessions()
	t.Cleanup(m.Close)
	return m
}

func TestExitedPaneCanBeInspectedWithoutAttaching(t *testing.T) {
	attached := false
	m := testModel(t, fakeService{
		preview: func(context.Context, session.Session, int) (session.Preview, error) {
			return session.Preview{Text: "process exited: retained output"}, nil
		},
		attach: func(context.Context, session.Session) (*exec.Cmd, error) {
			attached = true
			return nil, nil
		},
	})
	m.sessions[0].TerminalState = session.TerminalExited
	command := m.loadPreview()
	if command == nil {
		t.Fatal("retained terminal screen cannot be inspected")
	}
	m.Update(command())
	frame := m.renderPreviewPane(80, 20)
	if !strings.Contains(frame, "retained output") || !strings.Contains(frame, "不可输入") {
		t.Fatalf("retained output or exit boundary missing: %q", frame)
	}
	_, command = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if command != nil || attached {
		t.Fatal("inspecting an exited pane attempted to attach or restart it")
	}
}

// Covers the stale-selection portion of TC-UI-007/TC-UI-009 and FL-005.
// This is a deterministic unit check, not a native Windows result.
func TestSelectionCancelsCaptureAndRejectsLatePreview(t *testing.T) {
	started := make(chan context.Context, 1)
	release := make(chan struct{})
	result := make(chan tea.Msg, 1)
	var requested []string
	m := testModel(t, fakeService{preview: func(ctx context.Context, s session.Session, _ int) (session.Preview, error) {
		requested = append(requested, s.ID)
		if s.ID == "first" {
			started <- ctx
			<-release // deliberately return old output even after cancellation
			return session.Preview{Text: "old first session output"}, nil
		}
		return session.Preview{Text: "current second session output"}, nil
	}})
	first := m.loadPreview()
	go func() { result <- first() }()
	readContext := <-started
	_, next := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if next != nil {
		t.Fatal("started a second capture before cancelled capture finished")
	}
	select {
	case <-readContext.Done():
	default:
		t.Fatal("selection change did not cancel old capture")
	}
	close(release)
	_, next = m.Update(<-result)
	if m.preview != "" {
		t.Fatal("late output became the new session's preview")
	}
	if next == nil {
		t.Fatal("new selected session was not captured after old capture finished")
	}
	m.Update(next())
	if m.preview != "current second session output" {
		t.Fatalf("wrong preview: %q", m.preview)
	}
	if len(requested) != 2 || requested[0] != "first" || requested[1] != "second" {
		t.Fatalf("wrong capture targets: %v", requested)
	}
}

func TestRestartedRunCannotReceiveOldPreview(t *testing.T) {
	m := testModel(t, fakeService{})
	generation := m.previewGeneration
	updated := sampleSessions()
	updated[0].RunID = "new-run"
	m.Update(sessionsMsg{sessions: updated})
	m.Update(previewMsg{generation: generation, sessionID: "first", runID: "first-run", preview: session.Preview{Text: "old"}})
	if m.preview != "" {
		t.Fatal("old Run snapshot was displayed for the restarted Run")
	}
}

func TestReorderingPreservesStableSelection(t *testing.T) {
	m := testModel(t, fakeService{})
	m.cursor = 1
	items := sampleSessions()
	m.Update(sessionsMsg{sessions: []session.Session{items[1], items[0]}})
	s, ok := m.selected()
	if !ok || s.ID != "second" || m.cursor != 0 {
		t.Fatalf("selection followed row position instead of stable identity: %+v", s)
	}
}

func TestEmptyListAndPastedKeysNeverAttachOrQuit(t *testing.T) {
	attached := false
	m := testModel(t, fakeService{attach: func(context.Context, session.Session) (*exec.Cmd, error) {
		attached = true
		return nil, nil
	}})
	m.sessions = nil
	_, command := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if command != nil || attached {
		t.Fatal("empty list created an attach command")
	}
	m.Update(tea.PasteMsg{Content: "q\n\x1b\nj"})
	if m.closed || m.attaching {
		t.Fatal("pasted text was interpreted as manager commands")
	}
}

func TestV2KeyReleaseAndRepeatedEnterDoNotDuplicateActions(t *testing.T) {
	m := testModel(t, fakeService{})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m.Update(tea.KeyReleaseMsg{Code: tea.KeyDown})
	if m.cursor != 1 {
		t.Fatalf("one key press/release moved to cursor %d", m.cursor)
	}
	_, first := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	_, repeated := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	_, release := m.Update(tea.KeyReleaseMsg{Code: tea.KeyEnter})
	if first == nil || repeated != nil || release != nil {
		t.Fatal("single Enter generated duplicate foreground client commands")
	}
}

func TestPreviewFailurePreservesManagementKeys(t *testing.T) {
	m := testModel(t, fakeService{})
	m.Update(previewMsg{generation: m.previewGeneration, sessionID: "first", runID: "first-run", err: errors.New("capture failed")})
	if m.previewErr == nil {
		t.Fatal("capture error was hidden")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.cursor != 1 {
		t.Fatal("capture failure blocked session selection")
	}
	m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if !m.closed {
		t.Fatal("capture failure blocked quit")
	}
}

func TestAttachFreezesTargetAndDiscardsPreAttachList(t *testing.T) {
	var attached session.Session
	m := testModel(t, fakeService{attach: func(_ context.Context, s session.Session) (*exec.Cmd, error) {
		attached = s
		return nil, errors.New("test detach error")
	}})
	oldGeneration := m.listGeneration
	_, command := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if command == nil || !m.attaching {
		t.Fatal("Enter did not prepare attach")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m.Update(sessionsMsg{generation: oldGeneration, sessions: nil})
	if len(m.sessions) != 2 || m.cursor != 0 {
		t.Fatal("selection changed while preparing the foreground terminal")
	}
	m.Update(command())
	if attached.ID != "first" || attached.RunID != "first-run" {
		t.Fatalf("wrong attach target: %+v", attached)
	}
	if m.attaching || m.status == "" {
		t.Fatal("failed attach did not restore manager state with a visible error")
	}
	// A list captured before attach must also be rejected if it arrives later.
	m.Update(sessionsMsg{generation: oldGeneration, sessions: nil})
	if len(m.sessions) != 2 {
		t.Fatal("late pre-attach list overwrote reconnected state")
	}
}

func TestHiddenPreviewDoesNotPoll(t *testing.T) {
	m := testModel(t, fakeService{})
	for _, size := range [][2]int{{49, 24}, {120, 8}, {50, 14}, {0, 0}} {
		m.width, m.height = size[0], size[1]
		if m.loadPreview() != nil {
			t.Fatalf("preview captured while hidden at %v", size)
		}
	}
	m.width, m.height, m.help = 120, 30, true
	if m.loadPreview() != nil {
		t.Fatal("preview captured behind help screen")
	}
}

func TestQuitCancelsWorkWithoutStoppingSession(t *testing.T) {
	m := testModel(t, fakeService{})
	_, command := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if command == nil || !m.closed {
		t.Fatal("quit did not terminate manager")
	}
	select {
	case <-m.ctx.Done():
	default:
		t.Fatal("quit left pending backend reads uncancelled")
	}
	// Service deliberately has no Stop method. Quitting cannot stop a Run.
}

func (s fakeService) History(ctx context.Context, limit int) ([]session.HistoryRecord, error) {
	if s.history != nil {
		return s.history(ctx, limit)
	}
	return nil, nil
}
