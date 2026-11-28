//go:build !windows

package diagnostics

import "os"

func safeDirectory(info os.FileInfo) bool {
	return info != nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0
}
