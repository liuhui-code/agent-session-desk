package terminal

import (
	"context"
	"errors"
	"fmt"
)

// CopyDragConnection hands an original pane's live-page drag to native copy
// mode. A completed write is not confirmation that the server changed modes.
type CopyDragConnection interface {
	BeginCopyDrag(context.Context, int, int, int, int) error
}

func (c *PSMux) BeginCopyDrag(ctx context.Context, ax, ay, x, y int) error {
	if ax < 0 || ax >= MaxWidth || ay < 0 || ay >= MaxHeight || x < -MaxWidth || x >= MaxWidth || y < -MaxHeight || y >= MaxHeight {
		return errors.New("terminal drag endpoints exceed screen bounds")
	}
	return c.write(ctx, fmt.Sprintf("copy-drag-begin %d %d %d %d %d\ndump-state\n", c.options.Pane, ax, ay, x, y))
}
