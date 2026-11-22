package ui

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type workspaceClipboard struct {
	generation uint64
	text       string
	err        error
}
type workspaceClipboardSent struct{ generation uint64 }

func (m *Model) readWorkspaceClipboard() tea.Cmd {
	w := m.workspace
	if m.clipboardGeneration != 0 || w.clipboardReading {
		return nil
	}
	source, ok := w.conn.(terminal.ClipboardSource)
	if !ok {
		return nil
	}
	generation, ctx := w.generation, w.ctx
	w.clipboardReading = true
	return func() tea.Msg {
		text, err := source.NextClipboard(ctx)
		return workspaceClipboard{generation, text, err}
	}
}

func (m *Model) workspaceClipboardEvent(msg tea.Msg) (tea.Cmd, bool) {
	w := m.workspace
	switch msg := msg.(type) {
	case workspaceClipboard:
		if msg.generation != w.generation || w.conn == nil {
			return nil, true
		}
		w.clipboardReading = false
		if msg.err != nil {
			m.status = "复制通道不可用；请重新连接"
			return nil, true
		}
		if len(msg.text) == 0 || len(msg.text) > terminal.MaxClipboardBytes || !utf8.ValidString(msg.text) || strings.IndexByte(msg.text, 0) >= 0 {
			return m.readWorkspaceClipboard(), true
		}
		generation := w.generation
		m.clipboardGeneration = generation
		// Sequence submits one host copy before accepting the next event.
		// At most one latest pending payload lives in the terminal adapter.
		return tea.Sequence(tea.SetClipboard(msg.text), func() tea.Msg { return workspaceClipboardSent{generation} }), true
	case workspaceClipboardSent:
		if m.clipboardGeneration == 0 || msg.generation != m.clipboardGeneration {
			return nil, true
		}
		m.clipboardGeneration = 0
		if msg.generation == w.generation && w.conn != nil {
			m.status = "已向终端发送复制请求"
		}
		return m.readWorkspaceClipboard(), true
	}
	return nil, false
}
