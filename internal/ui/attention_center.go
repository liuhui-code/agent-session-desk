package ui

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// The center owns navigation only; titles/details stay in the existing snapshot.
type attentionCenter struct {
	selected                           string
	filter                             int
	busy, submitted                    bool
	err                                string
	returnFocused, returnSidebarHidden bool
}
type attentionSnoozeService interface {
	SnoozeAttention(context.Context, session.AttentionItem, time.Duration) (session.AttentionItem, error)
}
type attentionSnoozed struct {
	center *attentionCenter
	item   session.AttentionItem
	err    error
}

func (m *Model) acceptAttentionSnapshot(items []session.AttentionItem) {
	n := min(len(items), session.MaxAttentionItems)
	old := m.attention.items
	if cap(m.attention.items) < n {
		m.attention.items = make([]session.AttentionItem, n)
	} else {
		m.attention.items = m.attention.items[:n]
	}
	copy(m.attention.items, items[:n])
	if n < len(old) {
		clear(old[n:])
	} // release removed strings after overlap-safe copy
}

func (m *Model) beginAttentionCenter() tea.Cmd {
	if _, ok := m.service.(attentionService); !ok || m.attention.busy || m.workspace.transition != workspaceIdle || m.groupOrder != nil || m.sessionOrder != nil {
		return nil
	}
	m.attention.centerIntent = &attentionCenter{returnFocused: m.workspace.focused, returnSidebarHidden: m.workspace.sidebarHidden}
	return m.navigateWorkspace(workspaceAttentionCenter)
}

func (m *Model) attentionCenterRows(now time.Time) []int {
	rows := make([]int, 0, len(m.attention.items))
	for i, item := range m.attention.items {
		if item.State != "pending" {
			continue
		}
		due := item.ReminderDue(now)
		if m.attention.center.filter == 1 && !due || m.attention.center.filter == 2 && due {
			continue
		}
		rows = append(rows, i)
	}
	return rows
}
func (m *Model) attentionCenterSelection(rows []int) int {
	for i, row := range rows {
		if m.attention.items[row].ID == m.attention.center.selected {
			return i
		}
	}
	return 0
}
func (m *Model) closeAttentionCenter() tea.Cmd {
	d := m.attention.center
	m.attention.center = nil
	m.workspace.focused, m.workspace.sidebarHidden = d.returnFocused, d.returnSidebarHidden
	return m.connectWorkspace()
}
func (m *Model) attentionCenterSnooze(item session.AttentionItem, duration time.Duration) tea.Cmd {
	d := m.attention.center
	service, ok := m.service.(attentionSnoozeService)
	if !ok {
		d.err = "Reminder settings unavailable"
		return nil
	}
	d.busy, d.submitted = true, true
	ctx := m.ctx
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		item, err := service.SnoozeAttention(ctx, item, duration)
		return attentionSnoozed{d, item, err}
	}
}
func (m *Model) attentionCenterEvent(msg tea.Msg) (tea.Cmd, bool) {
	if result, ok := msg.(attentionSnoozed); ok {
		d := m.attention.center
		if d == nil || result.center != d {
			return nil, true
		}
		d.busy = false
		if result.err != nil {
			d.err = "Reminder change unconfirmed; close and refresh before retry"
			return m.loadSessions(), true
		}
		d.submitted = false
		d.err = ""
		m.listGeneration++
		if m.listCancel != nil {
			m.listCancel()
		}
		for i, old := range m.attention.items {
			if old.ID == result.item.ID && old.Version == result.item.Version && old.State == "pending" && sameRun(old.Source, result.item.Source) && old.ReminderRevision <= result.item.ReminderRevision {
				m.attention.items[i].SnoozedUntil = result.item.SnoozedUntil
				m.attention.items[i].ReminderRevision = result.item.ReminderRevision
			}
		}
		m.publishAttention(m.listErr == nil)
		return m.loadSessions(), true
	}
	d := m.attention.center
	if d == nil {
		if m.attention.centerIntent != nil {
			switch msg := msg.(type) {
			case tea.WindowSizeMsg:
				m.width, m.height = max(0, msg.Width), max(0, msg.Height)
				return nil, true
			case tea.KeyPressMsg, tea.KeyReleaseMsg, tea.PasteMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
				return nil, true
			}
		}
		return nil, false
	}
	rows := m.attentionCenterRows(time.Now())
	selected := m.attentionCenterSelection(rows)
	move := func(delta int) {
		if len(rows) > 0 {
			selected = max(0, min(len(rows)-1, selected+delta))
			d.selected = m.attention.items[rows[selected]].ID
		}
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		return nil, true
	case tea.KeyPressMsg:
		if d.busy {
			return nil, true
		}
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			return m.closeAttentionCenter(), true
		case "tab":
			d.filter = (d.filter + 1) % 3
			d.selected = ""
			return nil, true
		case "shift+tab":
			d.filter = (d.filter + 2) % 3
			d.selected = ""
			return nil, true
		case "up", "k":
			move(-1)
			return nil, true
		case "down", "j":
			move(1)
			return nil, true
		case "home":
			move(-len(rows))
			return nil, true
		case "end":
			move(len(rows))
			return nil, true
		case "pgup":
			move(-m.tokens.attention.centerRows)
			return nil, true
		case "pgdown":
			move(m.tokens.attention.centerRows)
			return nil, true
		case "r":
			return m.loadSessions(), true
		case "enter", "s", "S", "u":
			if len(rows) == 0 || d.submitted {
				return nil, true
			}
			if m.listErr != nil {
				d.err = "Last observed; refresh before acting"
				return nil, true
			}
			g := m.attentionCenterGeometry(len(rows), selected)
			if m.width < m.tokens.attention.actionWidth || m.height < m.tokens.attention.actionHeight || g.visible == 0 {
				d.err = "Enlarge the window to act"
				return nil, true
			}
			item := m.attention.items[rows[selected]]
			if msg.String() == "enter" {
				if !item.TerminalAvailable {
					d.err = "Original terminal unavailable; request remains pending"
					return nil, true
				}
				m.attention.center = nil
				return m.beginAttention(item), true
			}
			duration := time.Duration(0)
			if msg.String() == "s" {
				duration = 10 * time.Minute
			}
			if msg.String() == "S" {
				duration = 30 * time.Minute
			}
			return m.attentionCenterSnooze(item, duration), true
		}
		return nil, true
	case tea.MouseWheelMsg:
		if !d.busy {
			if msg.Button == tea.MouseWheelUp {
				move(-1)
			} else if msg.Button == tea.MouseWheelDown {
				move(1)
			}
		}
		return nil, true
	case tea.MouseClickMsg:
		if !d.busy && msg.Button == tea.MouseLeft {
			g := m.attentionCenterGeometry(len(rows), selected)
			row := msg.Y - g.rowY + g.start
			if msg.X >= g.x && msg.X < g.x+g.width && msg.Y >= g.rowY && msg.Y < g.rowY+g.visible && row < len(rows) {
				d.selected = m.attention.items[rows[row]].ID
			}
		}
		return nil, true
	case tea.KeyReleaseMsg, tea.PasteMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg:
		return nil, true
	}
	return nil, false
}

func attentionWait(item session.AttentionItem, now time.Time) string {
	if !item.ReminderDue(now) && item.SnoozedUntil != nil {
		return "Snoozed until " + item.SnoozedUntil.Local().Format("15:04")
	}
	minutes := max(0, int(now.Sub(item.CreatedAt)/time.Minute))
	return fmt.Sprintf("Waiting %dm", minutes)
}
