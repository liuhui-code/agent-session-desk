package pet

import (
	"bytes"
	"image"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

func TestIdlePetClickRequestsWorkbenchWithoutInventingAttention(t *testing.T) {
	for _, tc := range []struct {
		name  string
		items []session.AttentionItem
		known bool
	}{
		{name: "idle", known: true},
		{name: "unknown", known: false},
		{name: "unreachable", known: true, items: []session.AttentionItem{{AttentionEvent: session.AttentionEvent{ID: "unreachable", State: "pending"}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests []Activation
			c := NewCompanion(func(a Activation) { requests = append(requests, a) })
			c.Publish(tc.items, tc.known)
			if !c.Activate() {
				t.Fatal("pet click silently ignored instead of opening its workbench")
			}
			if len(requests) != 1 || requests[0] != (Activation{}) {
				t.Fatalf("unverified attention target invented: %+v", requests)
			}
		})
	}
}

func opaqueExtent(src *image.RGBA) (int, int) {
	left, right := src.Bounds().Max.X, -1
	for y := range src.Bounds().Dy() {
		for x := range src.Bounds().Dx() {
			if src.RGBAAt(x, y).A >= 200 {
				left, right = min(left, x), max(right, x)
			}
		}
	}
	return left, right
}

func TestBlinkKeepsTheCatsHorizontalAnchor(t *testing.T) {
	for _, dpi := range []int{96, 144, 192, 384} {
		tokens := ui.PetDesignAtDPI(dpi)
		images, err := sprites(tokens.SpriteSize)
		if err != nil {
			t.Fatal(err)
		}
		left, right := opaqueExtent(images[0])
		blinkLeft, blinkRight := opaqueExtent(images[1])
		delta := left + right - blinkLeft - blinkRight
		if delta < -1 || delta > 1 {
			t.Fatalf("DPI %d: blink shifts body horizontally by %g pixels", dpi, float64(delta)/2)
		}
		m, _ := New(Idle, 1)
		before := frame(m, tokens, images)
		m.Tick()
		if m.Cell() != 1 || bytes.Equal(before.Pix, frame(m, tokens, images).Pix) {
			t.Fatal("blink was removed instead of aligned")
		}
	}
}
