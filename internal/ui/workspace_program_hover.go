package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type workspaceProgramHover struct{ running, latest *workspaceProgramPointer }
type workspaceProgramHoverDone struct {
	done   workspaceInputDone
	target *workspaceProgramPointer
}

func (m *Model) hoverProgramPointer(msg tea.MouseMotionMsg) (tea.Cmd, bool) {
	w := m.workspace
	_, supported := w.conn.(terminal.ProgramMouseConnection)
	modifiers, valid := programMouseModifiers(msg.Mod)
	if msg.Button != tea.MouseNone || !valid || modifiers != 0 || !supported || !m.programPointerReady() {
		w.programHover.latest = nil
		return nil, false
	}
	g := m.workspaceGeometry()
	width, height := min(g.width, w.screen.Width), min(g.height, w.screen.Height)
	if msg.X < g.x || msg.X >= g.x+width || msg.Y < g.y || msg.Y >= g.y+height {
		w.programHover.latest = nil
		return nil, false
	}
	op := &workspaceProgramPointer{generation: w.generation, geometry: g, width: width, height: height, screenWidth: w.screen.Width, screenHeight: w.screen.Height, lastX: msg.X - g.x, lastY: msg.Y - g.y}
	w.programHover.latest = op
	return m.flushProgramHover(), true
}
func (m *Model) flushProgramHover() tea.Cmd {
	w := m.workspace
	h := &w.programHover
	if h.running != nil || h.latest == nil {
		return nil
	}
	op := h.latest
	h.latest = nil
	if w.programPointer != nil || !m.programPointerCurrent(op) {
		return nil
	}
	before := w.pending
	cmd := m.workspaceWrite(workspaceInput{kind: "program-mouse", text: "move", button: "none", x: op.lastX, y: op.lastY, hover: op})
	if w.pending > before {
		h.running = op
	}
	return cmd
}
func (m *Model) programHoverEvent(msg tea.Msg) (tea.Cmd, bool) {
	r, ok := msg.(workspaceProgramHoverDone)
	if !ok {
		return nil, false
	}
	w := m.workspace
	if r.done.generation != w.generation || w.programHover.running != r.target {
		return nil, true
	}
	w.programHover.running = nil
	cmd, _ := m.workspaceEvent(r.done)
	return tea.Batch(cmd, m.flushProgramHover()), true
}
func (m *Model) validateProgramHover() {
	w := m.workspace
	if w.programHover.latest != nil && !m.programPointerCurrent(w.programHover.latest) {
		w.programHover.latest = nil
	}
}
func workspacePointerReceipt(generation uint64, err error, target *workspaceProgramPointer) tea.Msg {
	done := workspaceInputDone{generation, err}
	if target != nil {
		return workspaceProgramHoverDone{done, target}
	}
	return done
}
