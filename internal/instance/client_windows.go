//go:build windows

package instance

import (
	"context"
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"time"
	"unsafe"
)

// Notify never creates a UI. found remains true if an existing owner is busy,
// unreachable or refuses foreground activation, preventing fallback duplicates.
func Notify(ctx context.Context, key string) (found, activated bool, err error) {
	name, err := objectName(key)
	if err != nil {
		return false, false, err
	}
	running, err := owned(ctx, name)
	if err != nil || !running {
		return running, false, err
	}
	hwnd, err := waitEndpoint(ctx, name)
	if err != nil {
		return true, false, err
	}
	var pid uint32
	windowPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, pid)
	if err != nil {
		return true, false, err
	}
	defer windows.CloseHandle(process)
	ownSID, err := currentSID(windows.CurrentProcess())
	if err != nil {
		return true, false, err
	}
	peerSID, err := currentSID(process)
	if err != nil {
		return true, false, err
	}
	var ownSession, peerSession uint32
	if windows.ProcessIdToSessionId(uint32(os.Getpid()), &ownSession) != nil || windows.ProcessIdToSessionId(pid, &peerSession) != nil || ownSID != peerSID || ownSession != peerSession {
		return true, false, fmt.Errorf("workbench endpoint identity differs")
	}
	if state, e := windows.WaitForSingleObject(process, 0); e != nil || state != uint32(windows.WAIT_TIMEOUT) {
		return true, false, fmt.Errorf("workbench endpoint exited")
	}
	allowForeground.Call(uintptr(pid)) // Best effort; the receiver checks actual foreground.
	msg, _, _ := registerMessage.Call(uintptr(unsafe.Pointer(wide("AgentSessionDesk.Workbench.Activate.v1"))))
	wait := 2 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		wait = min(wait, time.Until(deadline))
	}
	if wait <= 0 {
		if err := ctx.Err(); err != nil {
			return true, false, err
		}
		return true, false, context.DeadlineExceeded
	}
	var reply uintptr
	ok, _, _ := sendTimeout.Call(hwnd, msg, 0, 0, 0x0003, uintptr(max(1, wait.Milliseconds())), uintptr(unsafe.Pointer(&reply)))
	if ok == 0 {
		return true, false, fmt.Errorf("existing workbench activation timed out or was refused")
	}
	if state, e := windows.WaitForSingleObject(process, 0); e != nil || state != uint32(windows.WAIT_TIMEOUT) {
		return true, false, fmt.Errorf("workbench endpoint exited during activation")
	}
	var current uint32
	windowPID.Call(hwnd, uintptr(unsafe.Pointer(&current)))
	if current != pid {
		return true, false, fmt.Errorf("workbench endpoint identity changed")
	}
	switch reply {
	case 1:
		target := boundTarget(hwnd)
		if target != 0 {
			foreground, _, _ := getForeground.Call()
			if foreground != target {
				return true, false, nil
			}
		}
		return true, true, nil
	case 2:
		return true, false, nil
	default:
		return true, false, fmt.Errorf("unrecognized workbench activation reply")
	}

}
func waitEndpoint(ctx context.Context, name string) (uintptr, error) {
	for {
		if h := endpoint(name); h != 0 {
			return h, nil
		}
		if err := pause(ctx); err != nil {
			return 0, err
		}
	}
}

// WaitReady acknowledges a bound endpoint, not backend readiness or a rendered UI.
func WaitReady(ctx context.Context, key string) error {
	name, err := objectName(key)
	if err != nil {
		return err
	}
	for {
		if hwnd := endpoint(name); hwnd != 0 && boundTarget(hwnd) != 0 {
			return nil
		}
		if err = pause(ctx); err != nil {
			return err
		}
	}
}
