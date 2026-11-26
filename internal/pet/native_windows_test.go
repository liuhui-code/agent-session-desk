//go:build windows

package pet

import (
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestWindowsNativePetABISizes(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("prototype targets Windows x64")
	}
	for name, pair := range map[string][2]uintptr{
		"window class": {unsafe.Sizeof(windowClass{}), 80}, "message": {unsafe.Sizeof(message{}), 48},
		"bitmap": {unsafe.Sizeof(bitmapInfo{}), 40}, "tray": {unsafe.Sizeof(notifyData{}), 976},
	} {
		if pair[0] != pair[1] {
			t.Fatalf("%s: %d != %d", name, pair[0], pair[1])
		}
	}
}

func TestWindowsNativePetGDISurfacesReleaseOriginalObjects(t *testing.T) {
	guiResources := user.NewProc("GetGuiResources")
	process := windows.CurrentProcess()
	before, _, _ := guiResources.Call(uintptr(process), 0)
	for range 20 {
		s, err := newSurface(16, 16)
		if err != nil {
			t.Fatal(err)
		}
		data := unsafe.Slice((*byte)(s.bits), 16*16*4)
		data[0], data[1], data[2], data[3] = 0x10, 0x20, 0x30, 0x40
		if data[3] != 0x40 {
			t.Fatal("native DIB memory unavailable")
		}
		s.close()
		s.close()
	}
	after, _, _ := guiResources.Call(uintptr(process), 0)
	if after != before {
		t.Fatalf("GDI objects leaked: %d -> %d", before, after)
	}
}
