//go:build windows

package pet

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

type notifyData struct {
	Size               uint32
	Window             uintptr
	ID, Flags, Message uint32
	Icon               uintptr
	Tip                [128]uint16
	State, StateMask   uint32
	Info               [256]uint16
	Timeout            uint32
	InfoTitle          [64]uint16
	InfoFlags          uint32
	GUID               windows.GUID
	BalloonIcon        uintptr
}

func (w *nativeWindow) trayData() notifyData {
	icon, _, _ := loadIcon.Call(0, 32512)
	d := notifyData{Size: uint32(unsafe.Sizeof(notifyData{})), Window: w.hwnd, ID: 1, Flags: 7, Message: wmTray, Icon: icon}
	copy(d.Tip[:], windows.StringToUTF16("ASD Pet Demo · click to restore"))
	return d
}

func (w *nativeWindow) addTray() error {
	d := w.trayData()
	if ok, _, _ := notifyIcon.Call(0, uintptr(unsafe.Pointer(&d))); ok == 0 {
		return fmt.Errorf("desktop tray unavailable; demo closed")
	}
	w.tray = true
	return nil
}

func (w *nativeWindow) removeTray() {
	if w.tray {
		d := w.trayData()
		notifyIcon.Call(2, uintptr(unsafe.Pointer(&d)))
		w.tray = false
	}
}

func (w *nativeWindow) show() {
	w.hidden = false
	showWindow.Call(w.hwnd, 4)
	w.changed()
}

func (w *nativeWindow) menu() {
	menu, _, _ := createMenu.Call()
	if menu == 0 {
		return
	}
	defer destroyMenu.Call(menu)
	labels := []string{"Idle (demo)", "Approval (demo)", "Question (demo)", "Unknown (demo)", "Show / hide", "Exit demo"}
	for i, label := range labels {
		flags := uintptr(0)
		if i < 4 && w.model.Kind == []Kind{Idle, Approval, Question, Unknown}[i] {
			flags = 8
		} // MF_CHECKED
		appendMenu.Call(menu, flags, uintptr(i+1), uintptr(unsafe.Pointer(utf16(label))))
	}
	var cursor point
	getCursorPos.Call(uintptr(unsafe.Pointer(&cursor)))
	// User explicitly requested a menu. Do not activate the pet or manufacture
	// keyboard input to dismiss it; native menu handling owns this interaction.
	command, _, _ := trackMenu.Call(menu, 0x0102, uintptr(cursor.X), uintptr(cursor.Y), 0, w.hwnd, 0)
	w.command(command)
}

func (w *nativeWindow) command(command uintptr) {
	switch command {
	case 1, 2, 3, 4:
		w.model.Kind = []Kind{Idle, Approval, Question, Unknown}[command-1]
		w.model.blink = false
		w.show()
	case 5:
		if w.hidden {
			w.show()
		} else {
			w.hidden = true
			showWindow.Call(w.hwnd, 0)
			w.armAnimation()
		}
	case 6:
		destroyWindow.Call(w.hwnd)
	}
}
