//go:build !windows

package cli

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"io"
)

func openTerminalHost(context.Context, []string, bool, string, io.Reader, io.Writer) (bool, error) {
	return false, nil
}

func applyPortableOptions(opts backend.Options, stateDir string) (backend.Options, string, error) {
	return opts, stateDir, nil
}
