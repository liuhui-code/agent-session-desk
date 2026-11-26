//go:build windows

package pet

import (
	"context"
	"fmt"
	"golang.org/x/sys/windows"
	"sync"
	"time"
	"unsafe"
)

var enumWindows = user.NewProc("EnumWindows")
var isWindowVisible = user.NewProc("IsWindowVisible")
var getWindowText = user.NewProc("GetWindowTextW")
var setConsoleTitle = kernel.NewProc("SetConsoleTitleW")

func ownerCaption(hwnd uintptr) string {
	var buffer [128]uint16
	n, _, _ := getWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	return windows.UTF16ToString(buffer[:min(int(n), len(buffer))])
}

type ownerEnumeration struct {
	parents    map[uint32]uint32
	marker     string
	candidates []uintptr
}

var ownerEnumMu sync.Mutex
var currentOwnerEnum *ownerEnumeration
var ownerEnumCallback = windows.NewCallback(func(hwnd, extra uintptr) uintptr {
	scan := currentOwnerEnum
	if scan == nil {
		return 0
	}
	if visible, _, _ := isWindowVisible.Call(hwnd); visible == 0 {
		return 1
	}
	if scan.marker != "" && ownerCaption(hwnd) != scan.marker {
		return 1
	}
	var pid uint32
	windowPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	handle, err := ancestorHandle(pid, scan.parents)
	if err != nil {
		return 1
	}
	windows.CloseHandle(handle)
	scan.candidates = append(scan.candidates, hwnd)
	if len(scan.candidates) > 1 {
		return 0
	}
	return 1
})

func ownerWindows(parents map[uint32]uint32, marker string) []uintptr {
	ownerEnumMu.Lock()
	defer ownerEnumMu.Unlock()
	scan := &ownerEnumeration{parents: parents, marker: marker}
	currentOwnerEnum = scan
	defer func() { currentOwnerEnum = nil }()
	enumWindows.Call(ownerEnumCallback, 0)
	return scan.candidates
}

// CaptureWindow never falls back to the current foreground window. The token
// is authored by our launcher and must identify exactly one live ancestor HWND.
func CaptureWindow(ctx context.Context, token string) (*Owner, error) {
	marker, err := WindowMarker(token)
	if err != nil {
		return nil, err
	}
	parents, err := ownerParents()
	if err != nil {
		return nil, err
	}
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		candidates := ownerWindows(parents, marker)
		if len(candidates) > 1 {
			return nil, fmt.Errorf("workbench launch window is ambiguous")
		}
		if len(candidates) == 1 {
			owner, err := bindOwner(candidates[0], parents)
			if err != nil {
				return nil, err
			}
			owner.caption = marker
			return owner, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("workbench launch window unavailable: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

// Present replaces the transient token title only after binding the exact HWND.
// The managed launcher permits this application title update explicitly.
func (o *Owner) Present(ctx context.Context) error {
	if o == nil {
		return fmt.Errorf("workbench window unavailable")
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.current() {
		return fmt.Errorf("workbench window binding changed")
	}
	if ok, _, _ := setConsoleTitle.Call(uintptr(unsafe.Pointer(utf16(WorkbenchCaption)))); ok == 0 {
		return fmt.Errorf("could not name the workbench terminal")
	}
	for {
		if !o.current() {
			return fmt.Errorf("workbench window binding changed during title acknowledgement")
		}
		if ownerCaption(o.hwnd) == WorkbenchCaption {
			o.caption = WorkbenchCaption
			return nil
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("workbench title acknowledgement unavailable: %w", ctx.Err())
		case <-timer.C:
		}
	}
}
