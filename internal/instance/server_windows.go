//go:build windows

package instance

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"runtime"
	"time"
	"unsafe"
)

var ErrAlreadyRunning = errors.New("workbench already running")

type Server struct {
	hwnd              uintptr
	activationMessage uint32
	activate          func() bool
	cancel            context.CancelFunc
	done              chan struct{}
}

func Listen(ctx context.Context, key string, activate func() bool) (*Server, error) {
	name, err := objectName(key)
	if err != nil {
		return nil, err
	}
	child, cancel := context.WithCancel(ctx)
	s := &Server{activate: activate, cancel: cancel, done: make(chan struct{})}
	ready := make(chan error, 1)
	go s.run(child, name, ready)
	select {
	case err = <-ready:
		if err != nil {
			s.Close()
			return nil, err
		}
		return s, nil
	case <-ctx.Done():
		s.Close()
		return nil, ctx.Err()
	}
}
func (s *Server) run(ctx context.Context, name string, ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(s.done)
	h, err := openMutex(name)
	if err != nil {
		ready <- err
		return
	}
	defer windows.CloseHandle(h)
	acquired, err := waitMutex(ctx, h, false)
	if err != nil {
		ready <- err
		return
	}
	if !acquired {
		ready <- ErrAlreadyRunning
		return
	}
	defer windows.ReleaseMutex(h)
	id, _, _ := registerMessage.Call(uintptr(unsafe.Pointer(wide("AgentSessionDesk.Workbench.Activate.v1"))))
	if id == 0 {
		ready <- fmt.Errorf("register workbench activation message failed")
		return
	}
	s.activationMessage = uint32(id)
	module, _, _ := moduleHandle.Call(0)
	className := wide(name)
	class := windowClass{Size: uint32(unsafe.Sizeof(windowClass{})), Proc: procedure, Instance: module, Name: className}
	if atom, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&class))); atom == 0 {
		ready <- fmt.Errorf("register workbench endpoint failed")
		return
	}
	defer unregisterClass.Call(uintptr(unsafe.Pointer(className)), module)
	hwnd, _, _ := createWindow.Call(0, uintptr(unsafe.Pointer(className)), 0, 0, 0, 0, 0, 0, ^uintptr(2), 0, module, 0)
	if hwnd == 0 {
		ready <- fmt.Errorf("create workbench endpoint failed")
		return
	}
	s.hwnd = hwnd
	registry.Lock()
	registry.windows[hwnd] = s
	registry.Unlock()
	defer func() {
		registry.Lock()
		live := registry.windows[hwnd] == s
		if live {
			delete(registry.windows, hwnd)
		}
		registry.Unlock()
		if live {
			destroyWindow.Call(hwnd)
		}
	}()
	if err = ctx.Err(); err != nil {
		ready <- err
		return
	}
	ready <- nil
	go func() {
		select {
		case <-ctx.Done():
			registry.Lock()
			if registry.windows[hwnd] == s {
				postMessage.Call(hwnd, 0x0010, 0, 0)
			}
			registry.Unlock()
		case <-s.done:
		}
	}()
	for {
		var msg message
		result, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(result) <= 0 {
			return
		}
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
func (s *Server) Close() error {
	if s == nil {
		return nil
	}
	s.cancel()
	select {
	case <-s.done:
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("workbench activation endpoint still closing")
	}
}

// BindWindow publishes only the already validated native target identity. It
// carries no terminal input, command arguments or Session data.
func (s *Server) BindWindow(hwnd uintptr) error {
	registry.Lock()
	defer registry.Unlock()
	if hwnd == 0 || registry.windows[s.hwnd] != s {
		return fmt.Errorf("workbench endpoint or bound window unavailable")
	}
	if ok, _, _ := setProperty.Call(s.hwnd, uintptr(unsafe.Pointer(wide(targetProperty))), hwnd); ok == 0 {
		return fmt.Errorf("publish workbench window identity failed")
	}
	return nil
}
