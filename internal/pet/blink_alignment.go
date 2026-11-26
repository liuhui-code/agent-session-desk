package pet

import (
	"image"
	"image/draw"
)

// The source sheet's closed-eye pose is horizontally offset. Align its body
// with the open-eye pose once, while filling the bounded sprite cache; preserve
// blink timing, scale and vertical position. Do not shift the native window.
func alignBlink(images []*image.RGBA, alpha uint8) {
	center := func(src *image.RGBA) int {
		left, right := src.Bounds().Max.X, src.Bounds().Min.X-1
		for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
			for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
				if src.RGBAAt(x, y).A >= alpha {
					left, right = min(left, x), max(right, x)
				}
			}
		}
		return left + right
	}
	dx := (center(images[0]) - center(images[1])) / 2
	if dx == 0 {
		return
	}
	aligned := image.NewRGBA(images[1].Bounds())
	draw.Draw(aligned, aligned.Bounds().Add(image.Pt(dx, 0)), images[1], image.Point{}, draw.Src)
	images[1] = aligned
}
