package terminal

import (
	"context"
	"errors"
	"fmt"
)

// PointerConnection sends a bounded left-pointer gesture to the original pane.
// The server interprets it in its current mode; a screen observation is no lock.
type PointerConnection interface {
	Pointer(context.Context, string, int, int) error
}

// Pointer uses the pinned native semantic protocol, never keyboard input. Signed
// coordinates retain copy-mode edge crossing, within one maximum screen extent.
func (c *PSMux) Pointer(ctx context.Context, action string, x, y int) error {
	if x < -MaxWidth || x >= MaxWidth || y < -MaxHeight || y >= MaxHeight {
		return errors.New("terminal pointer coordinates exceed screen bounds")
	}
	button, suffix := 0, "M"
	switch action {
	case "down":
	case "drag":
		button = 32
	case "up":
		suffix = "m"
	default:
		return errors.New("unsupported terminal pointer action")
	}
	return c.write(ctx, fmt.Sprintf("pane-mouse %d %d %d %d %s\ndump-state\n", c.options.Pane, button, x, y, suffix))
}
