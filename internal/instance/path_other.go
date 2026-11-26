//go:build !windows

package instance

import "path/filepath"

func finalPath(path string) (string, error) { return filepath.EvalSymlinks(path) }
func normalizePath(path string) string      { return path }
