package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type workspaceProgramPointer struct {
	generation                  uint64
	geometry                    workspaceGeometry
	width, height, lastX, lastY int
	screenWidth, screenHeight   int
}

func (m *Model) programPointerReady() bool {
	w := m.workspace
	_, ok := w.conn.(terminal.PointerConnection)
	return ok && w.focused && !w.busy && w.err == nil && w.transition == workspaceIdle && w.screen.WantsMouse && !w.screen.CopyMode && w.screen.CopyPrompt == nil
}
func (m *Model) programPointerCurrent(op *workspaceProgramPointer) bool {
	w := m.workspace
	return op != nil && op.generation == w.generation && op.geometry == m.workspaceGeometry() && op.screenWidth == w.screen.Width && op.screenHeight == w.screen.Height && m.programPointerReady()
}
func (m *Model) startProgramPointer(msg tea.MouseClickMsg) (tea.Cmd, bool) {
	w := m.workspace
	if w.programPointer != nil {
		return nil, true
	}
	if msg.Button != tea.MouseLeft || msg.Mod != 0 || !m.programPointerReady() {
		return nil, false
	}
	g := m.workspaceGeometry()
	width, height := min(g.width, w.screen.Width), min(g.height, w.screen.Height)
	if msg.X < g.x || msg.X >= g.x+width || msg.Y < g.y || msg.Y >= g.y+height {
		return nil, false
	}
	op := &workspaceProgramPointer{generation: w.generation, geometry: g, width: width, height: height, screenWidth: w.screen.Width, screenHeight: w.screen.Height}
	w.programPointer = op
	before := w.pending
	cmd := m.writeProgramPointer(op, "down", msg.X-g.x, msg.Y-g.y)
	if w.pending == before {
		w.programPointer = nil
	}
	return cmd, true
}
func (m *Model) writeProgramPointer(op *workspaceProgramPointer, action string, x, y int) tea.Cmd {
	x, y = max(0, min(x, op.width-1)), max(0, min(y, op.height-1))
	before := m.workspace.pending
	cmd := m.workspaceWrite(workspaceInput{kind: "pointer", text: action, x: x, y: y})
	if m.workspace.pending > before {
		op.lastX, op.lastY = x, y
	}
	return cmd
}
func (m *Model) moveProgramPointer(msg tea.MouseMotionMsg) (tea.Cmd, bool) {
	w := m.workspace
	op := w.programPointer
	if op == nil {
		return nil, false
	}
	if !m.programPointerCurrent(op) {
		w.programPointer = nil
		return nil, true
	}
	if msg.Mod != 0 || msg.Button != tea.MouseLeft {
		return m.finishProgramPointer(), true
	}
	return m.writeProgramPointer(op, "drag", msg.X-op.geometry.x, msg.Y-op.geometry.y), true
}
func (m *Model) releaseProgramPointer(msg tea.MouseReleaseMsg) (tea.Cmd, bool) {
	w := m.workspace
	op := w.programPointer
	if op == nil {
		return nil, false
	}
	if !m.programPointerCurrent(op) {
		w.programPointer = nil
		return nil, true
	}
	if msg.Button != tea.MouseLeft && msg.Button != tea.MouseNone {
		return nil, true
	}
	w.programPointer = nil
	return m.writeProgramPointer(op, "up", msg.X-op.geometry.x, msg.Y-op.geometry.y), true
}

// End a known program hold before a deliberate UI action, never a copy-mode
// selection. The caller enqueues this before navigation/resize, on the old FIFO.
func (m *Model) finishProgramPointer() tea.Cmd {
	w := m.workspace
	op := w.programPointer
	w.programPointer = nil
	if op == nil || op.generation != w.generation || w.conn == nil || w.err != nil || w.busy || !w.screen.WantsMouse || w.screen.CopyMode || w.screen.CopyPrompt != nil {
		return nil
	}
	return m.writeProgramPointer(op, "up", op.lastX, op.lastY)
}
