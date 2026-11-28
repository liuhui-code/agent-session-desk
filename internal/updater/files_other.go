//go:build !windows

package updater

import (
	"os"
	"syscall"
)

func safeInfo(st os.FileInfo, dir bool) bool {
	return st != nil && st.Mode()&os.ModeSymlink == 0 && ((dir && st.IsDir()) || (!dir && st.Mode().IsRegular()))
}
func lockFile(f *os.File) error { return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) }
func unlockFile(f *os.File)     { syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }
