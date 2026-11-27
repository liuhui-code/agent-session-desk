package cli

import (
	"context"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"io"
)

func startWorkspace(ctx context.Context, m *manager.Manager, path string, args []string, out io.Writer) error {
	if len(args) != 1 || path == "" {
		return fmt.Errorf("workspace NAME requires --config FILE")
	}
	launches, e := config.LoadWorkspace(path, args[0])
	if e != nil {
		return e
	}
	var started []session.Session
	for _, l := range launches {
		v, e := m.Start(ctx, l)
		if e != nil {
			_ = writeJSON(out, started)
			return fmt.Errorf("workspace stopped after %d starts; do not retry whole workspace: %w", len(started), e)
		}
		started = append(started, v)
	}
	return writeJSON(out, started)
}
