package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// ReportAttention accepts an explicitly bound producer, never output heuristics.
func (m *Manager) ReportAttention(ctx context.Context, c session.AgentCredential, event session.AttentionEvent) (session.AttentionItem, error) {
	if _, err := event.Fingerprint(); err != nil {
		return session.AttentionItem{}, err
	}
	if err := m.authenticateAgent(ctx, c); err != nil {
		return session.AttentionItem{}, err
	}
	return m.metadata.ReportAttention(ctx, c, event)
}

// Snooze changes presentation only and performs no terminal discovery or input.
func (m *Manager) SnoozeAttention(ctx context.Context, expected session.AttentionItem, duration time.Duration) (session.AttentionItem, error) {
	return m.metadata.SnoozeAttention(ctx, expected, duration)
}

// OpenAttention revalidates the request and original live Run. Viewed is only a
// receipt for navigation; it is neither a decision nor proof of agent readiness.
func (m *Manager) OpenAttention(ctx context.Context, id string, version int64) (session.AttentionItem, error) {
	item, err := m.metadata.Attention(ctx, id)
	if err != nil {
		return item, err
	}
	if item.State != "pending" || item.Version != version {
		return item, fmt.Errorf("attention changed; refresh before opening")
	}
	target, err := m.messageTarget(ctx, item.Source.ID, item.Source.RunID)
	if err != nil {
		return item, fmt.Errorf("original terminal unavailable; request remains unresolved")
	}
	item, err = m.metadata.MarkAttentionViewed(ctx, id, version, target)
	if err == nil {
		item.Source = target
		item.TerminalAvailable = true
	}
	return item, err
}
