//go:build windows

package hotkey

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

type message struct {
	Window  uintptr
	ID      uint32
	W, L    uintptr
	Time    uint32
	X, Y    int32
	Private uint32
}

func Run(ctx context.Context, activate func()) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	u := syscall.NewLazyDLL("user32.dll")
	reg := u.NewProc("RegisterHotKey")
	un := u.NewProc("UnregisterHotKey")
	get := u.NewProc("GetMessageW")
	post := u.NewProc("PostThreadMessageW")
	peek := u.NewProc("PeekMessageW")
	var msg message
	peek.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0, 0)
	tid, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("GetCurrentThreadId").Call()
	ok, _, _ := reg.Call(0, 1, 0x4003, 0x20)
	if ok == 0 {
		return fmt.Errorf("Ctrl+Alt+Space unavailable")
	}
	defer un.Call(0, 1)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			post.Call(tid, 0x12, 0, 0)
		case <-done:
		}
	}()
	for {
		n, _, _ := get.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if n == 0 {
			return nil
		}
		if int32(n) == -1 {
			return fmt.Errorf("hotkey queue failed")
		}
		if msg.ID == 0x312 && msg.W == 1 {
			activate()
		}
	}
}
