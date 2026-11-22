package manager

import (
	"context"
	"errors"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type interactiveTerminal interface {
	OpenTerminal(context.Context, session.Session, int, int) (terminal.Connection, error)
}

func (m *Manager) SupportsTerminal() bool {
	if capability, ok := m.Terminal.(interface{ SupportsTerminal() bool }); ok {
		return capability.SupportsTerminal()
	}
	_, ok := m.Terminal.(interactiveTerminal)
	return ok
}

// OpenTerminal opens a view of the exact existing Run. The backend revalidates
// the live target; metadata history is never a source of control authority.
func (m *Manager) OpenTerminal(ctx context.Context, s session.Session, width, height int) (terminal.Connection, error) {
	backend, ok := m.Terminal.(interactiveTerminal)
	if !ok {
		return nil, errors.New("this backend does not provide a Tab terminal transport")
	}
	return backend.OpenTerminal(ctx, s, width, height)
}
