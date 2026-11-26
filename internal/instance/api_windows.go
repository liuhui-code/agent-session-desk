//go:build windows

package instance

import (
	"golang.org/x/sys/windows"
	"sync"
	"unsafe"
)

const targetProperty = "ASD.Workbench.Target.v1"

var user = windows.NewLazySystemDLL("user32.dll")
var (
	registerClass   = user.NewProc("RegisterClassExW")
	unregisterClass = user.NewProc("UnregisterClassW")
	createWindow    = user.NewProc("CreateWindowExW")
	destroyWindow   = user.NewProc("DestroyWindow")
	defWindow       = user.NewProc("DefWindowProcW")
	getMessage      = user.NewProc("GetMessageW")
	dispatchMessage = user.NewProc("DispatchMessageW")
	postMessage     = user.NewProc("PostMessageW")
	postQuit        = user.NewProc("PostQuitMessage")
	findWindow      = user.NewProc("FindWindowExW")
	windowPID       = user.NewProc("GetWindowThreadProcessId")
	registerMessage = user.NewProc("RegisterWindowMessageW")
	sendTimeout     = user.NewProc("SendMessageTimeoutW")
	allowForeground = user.NewProc("AllowSetForegroundWindow")
	moduleHandle    = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetModuleHandleW")
)

type windowClass struct {
	Size, Style                        uint32
	Proc                               uintptr
	ClassExtra, WindowExtra            int32
	Instance, Icon, Cursor, Background uintptr
	Menu, Name                         *uint16
	SmallIcon                          uintptr
}
type point struct{ X, Y int32 }
type message struct {
	Hwnd           uintptr
	ID             uint32
	Wparam, Lparam uintptr
	Time           uint32
	Point          point
	Private        uint32
}

func wide(s string) *uint16 { p, _ := windows.UTF16PtrFromString(s); return p }

var registry = struct {
	sync.Mutex
	windows map[uintptr]*Server
}{windows: map[uintptr]*Server{}}
var procedure = windows.NewCallback(windowProc)

func windowProc(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
	registry.Lock()
	s := registry.windows[hwnd]
	registry.Unlock()
	if s != nil {
		switch {
		case msg == s.activationMessage:
			if s.activate != nil && s.activate() {
				return 1
			}
			return 2
		case msg == 0x0010:
			destroyWindow.Call(hwnd)
			return 0
		case msg == 0x0002:
			registry.Lock()
			delete(registry.windows, hwnd)
			registry.Unlock()
			postQuit.Call(0)
			return 0
		}
	}
	result, _, _ := defWindow.Call(hwnd, uintptr(msg), wp, lp)
	return result
}
func endpoint(name string) uintptr {
	h, _, _ := findWindow.Call(^uintptr(2), 0, uintptr(unsafe.Pointer(wide(name))), 0)
	return h
}

var setProperty = user.NewProc("SetPropW")
var getProperty = user.NewProc("GetPropW")
var getForeground = user.NewProc("GetForegroundWindow")

func boundTarget(hwnd uintptr) uintptr {
	target, _, _ := getProperty.Call(hwnd, uintptr(unsafe.Pointer(wide(targetProperty))))
	return target
}
