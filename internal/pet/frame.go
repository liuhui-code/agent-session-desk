package pet

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"

	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

//go:embed assets/orange-cat.png
var spriteSheet []byte

// Only four small premultiplied frames are retained. The decoded source sheet
// is not an animation cache; no new decode occurs on native timer callbacks.
func sprites(size int) ([4]*image.RGBA, error) {
	var out [4]*image.RGBA
	src, err := png.Decode(bytes.NewReader(spriteSheet))
	if err != nil {
		return out, err
	}
	b := src.Bounds()
	if b.Dx() < 2 || b.Dy() < 2 {
		return out, fmt.Errorf("expected a 2x2 sprite sheet")
	}
	for i := range out {
		x0, x1 := b.Dx()*(i%2)/2, b.Dx()*(i%2+1)/2
		y0, y1 := b.Dy()*(i/2)/2, b.Dy()*(i/2+1)/2
		dst := image.NewRGBA(image.Rect(0, 0, size, size))
		for y := range size {
			for x := range size {
				left, right := x0+x*(x1-x0)/size, x0+(x+1)*(x1-x0)/size
				top, bottom := y0+y*(y1-y0)/size, y0+(y+1)*(y1-y0)/size
				var sum [4]uint64
				for sy := top; sy < max(top+1, bottom); sy++ {
					for sx := left; sx < max(left+1, right); sx++ {
						r, g, b, a := src.At(b.Min.X+sx, b.Min.Y+sy).RGBA()
						sum[0] += uint64(r)
						sum[1] += uint64(g)
						sum[2] += uint64(b)
						sum[3] += uint64(a)
					}
				}
				n := uint64(max(1, right-left) * max(1, bottom-top))
				dst.SetRGBA(x, y, color.RGBA{R: uint8(sum[0] / n >> 8), G: uint8(sum[1] / n >> 8), B: uint8(sum[2] / n >> 8), A: uint8(sum[3] / n >> 8)})
			}
		}
		out[i] = dst
	}
	for _, pose := range out[:2] {
		keepPrimaryInkBand(pose)
	}
	alignBlink(out[:], ui.PetDesign().SpriteAnchorAlpha)
	return out, nil
}

func frame(m Model, t ui.PetTokens, images [4]*image.RGBA) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, t.Width, t.Height))
	y := t.SpriteTop
	if m.Kind != Idle && m.blink {
		y += t.BreathOffset
	}
	r := image.Rect((t.Width-t.SpriteSize)/2, y, (t.Width+t.SpriteSize)/2, y+t.SpriteSize)
	draw.Draw(dst, r, images[m.Cell()], image.Point{}, draw.Over)
	caption := image.Rect((t.Width-t.CaptionWidth)/2, t.CaptionY, (t.Width+t.CaptionWidth)/2, t.CaptionY+t.CaptionHeight)
	pill(dst, caption, t.Cream)
	text(dst, caption, m.Label(), t.CaptionScale, t.Ink)
	if m.Kind != Idle {
		bg, symbol := t.Alert, "!"
		if m.Kind == Question {
			symbol = "?"
		}
		if m.Kind == Unknown {
			bg, symbol = t.Unknown, "-"
		}
		r := image.Rect(t.BadgeX-t.BadgeRadius, t.BadgeY-t.BadgeRadius, t.BadgeX+t.BadgeRadius, t.BadgeY+t.BadgeRadius)
		pill(dst, r, bg)
		if m.Count > 1 && m.Kind != Unknown {
			symbol = strconv.Itoa(m.Count) + symbol
		}
		scale := t.BadgeScale
		for len(symbol)*6*scale > r.Dx() && scale > 1 {
			scale--
		}
		text(dst, r, symbol, scale, t.Ink)
	}
	return dst
}

func pill(dst *image.RGBA, r image.Rectangle, c color.RGBA) {
	rx, ry := r.Dx()/2, r.Dy()/2
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			dx, dy := x-r.Min.X-rx, y-r.Min.Y-ry
			if dx*dx*ry*ry+dy*dy*rx*rx <= rx*rx*ry*ry {
				dst.SetRGBA(x, y, c)
			}
		}
	}
}

// This fixed 5x7 glyph data is structural; scale, position and colors are tokens.
var glyphs = map[rune][7]string{
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'S': {"01111", "10000", "10000", "01110", "00001", "00001", "11110"},
	'W': {"10001", "10001", "10001", "10101", "10101", "11011", "10001"},
	'I': {"01110", "00100", "00100", "00100", "00100", "00100", "01110"},
	'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
	'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'N': {"10001", "11001", "11001", "10101", "10011", "10011", "10001"},
	'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
	'D': {"11110", "10001", "10001", "10001", "10001", "10001", "11110"},
	'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
	'M': {"10001", "11011", "10101", "10101", "10001", "10001", "10001"},
	'O': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
	'!': {"00100", "00100", "00100", "00100", "00100", "00000", "00100"},
	'?': {"01110", "10001", "00001", "00010", "00100", "00000", "00100"},
	'-': {"00000", "00000", "00000", "11111", "00000", "00000", "00000"},
	'0': {"01110", "10001", "10011", "10101", "11001", "10001", "01110"},
	'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
	'2': {"01110", "10001", "00001", "00010", "00100", "01000", "11111"},
	'3': {"11110", "00001", "00001", "01110", "00001", "00001", "11110"},
	'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
	'5': {"11111", "10000", "10000", "11110", "00001", "00001", "11110"},
	'6': {"01110", "10000", "10000", "11110", "10001", "10001", "01110"},
	'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
	'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
	'9': {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
}

func text(dst *image.RGBA, r image.Rectangle, s string, scale int, c color.RGBA) {
	x := r.Min.X + (r.Dx()-(len(s)*6-1)*scale)/2
	y := r.Min.Y + (r.Dy()-7*scale)/2
	for _, ch := range s {
		for row, line := range glyphs[ch] {
			for col, bit := range line {
				if bit == '1' {
					draw.Draw(dst, image.Rect(x+col*scale, y+row*scale, x+(col+1)*scale, y+(row+1)*scale), image.NewUniform(c), image.Point{}, draw.Src)
				}
			}
		}
		x += 6 * scale
	}
}
