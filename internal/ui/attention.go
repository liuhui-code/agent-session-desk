package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type attentionService interface {
	OpenAttention(context.Context, string, int64) (session.AttentionItem, error)
}
type attentionState struct {
	items   []session.AttentionItem
	last    string
	opening *session.AttentionItem
	busy    bool
}
type attentionOpened struct {
	item session.AttentionItem
	err  error
}

func (m *Model) nextAttention() tea.Cmd {
	if _, ok := m.service.(attentionService); !ok || m.attention.busy || len(m.attention.items) == 0 {
		return nil
	}
	if m.listErr != nil {
		m.status = "Reminder observation unavailable; refresh before opening"
		return nil
	}
	items := m.attention.items
	index := 0
	for i, item := range items {
		if item.ID == m.attention.last {
			index = (i + 1) % len(items)
		}
	}
	for n := range items {
		i := (index + n) % len(items)
		if items[i].ViewedAt == nil {
			index = i
			break
		}
	}
	item := items[index]
	if m.workspace.transition != workspaceIdle || m.groupOrder != nil || m.sessionOrder != nil {
		return nil
	}
	m.attention.opening = &item
	m.attention.busy = true
	return m.navigateWorkspace(workspaceAttention)
}

func (m *Model) openAttention() tea.Cmd {
	item := *m.attention.opening
	m.workspace.transition = workspaceAttention
	ctx, service := m.ctx, m.service.(attentionService)
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		item, err := service.OpenAttention(ctx, item.ID, item.Version)
		return attentionOpened{item, err}
	}
}

func (m *Model) attentionEvent(msg tea.Msg) (tea.Cmd, bool) {
	result, ok := msg.(attentionOpened)
	if !ok {
		return nil, false
	}
	if !m.attention.busy {
		return nil, true
	}
	intent := m.attention.opening
	m.attention.busy, m.attention.opening = false, nil
	m.workspace.transition = workspaceIdle
	valid := false
	for _, current := range m.attention.items {
		if current.ID == intent.ID && current.Version == intent.Version && current.State == "pending" {
			valid = true
		}
	}
	item := result.item
	if result.err != nil || !valid || item.ID != intent.ID || item.Version != intent.Version || item.State != "pending" || !item.TerminalAvailable || !sameRun(item.Source, intent.Source) || item.Source.BackendID != intent.Source.BackendID {
		m.status = "Reminder changed or original terminal unavailable; refresh and check again"
		return tea.Batch(m.connectWorkspace(), m.loadSessions()), true
	}
	for i := range m.attention.items {
		if m.attention.items[i].ID == item.ID {
			m.attention.items[i] = item
		}
	}
	m.attention.last = item.ID
	w := m.workspace
	index := -1
	for i := range w.tabs {
		if sameRun(w.tabs[i].session, item.Source) {
			index = i
			w.tabs[i].session = item.Source
			w.tabs[i].live = true
			break
		}
	}
	if index < 0 {
		w.nextID++
		w.tabs = append(w.tabs, workspaceTab{viewID: w.nextID, session: item.Source, live: true})
		index = len(w.tabs) - 1
	}
	w.focused = true
	return m.activateWorkspaceNow(index), true
}

func (m *Model) sessionNeedsAttention(target session.Session) bool {
	for _, item := range m.attention.items {
		if item.State == "pending" && sameRun(item.Source, target) {
			return true
		}
	}
	return false
}

// The existing top margin hosts the reminder; terminal geometry never changes.
func (m *Model) attentionHeader() (string, int, int) {
	g := m.workspaceGeometry()
	if len(m.attention.items) == 0 || g.bodyY == 0 || g.bodyWidth == 0 {
		return "", 0, 0
	}
	text := fmt.Sprintf(m.tokens.attention.label, len(m.attention.items))
	uncertain := m.listErr != nil
	for _, item := range m.attention.items {
		uncertain = uncertain || !item.TerminalAvailable
	}
	if uncertain {
		text += m.tokens.attention.uncertain
	}
	text = cellTruncate(text, g.bodyWidth, "")
	return text, g.bodyX, g.bodyY - 1
}

func (m *Model) renderAttentionHeader(frame string) string {
	text, x, y := m.attentionHeader()
	if text == "" {
		return frame
	}
	lines := strings.Split(frame, "\n")
	lines[y] = strings.Repeat(" ", x) + m.styles.attention.Render(text) + strings.Repeat(" ", max(0, m.width-x-lipgloss.Width(text)))
	return strings.Join(lines, "\n")
}

func (m *Model) attentionHeaderHit(x, y int) bool {
	text, left, top := m.attentionHeader()
	return text != "" && y == top && x >= left && x < left+lipgloss.Width(text)
}
