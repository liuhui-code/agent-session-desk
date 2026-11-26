//go:build windows

package pet

import (
	"unsafe"

	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

func (w *nativeWindow) resize(dpi int) error {
	if dpi == 0 {
		value, _, _ := getDPI.Call(w.hwnd)
		dpi = int(value)
	}
	if dpi < 96 {
		dpi = 96
	}
	t := ui.PetDesignAtDPI(dpi)
	images, err := sprites(t.SpriteSize)
	if err != nil {
		return err
	}
	s, err := newSurface(t.Width, t.Height)
	if err != nil {
		return err
	}
	if w.surface != nil {
		w.surface.close()
	}
	w.tokens, w.images, w.surface = t, images, s
	return nil
}

func workArea(monitor uintptr) rect {
	info := monitorDetails{Size: uint32(unsafe.Sizeof(monitorDetails{}))}
	if ok, _, _ := monitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info))); ok != 0 {
		return info.Work
	}
	return rect{}
}

func (w *nativeWindow) place(initial bool) {
	monitor, _, _ := monitorForWindow.Call(w.hwnd, 2)
	area := workArea(monitor)
	if area.Right <= area.Left || area.Bottom <= area.Top {
		return
	}
	current := rect{}
	getWindowRect.Call(w.hwnd, uintptr(unsafe.Pointer(&current)))
	x, y := int(current.Left), int(current.Top)
	if initial {
		x = int(area.Right) - w.tokens.Width - w.tokens.Margin
		y = int(area.Bottom) - w.tokens.Height - w.tokens.Margin
	}
	x, y = clampPosition(x, y, w.tokens.Width, w.tokens.Height, int(area.Left), int(area.Top), int(area.Right), int(area.Bottom))
	setWindowPos.Call(w.hwnd, 0, uintptr(int32(x)), uintptr(int32(y)), uintptr(w.tokens.Width), uintptr(w.tokens.Height), 0x0014) // NOZORDER | NOACTIVATE
}

func (w *nativeWindow) drag() {
	var cursor point
	getCursorPos.Call(uintptr(unsafe.Pointer(&cursor)))
	dx, dy := int(cursor.X-w.startPointer.X), int(cursor.Y-w.startPointer.Y)
	if !w.moved && dx*dx+dy*dy <= w.tokens.DragSlop*w.tokens.DragSlop {
		return
	}
	w.moved = true
	packed := uint64(uint32(cursor.X)) | uint64(uint32(cursor.Y))<<32
	monitor, _, _ := user.NewProc("MonitorFromPoint").Call(uintptr(packed), 2)
	area := workArea(monitor)
	if area.Right <= area.Left || area.Bottom <= area.Top {
		return
	}
	x, y := clampPosition(int(w.startRect.Left)+dx, int(w.startRect.Top)+dy, w.tokens.Width, w.tokens.Height, int(area.Left), int(area.Top), int(area.Right), int(area.Bottom))
	setWindowPos.Call(w.hwnd, 0, uintptr(int32(x)), uintptr(int32(y)), 0, 0, 0x0015) // NOSIZE | NOZORDER | NOACTIVATE
}
