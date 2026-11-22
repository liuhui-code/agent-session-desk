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
)

type workspaceState struct {
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
	w.split.dragging = false
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
	if len(w.tabs) == 0 || m.closed || m.attaching || m.history.open || m.help || m.launch != nil || m.densitySettings != nil {
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
	if w.conn == nil {
		if !w.busy && w.err == nil {
			return m.workspaceSessions()
		}
		return nil
	}
	g := m.workspaceGeometry()
	if g.width < 1 || g.height < 1 {
		return nil
	}
	return m.workspaceWrite(workspaceInput{kind: "resize", width: g.width, height: g.height})
}

func (m *Model) readWorkspace() tea.Cmd {
	w := m.workspace
	w.readBusy = true
	generation, conn, ctx := w.generation, w.conn, w.ctx
	return func() tea.Msg { screen, err := conn.Next(ctx); return workspaceFrame{generation, screen, err} }
}

func (m *Model) workspaceEvent(msg tea.Msg) (tea.Cmd, bool) {
	w := m.workspace
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
			return m.workspaceSplitMotion(msg), true
		}
		return nil, true
	case tea.MouseReleaseMsg:
		if !m.attaching && !m.history.open && !m.help && m.launch == nil {
			return m.workspaceSplitRelease(msg), true
		}
		return nil, true
	case tea.MouseClickMsg:
		if !m.attaching && !m.history.open && !m.help && m.launch == nil {
			return m.workspaceClick(msg), true
		}
		return nil, true
	case tea.PasteMsg:
		if !m.attaching && !m.history.open && !m.help && w.focused && len(w.tabs) > 0 && w.transition != workspaceQuit && w.transition != workspaceClose {
			return m.workspaceWrite(workspaceInput{kind: "paste", text: msg.Content}), true
		}
		return nil, true
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
			return
		}
	}
}
