package terminal

import (
	"context"
	"errors"
	"fmt"
)

const (
	MouseShift uint8 = 1 << iota
	MouseAlt
	MouseCtrl
)

// ProgramMouseConnection carries program gestures separately from the original
// copy-mode Pointer interface. A screen's WantsMouse observation is not a lock.
type ProgramMouseConnection interface {
	Mouse(context.Context, string, string, uint8, int, int) error
}

// Mouse sends a finite, validated native event. The pinned server's any-event
// permission check applies only to exact button 35, so modified hover is refused.
func (c *PSMux) Mouse(ctx context.Context, action, button string, modifiers uint8, x, y int) error {
	if x < 0 || x >= MaxWidth || y < 0 || y >= MaxHeight || modifiers & ^uint8(MouseShift|MouseAlt|MouseCtrl) != 0 {
		return errors.New("unsupported program mouse coordinates or modifiers")
	}
	code, ok := map[string]int{"left": 0, "middle": 1, "right": 2, "none": 3}[button]
	if !ok {
		return errors.New("unsupported program mouse button")
	}
	suffix := "M"
	switch action {
	case "down":
	case "drag":
		code += 32
	case "up":
		suffix = "m"
	case "move":
		if button != "none" || modifiers != 0 {
			return errors.New("unsupported program hover")
		}
		code = 35
	default:
		return errors.New("unsupported program mouse action")
	}
	if action != "move" && button == "none" {
		return errors.New("program gesture needs a held button")
	}
	if modifiers&MouseShift != 0 {
		code |= 4
	}
	if modifiers&MouseAlt != 0 {
		code |= 8
	}
	if modifiers&MouseCtrl != 0 {
		code |= 16
	}
	return c.write(ctx, fmt.Sprintf("pane-mouse %d %d %d %d %s\ndump-state\n", c.options.Pane, code, x, y, suffix))
}
