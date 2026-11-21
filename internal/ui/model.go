// Package ui contains the management screen only. The terminal backend owns
// interactive agent I/O; this model never forwards manager keys to an agent.
package ui

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// Service is the narrow boundary between the migrated UI and the session core.
// Implementations must honor cancellation and bound captured output. Attach
// prepares a foreground client; it must not start it or bind its standard I/O.
type Service interface {
	List(context.Context) ([]session.Session, error)
	Preview(context.Context, session.Session, int) (session.Preview, error)
	Attach(context.Context, session.Session) (*exec.Cmd, error)
}

type sessionsMsg struct {
	generation uint64
	sessions   []session.Session
	err        error
}

type previewMsg struct {
	generation uint64
	sessionID  string
	runID      string
	preview    session.Preview
	err        error
}

type attachReadyMsg struct {
	cmd *exec.Cmd
	err error
}

type attachDoneMsg struct{ err error }
type pollMsg struct{ generation uint64 }

// Model implements the stable Bubble Tea v2 API. Mutable UI state is touched
// only by Init/Update; asynchronous commands return values, never mutate it.
type Model struct {
	ctx     context.Context
	cancel  context.CancelFunc
	service Service
	styles  styles

	width, height      int
	sessions           []session.Session
	cursor, viewOffset int
	listBusy           bool
	listGeneration     uint64
	listCancel         context.CancelFunc
	listErr            error
	help               bool
	attaching          bool
	closed             bool
	status             string

	preview           string
	previewTruncated  bool
	previewErr        error
	previewBusy       bool
	previewGeneration uint64
	previewCancel     context.CancelFunc
	pollGeneration    uint64
	pollPending       bool
}

func New(ctx context.Context, service Service) *Model {
	ctx, cancel := context.WithCancel(ctx)
	return &Model{ctx: ctx, cancel: cancel, service: service, styles: newStyles()}
}

// Close releases pending reads. Call it after Program.Run even if the program
// exits because of an external error or signal. It does not stop a session.
func (m *Model) Close() {
	m.closed = true
	m.cancel()
	if m.previewCancel != nil {
		m.previewCancel()
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.loadSessions(), m.schedulePoll())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.closed {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		if !m.previewVisible() {
			m.invalidatePreview()
		}
		return m, m.loadPreview()
	case sessionsMsg:
		m.listBusy = false
		if m.listCancel != nil {
			m.listCancel()
			m.listCancel = nil
		}
		if msg.generation != m.listGeneration || m.attaching {
			return m, m.loadSessions()
		}
		m.listErr = msg.err
		if msg.err != nil {
			m.invalidatePreview()
			return m, nil
		}
		selected, hadSelection := m.selected()
		m.sessions = msg.sessions
		m.cursor = min(m.cursor, max(0, len(m.sessions)-1))
		for i, s := range m.sessions {
			if hadSelection && sameRun(s, selected) {
				m.cursor = i
				break
			}
		}
		next, hasSelection := m.selected()
		if hadSelection != hasSelection || (hasSelection && !sameRun(selected, next)) {
			m.invalidatePreview()
		}
		return m, m.loadPreview()
	case previewMsg:
		// There is at most one outstanding capture. A cancelled capture is
		// allowed to finish before another starts, even when selection changes.
		m.previewBusy = false
		if m.previewCancel != nil {
			m.previewCancel()
			m.previewCancel = nil
		}
		selected, ok := m.selected()
		if msg.generation != m.previewGeneration || !ok ||
			selected.ID != msg.sessionID || selected.RunID != msg.runID ||
			m.attaching || !m.previewVisible() {
			return m, m.loadPreview()
		}
		m.previewErr = msg.err
		if msg.err == nil {
			var truncated bool
			m.preview, truncated = boundedPreview(msg.preview.Text)
			m.previewTruncated = msg.preview.Truncated || truncated
		} else {
			m.preview = ""
			m.previewTruncated = false
		}
	case pollMsg:
		if msg.generation != m.pollGeneration {
			return m, nil
		}
		m.pollPending = false
		return m, tea.Batch(m.loadSessions(), m.loadPreview(), m.schedulePoll())
	case attachReadyMsg:
		if msg.err != nil || msg.cmd == nil {
			if msg.err == nil {
				msg.err = fmt.Errorf("终端后台没有返回 attach 命令")
			}
			return m, m.attachFinished(msg.err)
		}
		return m, tea.ExecProcess(msg.cmd, func(err error) tea.Msg { return attachDoneMsg{err} })
	case attachDoneMsg:
		return m, m.attachFinished(msg.err)
	case tea.KeyPressMsg:
		if m.attaching {
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			m.Close()
			return m, tea.Quit
		case "?", "n":
			m.help = !m.help
			if m.help {
				m.invalidatePreview()
			}
			return m, m.loadPreview()
		case "esc":
			m.help = false
			return m, m.loadPreview()
		case "r":
			m.status = ""
			return m, tea.Batch(m.loadSessions(), m.loadPreview())
		case "up", "k":
			return m, m.move(-1)
		case "down", "j":
			return m, m.move(1)
		case "home", "g":
			return m, m.move(-len(m.sessions))
		case "end", "G":
			return m, m.move(len(m.sessions))
		case "enter":
			if m.help {
				return m, nil
			}
			return m, m.attach()
		}
	}
	return m, nil
}

func sameRun(a, b session.Session) bool { return a.ID == b.ID && a.RunID == b.RunID }

func (m *Model) selected() (session.Session, bool) {
	if m.cursor < 0 || m.cursor >= len(m.sessions) {
		return session.Session{}, false
	}
	return m.sessions[m.cursor], true
}

func (m *Model) move(delta int) tea.Cmd {
	if m.help || len(m.sessions) == 0 {
		return nil
	}
	previous := m.cursor
	m.cursor = min(max(0, m.cursor+delta), len(m.sessions)-1)
	if previous == m.cursor {
		return nil
	}
	m.invalidatePreview()
	return m.loadPreview()
}

func (m *Model) invalidatePreview() {
	m.previewGeneration++
	if m.previewCancel != nil {
		m.previewCancel()
	}
	m.preview, m.previewErr, m.previewTruncated = "", nil, false
}

func (m *Model) loadSessions() tea.Cmd {
	if m.listBusy || m.attaching || m.closed {
		return nil
	}
	m.listBusy = true
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	m.listCancel = cancel
	generation := m.listGeneration
	return func() tea.Msg {
		defer cancel()
		sessions, err := m.service.List(ctx)
		return sessionsMsg{generation: generation, sessions: sessions, err: err}
	}
}

func (m *Model) loadPreview() tea.Cmd {
	s, ok := m.selected()
	if !ok || m.previewBusy || m.attaching || m.closed || m.listErr != nil || !m.previewVisible() {
		return nil
	}
	if s.TerminalState != session.TerminalAvailable && s.TerminalState != session.TerminalExited {
		m.invalidatePreview()
		return nil
	}
	m.previewBusy = true
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	m.previewCancel = cancel
	generation := m.previewGeneration
	lines := min(maxPreviewLines, max(1, m.height))
	return func() tea.Msg {
		defer cancel()
		preview, err := m.service.Preview(ctx, s, lines)
		return previewMsg{generation: generation, sessionID: s.ID, runID: s.RunID, preview: preview, err: err}
	}
}

func (m *Model) previewVisible() bool {
	if m.help || m.width < layoutBreakpointSingle || m.height < 10 {
		return false
	}
	if m.width < layoutBreakpointStacked {
		contentHeight := m.height - 4
		listHeight := max(3, contentHeight/2)
		return contentHeight-listHeight-1 >= 5 // title, caption and one text row
	}
	return true
}

func (m *Model) schedulePoll() tea.Cmd {
	if m.pollPending || m.attaching || m.closed {
		return nil
	}
	m.pollPending = true
	generation := m.pollGeneration
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return pollMsg{generation} })
}

func (m *Model) attach() tea.Cmd {
	s, ok := m.selected()
	if !ok {
		return nil
	}
	if s.TerminalState != session.TerminalAvailable {
		m.status = "目标终端暂不可连接；按 r 重新检查"
		return nil
	}
	m.attaching = true
	m.invalidatePreview()
	m.listGeneration++
	if m.listCancel != nil {
		m.listCancel()
	}
	m.pollGeneration++
	m.pollPending = false
	m.status = "正在连接；Ctrl+b 后按 d 返回管理器"
	ctx := m.ctx
	return func() tea.Msg {
		cmd, err := m.service.Attach(ctx, s)
		return attachReadyMsg{cmd, err}
	}
}

func (m *Model) attachFinished(err error) tea.Cmd {
	m.attaching = false
	if err != nil {
		m.status = "连接结束：" + label(err.Error())
	} else {
		m.status = "已返回管理器"
	}
	return tea.Batch(m.loadSessions(), m.loadPreview(), m.schedulePoll())
}
