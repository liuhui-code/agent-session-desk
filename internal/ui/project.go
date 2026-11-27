package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type projectState struct {
	load            func(context.Context) ([]session.Launch, error)
	busy, uncertain bool
}
type projectResult struct {
	count int
	err   error
}

func (m *Model) startProject() tea.Cmd {
	if _, ok := m.service.(launchService); !ok {
		return nil
	}
	if m.project.uncertain {
		m.status = "Inspect sessions before reopening workspace"
		return nil
	}
	if m.project.busy || m.project.load == nil {
		return nil
	}
	if m.workspace.pending > 0 {
		m.status = "Finish input before opening workspace"
		return nil
	}
	m.project.busy = true
	m.status = "Opening workspace…"
	load, ctx, service := m.project.load, m.ctx, m.service.(launchService)
	return func() tea.Msg {
		rows, e := load(ctx)
		if e != nil {
			return projectResult{count: -1, err: e}
		}
		n := 0
		for _, l := range rows {
			if _, e = service.Start(ctx, l); e != nil {
				return projectResult{n, e}
			}
			n++
		}
		return projectResult{count: n}
	}
}
func (m *Model) projectEvent(msg tea.Msg) (tea.Cmd, bool) {
	v, ok := msg.(projectResult)
	if !ok {
		return nil, false
	}
	m.project.busy = false
	m.status = fmt.Sprintf("Opened %d sessions", v.count)
	if v.err != nil {
		if v.count >= 0 {
			m.project.uncertain = true
		}
		m.status = "Workspace incomplete; inspect sessions before retry"
	}
	return m.loadSessions(), true
}
