// Package ui provides Tab workspaces and management views. Terminal input is
// forwarded only while the selected native terminal owns focus.
package ui

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// Service is the narrow boundary between the migrated UI and the session core.
// Implementations must honor cancellation and bound captured output. Attach
// prepares a foreground client; it must not start it or bind its standard I/O.
type Service interface {
	History(context.Context, int) ([]session.HistoryRecord, error)
	List(context.Context) ([]session.Session, error)
	Preview(context.Context, session.Session, int) (session.Preview, error)
	Attach(context.Context, session.Session) (*exec.Cmd, error)
}

type sessionsMsg struct {
	generation uint64
	sessions   []session.Session
	groups     []session.Group
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
	groupDialog            *groupDialog
	groupIntent            *groupDialog
	launchGroup            session.Group
	groups                 []session.Group
	selectedGroupID        string
	collapsedGroups        map[string]bool
	preferences            uiPreferenceState
	saveUISettings         func(context.Context, config.UISettings) error
	sidebarMode            string
	sidebarCollapsed       bool
	rename                 *sessionRename
	renameTarget           *session.Session
	search                 *sessionSearch
	clipboardGeneration    uint64
	sidebarDensity         string
	saveSidebarDensity     func(context.Context, string) error
	sidebarPreferenceError error
	densitySettings        *densitySettings
	launch                 *launchDialog
	ctx                    context.Context
	cancel                 context.CancelFunc
	service                Service
	styles                 styles
	tokens                 designTokens
	history                historyState
	workspace              *workspaceState

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
	return newModelWithTokens(ctx, service, defaultDesignTokens())
}

func newModelWithTokens(ctx context.Context, service Service, tokens designTokens) *Model {
	ctx, cancel := context.WithCancel(ctx)
	m := &Model{ctx: ctx, cancel: cancel, service: service, tokens: tokens, styles: newStyles(tokens)}
	m.sidebarDensity = "compact"
	m.sidebarMode = "grouped"
	m.preferences.saved = config.UISettings{SidebarDensity: "compact", SidebarMode: "grouped"}
	m.preferences.wanted = m.preferences.saved
	if service, ok := service.(workspaceService); ok {
		capability, explicit := service.(interface{ SupportsTerminal() bool })
		if !explicit || capability.SupportsTerminal() {
			m.workspace = &workspaceState{service: service, focused: true}
		}
	}
	return m
}

// Close releases pending reads. Call it after Program.Run even if the program
// exits because of an external error or signal. It does not stop a session.
func (m *Model) Close() {
	m.closed = true
	if m.workspace != nil {
		m.workspace.disconnect()
		m.workspace.buffered = nil
	}
	m.launch = nil
	m.groupDialog, m.groupIntent = nil, nil
	m.search = nil
	m.rename = nil
	m.renameTarget = nil
	m.densitySettings = nil
	m.cancelHistory()
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
	if m.workspace != nil {
		if cmd, handled := m.groupDialogEvent(msg); handled {
			return m, cmd
		}
		if cmd, handled := m.uiPreferencesEvent(msg); handled {
			return m, cmd
		}
		if cmd, handled := m.sessionRenameEvent(msg); handled {
			return m, cmd
		}
		if cmd, handled := m.sessionSearchEvent(msg); handled {
			return m, cmd
		}
		if cmd, handled := m.densitySettingsEvent(msg); handled {
			return m, cmd
		}
		if cmd, handled := m.launchEvent(msg); handled {
			return m, cmd
		}
		if cmd, handled := m.workspaceEvent(msg); handled {
			return m, cmd
		}
	}
	switch msg := msg.(type) {
	case historyMsg:
		return m, m.historyResult(msg)
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		if m.workspace != nil {
			m.workspace.tabDrag = workspaceTabDrag{}
			m.workspace.split.dragging = false
			m.workspace.listClick = workspaceListClick{}
			return m, m.workspaceSize()
		}
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
		m.groups = msg.groups
		m.syncSessionSearch(true)
		m.cursor = min(m.cursor, max(0, len(m.sessions)-1))
		for i, s := range m.sessions {
			if hadSelection && sameRun(s, selected) {
				m.cursor = i
				break
			}
		}
		m.reconcileSidebarSelection(selected, hadSelection)
		next, hasSelection := m.selected()
		if hadSelection != hasSelection || (hasSelection && !sameRun(selected, next)) {
			m.invalidatePreview()
		}
		if m.workspace != nil {
			if len(m.workspace.tabs) == 0 && len(m.sessions) == 0 {
				m.workspace.focused = false
			}
			return m, m.workspaceSessions()
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
		return m, tea.Batch(m.loadSessions(), m.loadPreview(), m.refreshWorkspace(), m.schedulePoll())
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
		if m.history.open && msg.String() != "q" && msg.String() != "ctrl+c" {
			return m, m.historyKey(msg.String())
		}
		switch msg.String() {
		case "q", "ctrl+c":
			m.Close()
			return m, tea.Quit
		case "h":
			return m, m.openHistory()
		case "?", "n":
			m.help = !m.help
			if m.help {
				m.invalidatePreview()
			}
			if m.workspace != nil && !m.help {
				return m, m.connectWorkspace()
			}
			return m, m.loadPreview()
		case "esc":
			m.help = false
			if m.workspace != nil {
				return m, m.connectWorkspace()
			}
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
	if m.help {
		return nil
	}
	if m.groupedSidebar() {
		return m.moveGroupedSidebar(delta)
	}
	if len(m.sessions) == 0 {
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
		if service, ok := m.service.(groupService); ok {
			view, err := service.Workspace(ctx)
			return sessionsMsg{generation: generation, sessions: view.Sessions, groups: view.Groups, err: err}
		}
		sessions, err := m.service.List(ctx)
		return sessionsMsg{generation: generation, sessions: sessions, err: err}
	}
}

func (m *Model) loadPreview() tea.Cmd {
	if m.workspace != nil {
		return nil
	}
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
	if m.history.open || m.help || m.width < m.tokens.layout.singleBelow || m.height < m.tokens.layout.minPreviewHeight {
		return false
	}
	if m.width < m.tokens.layout.stackedBelow {
		contentHeight := m.height - m.tokens.chromeRows()
		listHeight := max(m.tokens.layout.stackedListMin, contentHeight/m.tokens.layout.stackedFraction)
		return contentHeight-listHeight-m.tokens.layout.stackedGap >= m.tokens.layout.previewMin // title, caption and one text row
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
	m.status = "正在连接；切至英文，Ctrl+b 后按 d 返回管理器"
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
