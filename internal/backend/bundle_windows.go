//go:build windows

package backend

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Observe native process/module identities directly: no PowerShell, WMI helper
// or extra resident process. Toolhelp snapshots are closed after this check.
func verifyBundleProcess(pid uint32, created uint64, image, dir string) error {
	server, err := bundleProcessHandle(pid, created, image)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(server)
	if err := bundleModuleLoaded(pid, filepath.Join(dir, "conpty.dll")); err != nil {
		return err
	}
	if err := bundleHostChild(pid, created, filepath.Join(dir, "OpenConsole.exe")); err != nil {
		return err
	}
	return bundleProcessAlive(server)
}

func bundleProcessAlive(handle windows.Handle) error {
	status, err := windows.WaitForSingleObject(handle, 0)
	if err != nil || status != uint32(windows.WAIT_TIMEOUT) {
		return errors.New("original terminal process is no longer live")
	}
	return nil
}

func bundleProcessHandle(pid uint32, created uint64, image string) (windows.Handle, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, pid)
	if err != nil {
		return 0, err
	}
	var start, exit, kernel, user windows.Filetime
	err = windows.GetProcessTimes(h, &start, &exit, &kernel, &user)
	actual := uint64(start.HighDateTime)<<32 | uint64(start.LowDateTime)
	if err == nil && actual != created {
		err = errors.New("terminal process creation time differs")
	}
	path := make([]uint16, 32768)
	size := uint32(len(path))
	if err == nil {
		err = windows.QueryFullProcessImageName(h, 0, &path[0], &size)
	}
	if err == nil && !bundleSameFile(windows.UTF16ToString(path[:size]), image) {
		err = errors.New("terminal process image differs")
	}
	if err == nil {
		err = bundleProcessAlive(h)
	}
	if err != nil {
		windows.CloseHandle(h)
		return 0, err
	}
	return h, nil
}

// File identity handles ordinary and extended DOS paths without weakening
// PID/creation-time checks or doing broad string-prefix normalization.
func bundleSameFile(a, b string) bool {
	x, err := os.Stat(a)
	if err != nil {
		return false
	}
	y, err := os.Stat(b)
	return err == nil && os.SameFile(x, y)
}

func bundleModuleLoaded(pid uint32, path string) error {
	h, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE|windows.TH32CS_SNAPMODULE32, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	entry := windows.ModuleEntry32{Size: uint32(windows.SizeofModuleEntry32)}
	err = windows.Module32First(h, &entry)
	for count := 0; err == nil && count < 4096; count++ {
		if strings.EqualFold(windows.UTF16ToString(entry.Module[:]), "conpty.dll") && bundleSameFile(windows.UTF16ToString(entry.ExePath[:]), path) {
			return nil
		}
		err = windows.Module32Next(h, &entry)
	}
	return errors.New("pinned ConPTY DLL is not loaded; system fallback rejected")
}

func bundleHostChild(parent uint32, parentCreated uint64, path string) error {
	h, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	err = windows.Process32First(h, &entry)
	matches := 0
	for count := 0; err == nil && count < 65536; count++ {
		if entry.ParentProcessID == parent && strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), "OpenConsole.exe") {
			child, openErr := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, entry.ProcessID)
			if openErr != nil {
				return openErr
			}
			var start, exit, kernel, user windows.Filetime
			timeErr := windows.GetProcessTimes(child, &start, &exit, &kernel, &user)
			windows.CloseHandle(child)
			created := uint64(start.HighDateTime)<<32 | uint64(start.LowDateTime)
			if timeErr != nil || created < parentCreated {
				return errors.New("console host creation identity differs")
			}
			verified, verifyErr := bundleProcessHandle(entry.ProcessID, created, path)
			if verifyErr != nil {
				return verifyErr
			}
			windows.CloseHandle(verified)
			matches++
		}
		err = windows.Process32Next(h, &entry)
	}
	if err != windows.ERROR_NO_MORE_FILES {
		return fmt.Errorf("console host enumeration did not complete: %v", err)
	}
	if matches != 1 {
		return errors.New("expected exactly one pinned OpenConsole child")
	}
	return nil
}
