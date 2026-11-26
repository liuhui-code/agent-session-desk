package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// AttentionActivate is only a navigation intent. Authority comes from the
// current scoped snapshot, Manager.OpenAttention and the original Run checks.
type AttentionActivate struct {
	ID      string
	Version int64
}
type AttentionCompanionFailure struct{ Err error }

func (m *Model) publishAttention(known bool) {
	if m.attentionObserver != nil {
		m.attentionObserver(append([]session.AttentionItem(nil), m.attention.items...), known)
	}
}

func (m *Model) companionEvent(msg tea.Msg) (tea.Cmd, bool) {
	switch v := msg.(type) {
	case AttentionCompanionFailure:
		m.status = "Desktop reminder unavailable; use Alt+N in the workbench"
		return nil, true
	case AttentionActivate:
		if m.workspace == nil || m.attaching || m.history.open || m.help || m.launch != nil || m.search != nil || m.rename != nil || m.groupDialog != nil || m.groupRemoval != nil || m.sessionStop != nil || m.densitySettings != nil {
			m.status = "Finish the current navigation before opening the pending item"
			return nil, true
		}
		for _, item := range m.attention.items {
			if item.ID == v.ID && item.Version == v.Version && item.State == "pending" {
				return m.beginAttention(item), true
			}
		}
		m.status = "Pending item changed; refresh before opening"
		return nil, true
	}
	return nil, false
}

func (m *Model) beginAttention(item session.AttentionItem) tea.Cmd {
	if _, ok := m.service.(attentionService); !ok || m.attention.busy || m.listErr != nil || m.workspace.transition != workspaceIdle || m.groupOrder != nil || m.sessionOrder != nil {
		return nil
	}
	m.attention.opening = &item
	m.attention.busy = true
	return m.navigateWorkspace(workspaceAttention)
}
