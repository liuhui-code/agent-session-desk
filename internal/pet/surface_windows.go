//go:build windows

package pet

import (
	"fmt"
	"image"
	"unsafe"
)

type bitmapInfo struct {
	Size                   uint32
	Width, Height          int32
	Planes, BitCount       uint16
	Compression, ImageSize uint32
	XPixels, YPixels       int32
	Used, Important        uint32
}
type surface struct {
	dc, bitmap, old uintptr
	bits            unsafe.Pointer
	width, height   int
}

func newSurface(width, height int) (*surface, error) {
	s := &surface{width: width, height: height}
	s.dc, _, _ = createDC.Call(0)
	if s.dc == 0 {
		return nil, fmt.Errorf("create native sprite DC failed")
	}
	info := bitmapInfo{Size: uint32(unsafe.Sizeof(bitmapInfo{})), Width: int32(width), Height: -int32(height), Planes: 1, BitCount: 32}
	s.bitmap, _, _ = createDIB.Call(s.dc, uintptr(unsafe.Pointer(&info)), 0, uintptr(unsafe.Pointer(&s.bits)), 0, 0)
	if s.bitmap == 0 || s.bits == nil {
		s.close()
		return nil, fmt.Errorf("create native sprite bitmap failed")
	}
	s.old, _, _ = selectObject.Call(s.dc, s.bitmap)
	if s.old == 0 || s.old == ^uintptr(0) {
		s.close()
		return nil, fmt.Errorf("select native sprite bitmap failed")
	}
	return s, nil
}

func (s *surface) close() {
	if s.dc != 0 && s.old != 0 && s.old != ^uintptr(0) {
		selectObject.Call(s.dc, s.old)
	}
	if s.bitmap != 0 {
		deleteObject.Call(s.bitmap)
		s.bitmap = 0
	}
	if s.dc != 0 {
		deleteDC.Call(s.dc)
		s.dc = 0
	}
}

// image.RGBA already contains premultiplied channels. ULW_ALPHA requires BGRA,
// not straight-alpha PNG bytes, and the DIB is top-down to match Go row order.
func (s *surface) paint(hwnd uintptr, pixels *image.RGBA) error {
	if pixels.Bounds().Dx() != s.width || pixels.Bounds().Dy() != s.height {
		return fmt.Errorf("native surface geometry changed")
	}
	data := unsafe.Slice((*byte)(s.bits), s.width*s.height*4)
	for y := range s.height {
		for x := range s.width {
			i, j := y*pixels.Stride+x*4, (y*s.width+x)*4
			data[j], data[j+1], data[j+2], data[j+3] = pixels.Pix[i+2], pixels.Pix[i+1], pixels.Pix[i], pixels.Pix[i+3]
		}
	}
	r := rect{}
	if ok, _, _ := getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r))); ok == 0 {
		return fmt.Errorf("native sprite bounds unavailable")
	}
	position, origin, dimensions := point{r.Left, r.Top}, point{}, size{int32(s.width), int32(s.height)}
	blend := [4]byte{0, 0, 255, 1}
	ok, _, err := updateLayer.Call(hwnd, 0, uintptr(unsafe.Pointer(&position)), uintptr(unsafe.Pointer(&dimensions)), s.dc, uintptr(unsafe.Pointer(&origin)), 0, uintptr(unsafe.Pointer(&blend)), 2)
	if ok == 0 {
		return fmt.Errorf("update native sprite failed: %w", err)
	}
	return nil
}
