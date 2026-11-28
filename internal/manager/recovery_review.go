package manager

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/recovery"
)

func (m *Manager) RecoveryReview(ctx context.Context, id string) (recovery.Review, error) {
	return m.metadata.RecoveryReview(ctx, id)
}
