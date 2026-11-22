package manager

import (
	"context"
	"fmt"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// RecordOutput captures before taking the metadata/file coordination lock.
// A successful capture alone never establishes that a record was saved.
func (m *Manager) RecordOutput(ctx context.Context, v session.Session) (session.OutputRecord, error) {
	p, err := m.Terminal.Preview(ctx, v, 200)
	if err != nil {
		return session.OutputRecord{}, fmt.Errorf("capture screen: %w", err)
	}
	r, err := m.metadata.RecordOutput(ctx, v, p)
	if err != nil {
		return r, fmt.Errorf("captured screen (session_id=%s run_id=%s); saved output not confirmed: %w", v.ID, v.RunID, err)
	}
	return r, nil
}
func (m *Manager) Output(ctx context.Context, runID string) ([]session.OutputRecord, error) {
	return m.metadata.Output(ctx, runID)
}
