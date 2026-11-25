package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/pathsuggest"
)

type recentDirectorySaved struct{ err error }

func (m *Model) rememberDirectory(dir string) tea.Cmd {
	s := &m.workspace.directories
	if s.save == nil || !pathsuggest.Safe(dir) {
		return nil
	}
	s.wanted = pathsuggest.Filter(append([]string{dir}, s.wanted...), "", nil)
	if s.writing {
		return nil
	}
	return m.writeRecentDirectory()
}
func (m *Model) writeRecentDirectory() tea.Cmd {
	s := &m.workspace.directories
	if s.writing || len(s.wanted) == 0 || m.ctx.Err() != nil {
		return nil
	}
	dirs := append([]string(nil), s.wanted[:min(len(s.wanted), config.MaxRecentDirs)]...)
	s.wanted = nil
	s.writing = true
	save, ctx := s.save, m.ctx
	return func() tea.Msg { return recentDirectorySaved{save(ctx, dirs)} }
}
func (m *Model) recentDirectoryEvent(msg tea.Msg) (tea.Cmd, bool) {
	r, ok := msg.(recentDirectorySaved)
	if !ok {
		return nil, false
	}
	s := &m.workspace.directories
	s.writing = false
	if r.err != nil {
		m.status = fmt.Sprintf("Terminal created; recent directory not saved: %s", label(r.err.Error()))
	}
	return m.writeRecentDirectory(), true
}
