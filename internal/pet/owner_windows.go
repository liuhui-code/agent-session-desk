//go:build windows

package pet

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"golang.org/x/sys/windows"
)

var (
	windowPID     = user.NewProc("GetWindowThreadProcessId")
	setProp       = user.NewProc("SetPropW")
	getProp       = user.NewProc("GetPropW")
	removeProp    = user.NewProc("RemovePropW")
	setForeground = user.NewProc("SetForegroundWindow")
	isIconic      = user.NewProc("IsIconic")
)

// Owner binds the workbench's foreground ancestor window once. A retained
// process handle and per-window nonce reject reused PID/HWND identities.
// Failure never falls back to title search or manufactured keyboard input.
type Owner struct {
	hwnd     uintptr
	pid      uint32
	process  windows.Handle
	property *uint16
}

func CaptureOwner() (*Owner, error) {
	hwnd, _, _ := getForeground.Call()
	var pid uint32
	windowPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)
	parents := map[uint32]uint32{}
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	for e := windows.Process32First(snapshot, &entry); e == nil; e = windows.Process32Next(snapshot, &entry) {
		parents[entry.ProcessID] = entry.ParentProcessID
	}
	if hwnd == 0 || pid == 0 {
		return nil, fmt.Errorf("foreground window is not the workbench host")
	}
	handle, err := ancestorHandle(pid, parents)
	if err != nil {
		return nil, err
	}
	nonce, err := session.NewID()
	if err != nil {
		windows.CloseHandle(handle)
		return nil, err
	}
	owner := &Owner{hwnd: hwnd, pid: pid, process: handle, property: utf16("ASD.Attention.Owner." + nonce)}
	if ok, _, err := setProp.Call(hwnd, uintptr(unsafe.Pointer(owner.property)), 1); ok == 0 {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("bind workbench window: %w", err)
	}
	if !owner.current() {
		owner.Close()
		return nil, fmt.Errorf("workbench window changed during binding")
	}
	return owner, nil
}

func ancestorHandle(pid uint32, parents map[uint32]uint32) (windows.Handle, error) {
	// A recycled parent PID starts after its original child. Check creation
	// ordering at every hop, keeping the target handle once reached.
	limit := ^uint64(0)
	for current, depth := uint32(os.Getpid()), 0; current != 0 && depth < 64; current, depth = parents[current], depth+1 {
		handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, current)
		if err != nil {
			return 0, err
		}
		var born, ended, kernelTime, userTime windows.Filetime
		err = windows.GetProcessTimes(handle, &born, &ended, &kernelTime, &userTime)
		created := uint64(born.HighDateTime)<<32 | uint64(born.LowDateTime)
		state, waitErr := windows.WaitForSingleObject(handle, 0)
		if err != nil || waitErr != nil || state != uint32(windows.WAIT_TIMEOUT) || created > limit {
			windows.CloseHandle(handle)
			return 0, fmt.Errorf("workbench ancestor identity unavailable")
		}
		if current == pid {
			return handle, nil
		}
		limit = created
		windows.CloseHandle(handle)
	}
	return 0, fmt.Errorf("foreground window is not the workbench host")
}

func (o *Owner) current() bool {
	if o == nil || o.process == 0 {
		return false
	}
	state, err := windows.WaitForSingleObject(o.process, 0)
	if err != nil || state != uint32(windows.WAIT_TIMEOUT) {
		return false
	}
	var pid uint32
	windowPID.Call(o.hwnd, uintptr(unsafe.Pointer(&pid)))
	nonce, _, _ := getProp.Call(o.hwnd, uintptr(unsafe.Pointer(o.property)))
	return pid == o.pid && nonce == 1
}

func (o *Owner) Activate() bool {
	if !o.current() {
		return false
	}
	if iconic, _, _ := isIconic.Call(o.hwnd); iconic != 0 {
		showWindow.Call(o.hwnd, 9)
	}
	setForeground.Call(o.hwnd) // May be refused by Windows foreground policy.
	foreground, _, _ := getForeground.Call()
	return foreground == o.hwnd && o.current()
}

func (o *Owner) Close() {
	if o == nil || o.process == 0 {
		return
	}
	if o.current() {
		removeProp.Call(o.hwnd, uintptr(unsafe.Pointer(o.property)))
	}
	windows.CloseHandle(o.process)
	o.process = 0
}
