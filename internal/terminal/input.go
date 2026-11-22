package terminal

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxInputBytes = 64 * 1024

// SendText writes literal UTF-8 through the target pane's byte input path.
// Success means the transport write completed, not that a program accepted or
// executed it. Failed or uncertain input is never retried automatically.
func (c *PSMux) SendText(ctx context.Context, text string) error {
	if !utf8.ValidString(text) || len(text) > MaxInputBytes {
		return errors.New("terminal input exceeds the UTF-8 byte limit")
	}
	if text == "" {
		return nil
	}
	var command strings.Builder
	fmt.Fprintf(&command, "send-keys -t %%%d -H", c.options.Pane)
	for _, b := range []byte(text) {
		fmt.Fprintf(&command, " %02x", b)
	}
	command.WriteByte('\n')
	return c.write(ctx, command.String())
}

func (c *PSMux) write(ctx context.Context, command string) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	select {
	case <-c.done:
		return c.failure()
	default:
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	deadline := time.Now().Add(3 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = c.conn.SetWriteDeadline(deadline)
	stop := context.AfterFunc(ctx, func() { _ = c.conn.Close() })
	defer stop()
	_, err := fmt.Fprint(c.conn, command)
	_ = c.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		_ = c.Close()
		return errors.New("terminal input write failed; result is uncertain, do not replay automatically")
	}
	return nil
}

// Resize changes only this attached client's viewport. Passive previews never
// invoke it. The server remains responsible for terminal reflow and child size.
func (c *PSMux) Resize(ctx context.Context, width, height int) error {
	if !validSize(width, height) {
		return errors.New("terminal viewport exceeds screen bounds")
	}
	return c.write(ctx, fmt.Sprintf("client-size %d %d\ndump-state\n", width, height))
}

// Named keys are converted by psmux using the target program's terminal modes.
var keyName = regexp.MustCompile(`^(?:C-|M-|S-)*(?:[a-zA-Z0-9]|Enter|BSpace|Tab|BTab|Escape|Left|Right|Up|Down|Home|End|PPage|NPage|DC|IC|F[1-9]|F1[0-2])$`)

func (c *PSMux) SendKey(ctx context.Context, name string) error {
	if len(name) > 64 || !keyName.MatchString(name) {
		return errors.New("unsupported terminal key")
	}
	return c.write(ctx, fmt.Sprintf("send-keys -t %%%d %s\n", c.options.Pane, name))
}

// Paste delegates bracketed-paste handling to psmux. It never appends Enter.
func (c *PSMux) Paste(ctx context.Context, text string) error {
	if !utf8.ValidString(text) || len(text) > MaxInputBytes {
		return errors.New("terminal paste exceeds the UTF-8 byte limit")
	}
	if text == "" {
		return nil
	}
	return c.write(ctx, fmt.Sprintf("send-paste %s -t %%%d\n", base64.StdEncoding.EncodeToString([]byte(text)), c.options.Pane))
}

// Refresh resynchronizes the active view without altering terminal size or
// sending user input. Callers bound its cadence and stop it for hidden views;
// native server pushes alone do not guarantee that a final frame was delivered.
func (c *PSMux) Refresh(ctx context.Context) error {
	return c.write(ctx, "dump-state\n")
}

// Scroll delegates one wheel notch at pane-relative coordinates to psmux.
// The server owns scrollback/copy mode and decides whether a child TUI instead
// receives the wheel. It must never be approximated with Up/Down input keys.
func (c *PSMux) Scroll(ctx context.Context, x, y int, up bool) error {
	if x < 0 || x >= MaxWidth || y < 0 || y >= MaxHeight {
		return errors.New("terminal scroll coordinates exceed screen bounds")
	}
	direction := "down"
	if up {
		direction = "up"
	}
	return c.write(ctx, fmt.Sprintf("pane-scroll %d %s %d %d\ndump-state\n", c.options.Pane, direction, x, y))
}
