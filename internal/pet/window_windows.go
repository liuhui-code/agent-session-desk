//go:build windows

package pet

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"runtime"
	"sync/atomic"
	"unsafe"

	"github.com/liuhui-code/agent-session-desk/internal/ui"
	"golang.org/x/sys/windows"
)

var nativeRunning atomic.Bool
var activeWindow *nativeWindow // accessed only by the locked native UI thread
var petProcedure = windows.NewCallback(petWindowProc)

type nativeWindow struct {
	hwnd                   uintptr
	model                  Model
	tokens                 ui.PetTokens
	images                 [4]*image.RGBA
	surface                *surface
	pressed, moved, hidden bool
	startPointer           point
	startRect              rect
	err                    error
	paints                 int
	tray                   bool
}

func runNative(ctx context.Context, o Options) error {
	if !nativeRunning.CompareAndSwap(false, true) {
		return fmt.Errorf("one pet demo per process")
	}
	defer nativeRunning.Store(false)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := ctx.Err(); err != nil {
		return err
	}
	previous, _, _ := setThreadDPI.Call(^uintptr(3)) // per-monitor-v2 thread context
	if previous == 0 {
		return fmt.Errorf("per-monitor DPI context unavailable")
	}
	defer setThreadDPI.Call(previous)
	model, _ := New(o.State, o.Count)
	w := &nativeWindow{model: model, tokens: ui.PetDesign()}
	activeWindow = w
	defer func() { activeWindow = nil }()
	module, _, _ := getModule.Call(0)
	cursor, _, _ := loadCursor.Call(0, 32512)
	className := utf16("ASD.Pet.Demo")
	class := windowClass{Size: uint32(unsafe.Sizeof(windowClass{})), Proc: petProcedure, Instance: module, Cursor: cursor, Name: className}
	if atom, _, err := registerClass.Call(uintptr(unsafe.Pointer(&class))); atom == 0 {
		return fmt.Errorf("register pet window: %w", err)
	}
	defer unregisterClass.Call(uintptr(unsafe.Pointer(className)), module)
	foregroundBefore, _, _ := getForeground.Call()
	w.hwnd, _, _ = createWindow.Call(petExStyle, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(utf16("Agent Session Desk · Pet Demo"))), petStyle, 0, 0, uintptr(w.tokens.Width), uintptr(w.tokens.Height), 0, 0, module, 0)
	if w.hwnd == 0 {
		return fmt.Errorf("create pet window failed")
	}
	defer destroyWindow.Call(w.hwnd)
	defer func() {
		if w.surface != nil {
			w.surface.close()
		}
	}()
	if err := w.resize(0); err != nil {
		return err
	}
	w.place(true)
	if err := w.paint(); err != nil {
		return err
	}
	showWindow.Call(w.hwnd, 4) // SW_SHOWNOACTIVATE
	foregroundAfter, _, _ := getForeground.Call()
	actualStyle, _, _ := getWindowStyle.Call(w.hwnd, ^uintptr(19)) // GWL_EXSTYLE
	if err := w.addTray(); err != nil {
		return err
	}
	defer w.removeTray()
	w.armAnimation()
	if o.Duration > 0 {
		setTimer.Call(w.hwnd, 2, uintptr(max(1, o.Duration.Milliseconds())), 0)
	}
	closed := make(chan struct{})
	defer close(closed)
	hwnd := w.hwnd
	go func() {
		select {
		case <-ctx.Done():
			postMessage.Call(hwnd, wmClose, 0, 0)
		case <-closed:
		}
	}()
	for {
		var msg message
		result, _, err := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(result) == -1 {
			return fmt.Errorf("pet message loop: %w", err)
		}
		if result == 0 {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
	if o.Report != "" {
		report := struct {
			Demo                              bool    `json:"demo"`
			PID                               int     `json:"pid"`
			HWND                              uintptr `json:"hwnd"`
			Paints                            int     `json:"paints"`
			ForegroundBefore, ForegroundAfter uintptr
			ExStyle                           uint32 `json:"ex_style"`
			State                             Kind   `json:"state"`
		}{true, os.Getpid(), hwnd, w.paints, foregroundBefore, foregroundAfter, uint32(actualStyle), w.model.Kind}
		f, err := os.OpenFile(o.Report, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		err = json.NewEncoder(f).Encode(report)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return w.err
}

func (w *nativeWindow) paint() error {
	if w.hidden {
		return nil
	}
	setWindowText.Call(w.hwnd, uintptr(unsafe.Pointer(utf16("ASD Pet Demo · "+string(w.model.Kind)))))
	if err := w.surface.paint(w.hwnd, frame(w.model, w.tokens, w.images)); err != nil {
		return err
	}
	w.paints++
	return nil
}

func (w *nativeWindow) armAnimation() {
	killTimer.Call(w.hwnd, 1)
	if w.hidden || w.model.Kind == Unknown {
		return
	}
	delay := w.tokens.AlertMS
	if w.model.Kind == Idle {
		delay = w.tokens.OpenMS
		if w.model.blink {
			delay = w.tokens.BlinkMS
		}
	}
	setTimer.Call(w.hwnd, 1, uintptr(delay), 0)
}

func (w *nativeWindow) changed() {
	if err := w.paint(); err != nil {
		w.err = err
		destroyWindow.Call(w.hwnd)
		return
	}
	w.armAnimation()
}

func petWindowProc(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
	w := activeWindow
	if w == nil {
		return defaultPetMessage(hwnd, msg, wp, lp)
	}
	switch msg {
	case wmCommand:
		w.command(wp & 0xffff)
		return 0
	case wmMouseActivate:
		return 3 // MA_NOACTIVATE; ordinary clicks retain keyboard owner
	case wmTimer:
		if wp == 2 {
			destroyWindow.Call(hwnd)
		} else if wp == 1 {
			w.model.Tick()
			w.changed()
		}
		return 0
	case wmLeftDown:
		getCursorPos.Call(uintptr(unsafe.Pointer(&w.startPointer)))
		getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&w.startRect)))
		w.pressed, w.moved = true, false
		setCapture.Call(hwnd)
		return 0
	case wmMouseMove:
		if w.pressed {
			w.drag()
		}
		return 0
	case wmLeftUp:
		pressed, moved := w.pressed, w.moved
		w.pressed = false
		releaseCapture.Call()
		if pressed && !moved {
			w.model.Next()
			w.changed()
		}
		return 0
	case wmCaptureChanged:
		w.pressed = false
		return 0
	case wmRightUp:
		w.menu()
		return 0
	case wmTray:
		if uint32(lp) == wmRightUp {
			w.menu()
		} else if uint32(lp) == wmLeftUp {
			w.show()
		}
		return 0
	case wmDPIChanged:
		if w.hwnd != 0 {
			if err := w.resize(int(wp & 0xffff)); err != nil {
				w.err = err
				destroyWindow.Call(hwnd)
				return 0
			}
			w.place(false)
			w.changed()
		}
		return 0
	case wmDisplayChange:
		if w.hwnd != 0 {
			w.place(false)
		}
		return 0
	case wmClose:
		destroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		killTimer.Call(hwnd, 1)
		killTimer.Call(hwnd, 2)
		w.removeTray()
		postQuit.Call(0)
		return 0
	}
	return defaultPetMessage(hwnd, msg, wp, lp)
}

func defaultPetMessage(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
	r, _, _ := defWindow.Call(hwnd, uintptr(msg), wp, lp)
	return r
}
