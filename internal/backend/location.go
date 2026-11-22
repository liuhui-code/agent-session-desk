package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Location identifies the backend scope without probing or launching a binary.
// History uses the same path/namespace contract even when the backend is absent.
func Location(opts Options) (kind, dir string, err error) {
	kind = "tmux"
	if runtime.GOOS == "windows" {
		kind = "psmux"
	}
	if opts.Namespace == "" {
		opts.Namespace = "asd-m0"
	}
	if !namespacePattern.MatchString(opts.Namespace) {
		return "", "", fmt.Errorf("namespace must start with asd- and contain 1–48 letters, digits, underscores or hyphens after it")
	}
	if opts.RuntimeDir == "" {
		base, e := os.UserCacheDir()
		if e != nil {
			return "", "", fmt.Errorf("find application cache directory: %w", e)
		}
		opts.RuntimeDir = filepath.Join(base, "agent-session-desk", "m0")
	}
	dir, err = filepath.Abs(opts.RuntimeDir)
	if err != nil {
		return "", "", err
	}
	return kind, filepath.Join(dir, opts.Namespace), nil
}
