//go:build windows

package pet

import "golang.org/x/sys/windows"

var (
	user             = windows.NewLazySystemDLL("user32.dll")
	gdi              = windows.NewLazySystemDLL("gdi32.dll")
	shell            = windows.NewLazySystemDLL("shell32.dll")
	kernel           = windows.NewLazySystemDLL("kernel32.dll")
	createWindow     = user.NewProc("CreateWindowExW")
	registerClass    = user.NewProc("RegisterClassExW")
	unregisterClass  = user.NewProc("UnregisterClassW")
	defWindow        = user.NewProc("DefWindowProcW")
	getMessage       = user.NewProc("GetMessageW")
	translateMessage = user.NewProc("TranslateMessage")
	dispatchMessage  = user.NewProc("DispatchMessageW")
	destroyWindow    = user.NewProc("DestroyWindow")
	postQuit         = user.NewProc("PostQuitMessage")
	postMessage      = user.NewProc("PostMessageW")
	showWindow       = user.NewProc("ShowWindow")
	setWindowText    = user.NewProc("SetWindowTextW")
	setWindowPos     = user.NewProc("SetWindowPos")
	getWindowRect    = user.NewProc("GetWindowRect")
	getWindowStyle   = user.NewProc("GetWindowLongPtrW")
	getCursorPos     = user.NewProc("GetCursorPos")
	setCapture       = user.NewProc("SetCapture")
	releaseCapture   = user.NewProc("ReleaseCapture")
	setTimer         = user.NewProc("SetTimer")
	killTimer        = user.NewProc("KillTimer")
	getDC            = user.NewProc("GetDC")
	releaseDC        = user.NewProc("ReleaseDC")
	updateLayer      = user.NewProc("UpdateLayeredWindow")
	getDPI           = user.NewProc("GetDpiForWindow")
	setThreadDPI     = user.NewProc("SetThreadDpiAwarenessContext")
	monitorForWindow = user.NewProc("MonitorFromWindow")
	monitorInfo      = user.NewProc("GetMonitorInfoW")
	getForeground    = user.NewProc("GetForegroundWindow")
	loadCursor       = user.NewProc("LoadCursorW")
	loadIcon         = user.NewProc("LoadIconW")
	getModule        = kernel.NewProc("GetModuleHandleW")
	createDC         = gdi.NewProc("CreateCompatibleDC")
	deleteDC         = gdi.NewProc("DeleteDC")
	createDIB        = gdi.NewProc("CreateDIBSection")
	selectObject     = gdi.NewProc("SelectObject")
	deleteObject     = gdi.NewProc("DeleteObject")
	createMenu       = user.NewProc("CreatePopupMenu")
	appendMenu       = user.NewProc("AppendMenuW")
	trackMenu        = user.NewProc("TrackPopupMenu")
	destroyMenu      = user.NewProc("DestroyMenu")
	notifyIcon       = shell.NewProc("Shell_NotifyIconW")
)

const (
	wmDestroy        = 0x0002
	wmClose          = 0x0010
	wmMouseActivate  = 0x0021
	wmTimer          = 0x0113
	wmMouseMove      = 0x0200
	wmLeftDown       = 0x0201
	wmLeftUp         = 0x0202
	wmRightUp        = 0x0205
	wmCaptureChanged = 0x0215
	wmDisplayChange  = 0x007e
	wmDPIChanged     = 0x02e0
	wmTray           = 0x8001
	wmLiveUpdate     = 0x8002
	wmCommand        = 0x0111
	petStyle         = 0x80000000 // WS_POPUP
	petExStyle       = 0x08080088 // NOACTIVATE | LAYERED | TOOLWINDOW | TOPMOST
)

type point struct{ X, Y int32 }
type rect struct{ Left, Top, Right, Bottom int32 }
type size struct{ Width, Height int32 }
type message struct {
	Hwnd           uintptr
	ID             uint32
	Wparam, Lparam uintptr
	Time           uint32
	Point          point
	Private        uint32
}
type windowClass struct {
	Size, Style                        uint32
	Proc                               uintptr
	ClassExtra, WindowExtra            int32
	Instance, Icon, Cursor, Background uintptr
	Menu, Name                         *uint16
	SmallIcon                          uintptr
}
type monitorDetails struct {
	Size          uint32
	Monitor, Work rect
	Flags         uint32
}

func utf16(s string) *uint16 { p, _ := windows.UTF16PtrFromString(s); return p }
