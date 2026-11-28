//go:build windows

package cli

import (
	"context"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/instance"
	"path/filepath"
)

func offlinePortableUpdate(ctx context.Context, root string, fn func() error) error {
	opts, dir, err := portableOptions(filepath.Join(root, "asd.exe"), backend.Options{}, "")
	if err != nil {
		return err
	}
	if opts.TerminalBundle == "" || dir != filepath.Join(root, "data", "state") {
		return fmt.Errorf("managed portable bundle required")
	}
	metadata, _, err := metadataPath(opts, dir)
	if err != nil {
		return err
	}
	kind, location, err := backend.Location(opts)
	if err != nil {
		return err
	}
	key, err := instance.ScopeKey(kind, location, metadata)
	if err != nil {
		return err
	}
	return instance.WithStopped(ctx, key, fn)
}
