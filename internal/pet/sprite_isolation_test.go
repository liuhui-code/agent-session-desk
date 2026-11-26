package pet

import (
	"image"
	"image/color"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

func TestIdleSpritesHaveNoDetachedInkBelowTheCat(t *testing.T) {
	for _, dpi := range []int{96, 144, 192, 384} {
		images, err := sprites(ui.PetDesignAtDPI(dpi).SpriteSize)
		if err != nil {
			t.Fatal(err)
		}
		for cell := 0; cell < 2; cell++ {
			bands := 0
			wasInk := false
			for y := 0; y < images[cell].Bounds().Dy(); y++ {
				ink := false
				for x := 0; x < images[cell].Bounds().Dx(); x++ {
					if images[cell].RGBAAt(x, y).A > 0 {
						ink = true
						break
					}
				}
				if ink && !wasInk {
					bands++
				}
				wasInk = ink
			}
			if bands != 1 {
				t.Errorf("DPI %d cell %d: %d detached horizontal ink bands; neighboring pose leaked below the cat", dpi, cell, bands)
			}
		}
	}
}

func TestSpriteBandCleanupPreservesBodyPixelsAndCanvas(t *testing.T) {
	im := image.NewRGBA(image.Rect(0, 0, 12, 12))
	body := color.RGBA{R: 150, G: 60, B: 20, A: 200}
	for y := 2; y < 8; y++ {
		for x := 3; x < 9; x++ {
			im.SetRGBA(x, y, body)
		}
	}
	im.SetRGBA(0, 0, color.RGBA{A: 1})
	im.SetRGBA(5, 10, color.RGBA{R: 255, A: 255})
	im.SetRGBA(5, 11, color.RGBA{R: 255, A: 255})
	keepPrimaryInkBand(im)
	if im.Bounds() != image.Rect(0, 0, 12, 12) {
		t.Fatal("cleanup changed the cached canvas")
	}
	for y := range 12 {
		for x := range 12 {
			want := color.RGBA{}
			if x >= 3 && x < 9 && y >= 2 && y < 8 {
				want = body
			}
			if im.RGBAAt(x, y) != want {
				t.Fatalf("body moved/changed or fragment survived at %d,%d", x, y)
			}
		}
	}
}
