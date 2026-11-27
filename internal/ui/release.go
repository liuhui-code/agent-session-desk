package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/release"
	"time"
)

type releaseResult struct {
	info release.Info
	err  error
}

func (m *Model) checkRelease() tea.Cmd {
	if m.releaseBusy {
		return nil
	}
	m.releaseBusy = true
	m.status = "Checking public releases…"
	ctx := m.ctx
	return func() tea.Msg {
		c, x := context.WithTimeout(ctx, 5*time.Second)
		defer x()
		v, e := release.Check(c)
		return releaseResult{v, e}
	}
}
func (m *Model) releaseEvent(msg tea.Msg) (tea.Cmd, bool) {
	v, ok := msg.(releaseResult)
	if !ok {
		return nil, false
	}
	m.releaseBusy = false
	if v.err != nil {
		m.status = "Release check unavailable; current " + release.Build()
	} else {
		m.status = fmt.Sprintf("Current %s · Published %s · %s", release.Build(), v.info.Tag, v.info.URL)
	}
	return nil, true
}
