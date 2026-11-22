package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func historyLimit(args []string, output io.Writer) (int, error) {
	f := flag.NewFlagSet("history", flag.ContinueOnError)
	f.SetOutput(output)
	limit := 100
	f.IntVar(&limit, "limit", 100, "most recent records (1–1000)")
	if err := f.Parse(args); err != nil {
		return 0, err
	}
	if f.NArg() != 0 || limit < 1 || limit > 1000 {
		return 0, fmt.Errorf("usage: asd history [--limit 1–1000]")
	}
	return limit, nil
}

func openMetadata(ctx context.Context, opts backend.Options, dir string) (*store.Store, error) {
	path, scope, err := metadataPath(opts, dir)
	if err != nil {
		return nil, err
	}
	return store.Open(ctx, path, scope)
}

func metadataPath(opts backend.Options, dir string) (string, string, error) {
	kind, location, err := backend.Location(opts)
	if err != nil {
		return "", "", err
	}
	scope := kind + ":" + location
	if runtime.GOOS == "windows" {
		scope = strings.ToLower(scope)
	}
	if dir == "" {
		if opts.RuntimeDir != "" {
			dir = location
		} else {
			base, e := os.UserConfigDir()
			if e != nil {
				return "", "", e
			}
			dir = filepath.Join(base, "agent-session-desk")
		}
	}
	return filepath.Join(dir, "metadata.sqlite3"), scope, nil
}
