//go:build !windows

package cli

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"io"
)

func prepareWorkbench(context.Context, []string, bool, string, backend.Options, string, io.Reader, io.Writer, io.Writer) (*workbenchHost, bool, error) {
	return &workbenchHost{}, false, nil
}
