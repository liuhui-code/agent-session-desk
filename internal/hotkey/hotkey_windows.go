//go:build windows

package hotkey

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
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

func Run(ctx context.Context, key string, activate func()) error {
	if key == "" {
		return nil
	}
	mods, vk, e := nativeKey(key)
	if e != nil {
		return e
	}
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
	ok, _, _ := reg.Call(0, 1, mods|0x4000, vk)
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

func nativeKey(key string) (uintptr, uintptr, error) {
	var mods uintptr
	parts := strings.Split(key, "+")
	for _, p := range parts[:len(parts)-1] {
		switch p {
		case "ctrl":
			mods |= 2
		case "alt":
			mods |= 1
		case "shift":
			mods |= 4
		default:
			return 0, 0, fmt.Errorf("invalid modifier")
		}
	}
	k := parts[len(parts)-1]
	if k == "space" {
		return mods, 32, nil
	}
	if len(k) == 1 && ((k[0] >= 'a' && k[0] <= 'z') || (k[0] >= '0' && k[0] <= '9')) {
		return mods, uintptr(strings.ToUpper(k)[0]), nil
	}
	if strings.HasPrefix(k, "f") {
		if n, e := strconv.Atoi(k[1:]); e == nil && n >= 1 && n <= 11 {
			return mods, uintptr(111 + n), nil
		}
	}
	return 0, 0, fmt.Errorf("unsupported global key")
}
