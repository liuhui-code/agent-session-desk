//go:build windows

package updater

import (
	"golang.org/x/sys/windows"
	"os"
	"syscall"
)

func safeInfo(st os.FileInfo, dir bool) bool {
	if st == nil || st.Mode()&os.ModeSymlink != 0 || !((dir && st.IsDir()) || (!dir && st.Mode().IsRegular())) {
		return false
	}
	a, ok := st.Sys().(*syscall.Win32FileAttributeData)
	return ok && a.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT == 0
}
func lockFile(f *os.File) error {
	return windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{})
}
func unlockFile(f *os.File) {
	windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &windows.Overlapped{})
}
