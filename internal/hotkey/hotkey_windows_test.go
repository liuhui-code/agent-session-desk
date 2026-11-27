//go:build windows

package hotkey

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNativeLifecycle(t *testing.T) {
	if os.Getenv("ASD_NATIVE_HOTKEY") != "1" {
		t.Skip("explicit desktop")
	}
	for i := 0; i < 3; i++ {
		ctx, c := context.WithTimeout(context.Background(), 100*time.Millisecond)
		e := Run(ctx, "ctrl+alt+space", func() { t.Error("unexpected activation") })
		c()
		if e != nil {
			t.Fatal(e)
		}
	}
}
