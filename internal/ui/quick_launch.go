package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type quickLaunchState struct{ quitting bool }
type quickLaunchResult struct {
	request          *quickLaunchState
	session          session.Session
	dir              string
	err, rememberErr error
}

func (m *Model) startQuickLaunch() tea.Cmd {
	if m.quickCreate != nil {
		return nil
	}
	if m.quickUncertain {
		m.status = "Launch outcome unknown; verify sessions before reopening the UI. No retry sent."
		return m.connectWorkspace()
	}
	service, ok := m.service.(launchService)
	if !ok {
		m.status = "This backend cannot create terminals"
		return m.connectWorkspace()
	}
	request := &quickLaunchState{}
	m.quickCreate = request
	load, save, ctx := m.loadQuickLaunch, m.rememberLaunch, m.ctx
	group := m.launchGroup.ID
	m.launchGroup = session.Group{}
	m.status = "Creating terminal…"
	return func() tea.Msg {
		result := quickLaunchResult{request: request}
		if result.err = ctx.Err(); result.err != nil {
			return result
		}
		l, err := load(ctx)
		if err != nil {
			result.err = fmt.Errorf("Launch settings unavailable; check the default configuration")
			return result
		}
		l.GroupID = group
		l, err = session.ResolveLaunch(l)
		if err != nil {
			result.err = fmt.Errorf("Launch configuration could not be resolved; check program and directory")
			return result
		}
		result.dir = l.Dir
		result.session, result.err = service.Start(ctx, l)
		if result.err == nil && save != nil {
			result.rememberErr = save(ctx, l)
		}
		return result
	}
}

func (m *Model) quickLaunchEvent(msg tea.Msg) (tea.Cmd, bool) {
	result, ok := msg.(quickLaunchResult)
	if !ok {
		return nil, false
	}
	if result.request != m.quickCreate {
		return nil, true
	}
	quitting := m.quickCreate.quitting
	m.quickCreate = nil
	var next tea.Cmd
	if result.err != nil {
		m.quickUncertain = result.session.ID != ""
		m.status = "Creation failed; no automatic retry. Check launch settings."
		if m.quickUncertain {
			m.status = "Launch outcome unknown; verify sessions before reopening the UI. No retry sent."
		}
		next = m.connectWorkspace()
	} else {
		m.listGeneration++
		if m.listCancel != nil {
			m.listCancel()
		}
		m.status = ""
		if result.rememberErr != nil {
			m.status = "Terminal created; last launch settings could not be saved."
		}
		next = tea.Batch(m.acceptLaunchedSession(result.session), m.rememberDirectory(result.dir))
	}
	if quitting {
		m.workspace.focused = false
		return tea.Batch(next, m.workspaceKey(tea.KeyPressMsg{Code: 'q', Text: "q"})), true
	}
	return next, true
}

// Modal creation remains available explicitly with N. Record only a confirmed
// success; persistence failures do not turn a created terminal into a retry.
func rememberSuccessfulLaunch(ctx context.Context, save func(context.Context, session.Launch) error, l session.Launch, s session.Session, err error) error {
	if err != nil || s.ID == "" || save == nil {
		return nil
	}
	return save(ctx, l)
}
