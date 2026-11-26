package pet

import "image"

// The generated sheet leaks small, separated fragments into its upper cells.
// Keep the main continuous band of the two loaf poses, without resizing or
// moving any body pixels. Run once while preparing the sprite cache.
func keepPrimaryInkBand(src *image.RGBA) {
	b := src.Bounds()
	bestStart, bestEnd, start := b.Min.Y, b.Min.Y, -1
	for y := b.Min.Y; y <= b.Max.Y; y++ {
		ink := false
		if y < b.Max.Y {
			for x := b.Min.X; x < b.Max.X; x++ {
				if src.RGBAAt(x, y).A != 0 {
					ink = true
					break
				}
			}
		}
		if ink && start < 0 {
			start = y
		}
		if !ink && start >= 0 {
			if y-start > bestEnd-bestStart {
				bestStart, bestEnd = start, y
			}
			start = -1
		}
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		if y < bestStart || y >= bestEnd {
			i := src.PixOffset(b.Min.X, y)
			clear(src.Pix[i : i+b.Dx()*4])
		}
	}
}
