//go:build windows

package pet

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

func TestWindowsCaptureWindowBindsTokenInsteadOfForegroundAndRejectsChangedCaption(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var owned []uintptr
	defer func() {
		for _, h := range owned {
			destroyWindow.Call(h)
		}
	}()
	token, _ := session.NewID()
	marker, _ := WindowMarker(token)
	makeWindow := func(title string) uintptr {
		h, _, _ := createWindow.Call(0, uintptr(unsafe.Pointer(utf16("STATIC"))), uintptr(unsafe.Pointer(utf16(title))), 0x80000000, 0, 0, 64, 64, 0, 0, 0, 0)
		if h == 0 {
			t.Fatal("owned test window failed")
		}
		showWindow.Call(h, 4)
		owned = append(owned, h)
		return h
	}
	target := makeWindow(marker)
	makeWindow("ASD unrelated test window")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	owner, err := CaptureWindow(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	if owner.hwnd != target || !owner.current() {
		t.Fatal("did not bind the explicit live target")
	}
	before, _, _ := getForeground.Call()
	setWindowText.Call(target, uintptr(unsafe.Pointer(utf16("Another native tab"))))
	if owner.Activate() {
		t.Fatal("caption change was silently treated as original tab")
	}
	after, _, _ := getForeground.Call()
	if after != before {
		t.Fatal("refused activation changed foreground")
	}
	setWindowText.Call(target, uintptr(unsafe.Pointer(utf16(marker))))
	makeWindow(marker)
	if duplicate, e := CaptureWindow(ctx, token); e == nil {
		duplicate.Close()
		t.Fatal("ambiguous token matched one window arbitrarily")
	}
	owner.Close()
	if owner.Activate() {
		t.Fatal("closed owner remained activatable")
	}
}
