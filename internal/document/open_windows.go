//go:build windows

package document

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	shellExecute    = windows.NewLazySystemDLL("shell32.dll").NewProc("ShellExecuteExW")
	initializeCOM   = windows.NewLazySystemDLL("ole32.dll").NewProc("CoInitializeEx")
	uninitializeCOM = windows.NewLazySystemDLL("ole32.dll").NewProc("CoUninitialize")
)

type shellExecuteInfo struct {
	Size, Mask                        uint32
	Window                            uintptr
	Verb, File, Parameters, Directory *uint16
	Show                              int32
	Instance                          uintptr
	IDList                            uintptr
	Class                             *uint16
	ClassKey                          uintptr
	HotKey                            uint32
	Icon, Process                     uintptr
}

// Open delegates a document to the user's default association. No shell,
// executable argument string, automatic selection or Enter is involved.
func Open(ctx context.Context, path string) error {
	return openWith(ctx, path, func(info *shellExecuteInfo) error {
		ok, _, err := shellExecute.Call(uintptr(unsafe.Pointer(info)))
		if ok == 0 {
			return fmt.Errorf("Windows could not open Markdown; check its default app: %w", err)
		}
		return nil
	})
}
func openWith(ctx context.Context, path string, invoke func(*shellExecuteInfo) error) error {
	resolved, err := Resolve(path, "")
	if err != nil {
		return err
	}
	file, err := windows.UTF16PtrFromString(resolved)
	if err != nil {
		return err
	}
	directory, err := windows.UTF16PtrFromString(filepath.Dir(resolved))
	if err != nil {
		return err
	}
	verb, _ := windows.UTF16PtrFromString("open")
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	result, _, _ := initializeCOM.Call(0, 6) // STA | DISABLE_OLE1DDE
	if int32(result) < 0 {
		return fmt.Errorf("Windows document service unavailable (0x%x)", uint32(result))
	}
	defer uninitializeCOM.Call()
	if err := ctx.Err(); err != nil {
		return err
	}
	info := shellExecuteInfo{Size: uint32(unsafe.Sizeof(shellExecuteInfo{})), Mask: 0x0100 | 0x0400, Verb: verb, File: file, Directory: directory, Show: 1}
	return invoke(&info) // NOASYNC: calling thread has no Windows message loop.
}
