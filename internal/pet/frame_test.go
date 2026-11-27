package pet

import (
	"crypto/sha256"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

func TestDemoSpritesAndFramesRemainTransparentPremultipliedAndBounded(t *testing.T) {
	for _, dpi := range []int{96, 144, 192} {
		tokens := ui.PetDesignAtDPI(dpi)
		images, err := sprites(tokens.SpriteSize)
		if err != nil {
			t.Fatal(err)
		}
		seen := map[[32]byte]bool{}
		for i, src := range images {
			hash := sha256.Sum256(src.Pix)
			if seen[hash] {
				t.Fatal("distinct sprite pose missing", i)
			}
			seen[hash] = true
			if len(src.Pix) != tokens.SpriteSize*tokens.SpriteSize*4 {
				t.Fatal("unbounded sprite cache")
			}
		}
		for _, kind := range []Kind{Idle, Approval, Question, Unknown} {
			m, _ := New(kind, 99)
			pixels := frame(m, tokens, images)
			if len(pixels.Pix) != tokens.Width*tokens.Height*4 {
				t.Fatal("frame overflow")
			}
			clear, solid := 0, 0
			for i := 0; i < len(pixels.Pix); i += 4 {
				a := pixels.Pix[i+3]
				if a == 0 {
					clear++
				}
				if a == 255 {
					solid++
				}
				for j := range 3 {
					if pixels.Pix[i+j] > a {
						t.Fatal("ULW_ALPHA received straight-alpha pixel")
					}
				}
			}
			if clear == 0 || solid == 0 {
				t.Fatal("lost transparent shape or visible character")
			}
			if path := os.Getenv("ASD_PET_PREVIEW_DIR"); path != "" {
				if err = os.MkdirAll(path, 0700); err != nil {
					t.Fatal(err)
				}
				f, err := os.OpenFile(filepath.Join(path, fmt.Sprintf("%s-%d.png", kind, dpi)), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(f, pixels)
				f.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestDemoPositionHandlesNegativeMonitorCoordinates(t *testing.T) {
	for _, tc := range []struct{ x, y, w, h, l, t, r, b, wx, wy int }{
		{-2000, -900, 160, 176, -1920, 0, 0, 1080, -1920, 0},
		{-10, 1070, 160, 176, -1920, 0, 0, 1080, -160, 904},
		{500, 500, 160, 176, 0, 0, 1920, 1080, 500, 500},
		{500, 500, 160, 176, 0, 0, 80, 80, 0, 0},
	} {
		x, y := clampPosition(tc.x, tc.y, tc.w, tc.h, tc.l, tc.t, tc.r, tc.b)
		if x != tc.wx || y != tc.wy {
			t.Fatalf("position %d,%d", x, y)
		}
	}
}

func TestUnknownPetHasNoBadgeAndKnownRequestsKeepTheirBadge(t *testing.T) {
	for _, dpi := range []int{96, 144, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			tokens := ui.PetDesignAtDPI(dpi)
			images, err := sprites(tokens.SpriteSize)
			if err != nil {
				t.Fatal(err)
			}
			base := frame(Model{Kind: Idle}, tokens, images)
			for _, kind := range []Kind{Unknown, Approval, Question} {
				pixels := frame(Model{Kind: kind, Count: 2}, tokens, images)
				different := false
				for y := tokens.BadgeY - tokens.BadgeRadius; y < tokens.BadgeY+tokens.BadgeRadius; y++ {
					for x := tokens.BadgeX - tokens.BadgeRadius; x < tokens.BadgeX+tokens.BadgeRadius; x++ {
						if pixels.RGBAAt(x, y) != base.RGBAAt(x, y) {
							different = true
						}
					}
				}
				if kind == Unknown && different {
					t.Fatal("LINK still draws a badge")
				}
				if kind != Unknown && !different {
					t.Fatal("known blocking request lost its badge", kind)
				}
			}
		})
	}
}
