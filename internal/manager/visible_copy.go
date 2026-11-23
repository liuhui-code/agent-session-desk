package manager

import (
	"context"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// CaptureVisible delegates only to the terminal. Copying does not create an
// output record or read/write metadata, and cannot fall back to history.
func (m *Manager) CaptureVisible(ctx context.Context, s session.Session) (session.Preview, error) {
	source, ok := m.Terminal.(interface {
		CaptureVisible(context.Context, session.Session) (session.Preview, error)
	})
	if !ok {
		return session.Preview{}, fmt.Errorf("backend has no fresh visible-page capture")
	}
	return source.CaptureVisible(ctx, s)
}
