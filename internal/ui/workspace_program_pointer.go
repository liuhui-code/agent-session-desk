package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type workspaceProgramPointer struct {
	button                      tea.MouseButton
	buttonName                  string
	modifiers                   uint8
	generation                  uint64
	geometry                    workspaceGeometry
	width, height, lastX, lastY int
	screenWidth, screenHeight   int
}

func (m *Model) programPointerReady() bool {
	w := m.workspace
	_, basic := w.conn.(terminal.PointerConnection)
	_, full := w.conn.(terminal.ProgramMouseConnection)
	return (basic || full) && w.focused && !w.busy && w.err == nil && w.transition == workspaceIdle && w.screen.WantsMouse && !w.screen.CopyMode && w.screen.CopyPrompt == nil
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
	button, ok := programMouseButton(msg.Button)
	modifiers, valid := programMouseModifiers(msg.Mod)
	_, full := w.conn.(terminal.ProgramMouseConnection)
	if !ok || !valid || (!full && (msg.Button != tea.MouseLeft || modifiers != 0)) || !m.programPointerReady() {
		return nil, false
	}
	g := m.workspaceGeometry()
	width, height := min(g.width, w.screen.Width), min(g.height, w.screen.Height)
	if msg.X < g.x || msg.X >= g.x+width || msg.Y < g.y || msg.Y >= g.y+height {
		return nil, false
	}
	op := &workspaceProgramPointer{button: msg.Button, buttonName: button, modifiers: modifiers, generation: w.generation, geometry: g, width: width, height: height, screenWidth: w.screen.Width, screenHeight: w.screen.Height}
	w.programPointer = op
	before := w.pending
	cmd := m.writeProgramPointer(op, "down", msg.X-g.x, msg.Y-g.y, modifiers)
	if w.pending == before {
		w.programPointer = nil
	}
	return cmd, true
}
func (m *Model) writeProgramPointer(op *workspaceProgramPointer, action string, x, y int, modifiers uint8) tea.Cmd {
	x, y = max(0, min(x, op.width-1)), max(0, min(y, op.height-1))
	before := m.workspace.pending
	cmd := m.workspaceWrite(workspaceInput{kind: "program-mouse", text: action, button: op.buttonName, modifiers: modifiers, x: x, y: y})
	if m.workspace.pending > before {
		op.lastX, op.lastY, op.modifiers = x, y, modifiers
	}
	return cmd
}
func (m *Model) moveProgramPointer(msg tea.MouseMotionMsg) (tea.Cmd, bool) {
	w := m.workspace
	op := w.programPointer
	if op == nil {
		return m.hoverProgramPointer(msg)
	}
	if !m.programPointerCurrent(op) {
		w.programPointer = nil
		return nil, true
	}
	modifiers, valid := programMouseModifiers(msg.Mod)
	_, full := w.conn.(terminal.ProgramMouseConnection)
	if !valid || msg.Button != op.button || (!full && modifiers != 0) {
		return m.finishProgramPointer(), true
	}
	return m.writeProgramPointer(op, "drag", msg.X-op.geometry.x, msg.Y-op.geometry.y, modifiers), true
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
	if msg.Button != op.button && msg.Button != tea.MouseNone {
		return nil, true
	}
	modifiers, valid := programMouseModifiers(msg.Mod)
	_, full := w.conn.(terminal.ProgramMouseConnection)
	if !valid || (!full && modifiers != 0) {
		return m.finishProgramPointer(), true
	}
	w.programPointer = nil
	return m.writeProgramPointer(op, "up", msg.X-op.geometry.x, msg.Y-op.geometry.y, modifiers), true
}

// End a known program hold before a deliberate UI action, never a copy-mode
// selection. The caller enqueues this before navigation/resize, on the old FIFO.
func (m *Model) finishProgramPointer() tea.Cmd {
	w := m.workspace
	op := w.programPointer
	w.programHover.latest = nil
	w.programPointer = nil
	if op == nil || op.generation != w.generation || w.conn == nil || w.err != nil || w.busy || !w.screen.WantsMouse || w.screen.CopyMode || w.screen.CopyPrompt != nil {
		return nil
	}
	return m.writeProgramPointer(op, "up", op.lastX, op.lastY, op.modifiers)
}

func programMouseButton(button tea.MouseButton) (string, bool) {
	switch button {
	case tea.MouseLeft:
		return "left", true
	case tea.MouseMiddle:
		return "middle", true
	case tea.MouseRight:
		return "right", true
	}
	return "", false
}
func programMouseModifiers(mod tea.KeyMod) (uint8, bool) {
	known := tea.ModShift | tea.ModAlt | tea.ModCtrl
	locks := tea.ModCapsLock | tea.ModNumLock | tea.ModScrollLock
	if mod & ^(known|locks) != 0 {
		return 0, false
	}
	var result uint8
	if mod&tea.ModShift != 0 {
		result |= terminal.MouseShift
	}
	if mod&tea.ModAlt != 0 {
		result |= terminal.MouseAlt
	}
	if mod&tea.ModCtrl != 0 {
		result |= terminal.MouseCtrl
	}
	return result, true
}
