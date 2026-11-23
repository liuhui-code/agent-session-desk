package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type workspaceService interface {
	OpenTerminal(context.Context, session.Session, int, int) (terminal.Connection, error)
}
type workspaceTab struct {
	viewID  uint64
	session session.Session
	live    bool
}
type workspaceTransition uint8

const (
	workspaceIdle workspaceTransition = iota
	workspaceSwitch
	workspaceQuit
	workspaceClose
	workspaceHistory
	workspaceHelp
	workspaceAttach
	workspaceLaunch
	workspaceSettings
	workspaceSearch
	workspaceRename
	workspaceGroup
	workspaceRemoval
	workspaceStop
)

type workspaceState struct {
	directories                directoryState
	programPointer             *workspaceProgramPointer
	copyPointer                *workspaceCopyPointer
	visibleCopy                *visiblePaneCopy
	pendingClipboard           *workspaceClipboard
	clipboardReading           bool
	tabDrag                    workspaceTabDrag
	listClick                  workspaceListClick
	split                      workspaceSplit
	readBusy                   bool
	transition                 workspaceTransition
	pending                    int
	switchTarget               uint64
	buffered                   []workspaceInput
	inputs                     chan workspaceInput
	service                    workspaceService
	tabs                       []workspaceTab
	active                     int
	nextID                     uint64
	initialized, focused, busy bool
	generation                 uint64
	cancel                     context.CancelFunc
	ctx                        context.Context
	conn                       terminal.Connection
	screen                     terminal.Screen
	err                        error
}
type workspaceOpened struct {
	generation    uint64
	conn          terminal.Connection
	err           error
	width, height int
}
type workspaceFrame struct {
	generation uint64
	screen     terminal.Screen
	err        error
}
type workspaceResized struct {
	generation uint64
	err        error
}

func (w *workspaceState) disconnect() {
	w.programPointer = nil
	w.copyPointer = nil
	if w.visibleCopy != nil {
		w.visibleCopy.cancel()
		w.visibleCopy = nil
	}
	w.pendingClipboard = nil
	w.clipboardReading = false
	w.tabDrag = workspaceTabDrag{}
	w.split.dragging = false
	w.split.hovered = false
	w.generation++
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	if w.conn != nil {
		_ = w.conn.Close()
		w.conn = nil
	}
	w.busy = false
	w.readBusy = false
	w.inputs = nil
	w.pending = 0
}

func (m *Model) workspaceSessions() tea.Cmd {
	w := m.workspace
	if !w.initialized && len(m.sessions) > 0 {
		for _, s := range m.sessions {
			w.nextID++
			w.tabs = append(w.tabs, workspaceTab{w.nextID, s, true})
		}
		w.initialized = true
	}
	for i := range w.tabs {
		tab := &w.tabs[i]
		tab.live = false
		for _, s := range m.sessions {
			if sameRun(s, tab.session) {
				tab.session = s
				tab.live = true
				break
			}
		}
	}
	if len(w.tabs) == 0 {
		return nil
	}
	tab := w.tabs[w.active]
	if !tab.live || tab.session.TerminalState != session.TerminalAvailable {
		m.failWorkspace(fmt.Errorf("目标终端当前不可交互；保留视图，不会自动重启"))
		return nil
	}
	if w.conn == nil && !w.busy && w.err == nil {
		return m.connectWorkspace()
	}
	return nil
}

func (m *Model) connectWorkspace() tea.Cmd {
	w := m.workspace
	if len(w.tabs) == 0 || m.closed || m.attaching || m.history.open || m.help || m.launch != nil || m.densitySettings != nil || m.search != nil || m.rename != nil || m.groupDialog != nil || m.groupRemoval != nil || m.sessionStop != nil {
		return nil
	}
	g := m.workspaceGeometry()
	if g.width < 1 || g.height < 1 {
		return nil
	}
	w.disconnect()
	w.screen = terminal.Screen{}
	w.err = nil
	w.busy = true
	w.ctx, w.cancel = context.WithCancel(m.ctx)
	generation, ctx, service, target := w.generation, w.ctx, w.service, w.tabs[w.active].session
	return func() tea.Msg {
		conn, err := service.OpenTerminal(ctx, target, g.width, g.height)
		if ctx.Err() != nil && conn != nil {
			_ = conn.Close()
			conn = nil
			err = ctx.Err()
		}
		return workspaceOpened{generation: generation, conn: conn, err: err, width: g.width, height: g.height}
	}
}

func (m *Model) workspaceSize() tea.Cmd {
	w := m.workspace
	release := m.finishProgramPointer()
	w.copyPointer = nil
	w.split.hovered = false
	if w.conn == nil {
		if !w.busy && w.err == nil {
			return tea.Batch(release, m.workspaceSessions())
		}
		return release
	}
	g := m.workspaceGeometry()
	if g.width < 1 || g.height < 1 {
		return release
	}
	return tea.Batch(release, m.workspaceWrite(workspaceInput{kind: "resize", width: g.width, height: g.height}))
}

func (m *Model) readWorkspace() tea.Cmd {
	w := m.workspace
	w.readBusy = true
	generation, conn, ctx := w.generation, w.conn, w.ctx
	return func() tea.Msg { screen, err := conn.Next(ctx); return workspaceFrame{generation, screen, err} }
}

func (m *Model) workspaceEvent(msg tea.Msg) (tea.Cmd, bool) {
	w := m.workspace
	if cmd, handled := m.workspaceClipboardEvent(msg); handled {
		return cmd, true
	}
	switch msg := msg.(type) {
	case workspaceSplitTick:
		return m.workspaceSplitExpired(), true
	case tea.MouseWheelMsg:
		if !m.attaching && !m.history.open && !m.help && m.launch == nil {
			return m.workspaceWheel(msg), true
		}
		return nil, true
	case tea.MouseMotionMsg:
		if !m.attaching && !m.history.open && !m.help && m.launch == nil {
			if cmd, handled := m.moveProgramPointer(msg); handled {
				return cmd, true
			}
			if cmd, handled := m.moveCopyPointer(msg); handled {
				return cmd, true
			}
			if m.workspaceTabMotion(msg) {
				return nil, true
			}
			return m.workspaceSplitMotion(msg), true
		}
		return nil, true
	case tea.MouseReleaseMsg:
		if !m.attaching && !m.history.open && !m.help && m.launch == nil {
			if cmd, handled := m.releaseProgramPointer(msg); handled {
				return cmd, true
			}
			if cmd, handled := m.releaseCopyPointer(msg); handled {
				return cmd, true
			}
			if m.workspaceTabRelease(msg) {
				return nil, true
			}
			return m.workspaceSplitRelease(msg), true
		}
		return nil, true
	case tea.MouseClickMsg:
		if !m.attaching && !m.history.open && !m.help && m.launch == nil {
			return m.workspaceClick(msg), true
		}
		return nil, true
	case tea.PasteMsg:
		release := m.finishProgramPointer()
		w.copyPointer = nil
		w.tabDrag = workspaceTabDrag{}
		if !m.attaching && !m.history.open && !m.help && w.focused && len(w.tabs) > 0 && w.transition != workspaceQuit && w.transition != workspaceClose {
			return tea.Batch(release, m.workspaceWrite(workspaceInput{kind: "paste", text: msg.Content})), true
		}
		return release, true
	case tea.KeyPressMsg:
		if m.attaching {
			return nil, true
		}
		if !m.history.open && !m.help {
			return m.workspaceKey(msg), true
		}

	case workspaceInputDone:
		if msg.generation != w.generation {
			return nil, true
		}
		if w.pending > 0 {
			w.pending--
		}
		if msg.err != nil {
			m.failWorkspace(msg.err)
			return nil, true
		}
		if w.pending == 0 && w.transition >= workspaceHistory {
			return m.finishWorkspaceNavigation(), true
		}
		if w.pending == 0 && m.launch != nil {
			w.disconnect()
			return nil, true
		}
		if w.pending == 0 && w.transition == workspaceClose {
			return m.closeWorkspaceTabNow(), true
		}
		if w.pending == 0 && w.transition == workspaceQuit {
			if m.preferences.busy {
				return nil, true
			}
			m.Close()
			return tea.Quit, true
		}
		if w.pending == 0 && w.transition == workspaceSwitch && !w.busy {
			for i, tab := range w.tabs {
				if tab.viewID == w.switchTarget {
					return m.activateWorkspaceNow(i), true
				}
			}
		}
		return nil, true

	case workspaceOpened:
		if msg.generation != w.generation {
			if msg.conn != nil {
				_ = msg.conn.Close()
			}
			return nil, true
		}
		w.busy = false
		w.err = msg.err
		if msg.err != nil {
			if msg.conn != nil {
				_ = msg.conn.Close()
			}
			m.failWorkspace(msg.err)
			return nil, true
		}
		if msg.err == nil && msg.conn != nil {
			w.conn = msg.conn
			w.inputs = make(chan workspaceInput, workspaceInputCapacity)
			startWorkspaceWriter(w.ctx, w.conn, w.inputs)
			if w.transition == workspaceSwitch {
				w.transition = workspaceIdle
			}
			buffered := w.buffered
			w.buffered = nil
			var commands []tea.Cmd
			g := m.workspaceGeometry()
			if g.width != msg.width || g.height != msg.height {
				commands = append(commands, m.workspaceSize())
			}
			if m.launch == nil {
				commands = append(commands, m.readWorkspace())
			}
			commands = append(commands, m.readWorkspaceClipboard())
			for _, action := range buffered {
				commands = append(commands, m.workspaceWrite(action))
			}
			if m.launch != nil && w.pending == 0 {
				w.disconnect()
			}
			if w.pending == 0 && w.transition >= workspaceHistory {
				commands = append(commands, m.finishWorkspaceNavigation())
			}
			return tea.Batch(commands...), true
		}
		return nil, true
	case workspaceFrame:
		if msg.generation != w.generation {
			return nil, true
		}
		if msg.err != nil {
			m.failWorkspace(msg.err)
			return nil, true
		}
		w.readBusy = false
		w.screen = msg.screen
		if w.programPointer != nil && !m.programPointerCurrent(w.programPointer) {
			w.programPointer = nil
			m.status = "程序鼠标手势已失效；释放结果未知，请核对原程序"
		}
		if w.copyPointer != nil && w.screen.CopyMode {
			w.copyPointer.observed = true
		}
		if w.copyPointer != nil && !m.copyPointerCurrent() {
			w.copyPointer = nil
		}
		if m.launch != nil {
			return nil, true
		}
		return m.readWorkspace(), true
	case workspaceResized:
		if msg.generation == w.generation && msg.err != nil {
			m.failWorkspace(msg.err)
		}
		return nil, true
	}
	return nil, false
}

func (m *Model) activateWorkspaceNow(index int) tea.Cmd {
	w := m.workspace
	if len(w.tabs) == 0 {
		return nil
	}
	index = (index + len(w.tabs)) % len(w.tabs)
	if index == w.active && w.conn != nil {
		return nil
	}
	w.disconnect()
	w.active = index
	w.err = nil
	m.status = ""
	m.selectWorkspaceSession(w.tabs[index].session)
	cmd := m.connectWorkspace()
	if !w.busy {
		w.transition = workspaceIdle
	}
	return cmd
}

func (m *Model) activateWorkspace(index int) tea.Cmd {
	w := m.workspace
	if len(w.tabs) == 0 || w.transition != workspaceIdle {
		return nil
	}
	if w.busy {
		m.status = "当前 Tab 正在连接；连接完成后可切换"
		return nil
	}
	index = (index + len(w.tabs)) % len(w.tabs)
	if index == w.active && w.conn != nil {
		return nil
	}
	w.transition = workspaceSwitch
	w.switchTarget = w.tabs[index].viewID
	m.selectWorkspaceSession(w.tabs[index].session)
	if w.pending > 0 {
		m.status = "正在完成原终端输入后切换 Tab…"
		return nil
	}
	return m.activateWorkspaceNow(index)
}

func (m *Model) selectWorkspaceSession(target session.Session) {
	for i, s := range m.sessions {
		if sameRun(s, target) {
			m.cursor = i
			m.revealSidebarSession()
			return
		}
	}
}
