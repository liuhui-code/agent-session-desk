package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type workspaceCopyPointer struct {
	generation    uint64
	geometry      workspaceGeometry
	width, height int
}

func (m *Model) copyPointerReady() bool {
	w := m.workspace
	_, capable := w.conn.(terminal.PointerConnection)
	return capable && w.focused && !w.busy && w.err == nil && w.transition == workspaceIdle && w.screen.CopyMode && w.screen.CopyPrompt == nil
}

func (m *Model) startCopyPointer(msg tea.MouseClickMsg) (tea.Cmd, bool) {
	w := m.workspace
	if !m.copyPointerReady() || msg.Mod != 0 || msg.Button != tea.MouseLeft {
		w.copyPointer = nil
		return nil, false
	}
	if w.copyPointer != nil {
		return nil, true // A repeated press is not another selection.
	}
	g := m.workspaceGeometry()
	width, height := min(g.width, w.screen.Width), min(g.height, w.screen.Height)
	if msg.X < g.x || msg.X >= g.x+width || msg.Y < g.y || msg.Y >= g.y+height {
		return nil, false
	}
	w.copyPointer = &workspaceCopyPointer{generation: w.generation, geometry: g, width: w.screen.Width, height: w.screen.Height}
	return m.writeCopyPointer("down", msg.X, msg.Y), true
}

func (m *Model) copyPointerCurrent() bool {
	w := m.workspace
	op := w.copyPointer
	return op != nil && op.generation == w.generation && op.geometry == m.workspaceGeometry() && op.width == w.screen.Width && op.height == w.screen.Height && m.copyPointerReady()
}

func (m *Model) writeCopyPointer(action string, x, y int) tea.Cmd {
	w := m.workspace
	g := w.copyPointer.geometry
	x, y = x-g.x, y-g.y
	if x < -terminal.MaxWidth || x >= terminal.MaxWidth || y < -terminal.MaxHeight || y >= terminal.MaxHeight {
		w.copyPointer = nil
		m.status = "鼠标位置超出终端边界；手势已取消"
		return nil
	}
	before := w.pending
	cmd := m.workspaceWrite(workspaceInput{kind: "pointer", text: action, x: x, y: y})
	if w.pending == before { // A refused queue write cannot start/continue a drag.
		w.copyPointer = nil
	}
	return cmd
}

func (m *Model) moveCopyPointer(msg tea.MouseMotionMsg) (tea.Cmd, bool) {
	w := m.workspace
	if w.copyPointer == nil {
		return nil, false
	}
	if !m.copyPointerCurrent() || msg.Mod != 0 || msg.Button != tea.MouseLeft {
		w.copyPointer = nil
		return nil, true
	}
	return m.writeCopyPointer("drag", msg.X, msg.Y), true
}

func (m *Model) releaseCopyPointer(msg tea.MouseReleaseMsg) (tea.Cmd, bool) {
	w := m.workspace
	if w.copyPointer == nil {
		return nil, false
	}
	defer func() { w.copyPointer = nil }()
	if !m.copyPointerCurrent() || msg.Mod != 0 || (msg.Button != tea.MouseLeft && msg.Button != tea.MouseNone) {
		return nil, true
	}
	return m.writeCopyPointer("up", msg.X, msg.Y), true
}
