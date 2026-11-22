package terminal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// PSMuxOptions selects one already validated product-owned session and pane.
// Key is an authentication secret and must never appear in diagnostics.
type PSMuxOptions struct {
	Key, Session        string
	Pane, Width, Height int
}

// PSMux is an adapter to the pinned 3.4.0 native client's persistent protocol.
// psmux still owns the pseudoconsole, parser, scrollback and child lifetime.
type PSMux struct {
	writeMu sync.Mutex
	conn    net.Conn
	scanner *bufio.Scanner
	options PSMuxOptions
	frames  chan Screen
	done    chan struct{}
	errMu   sync.Mutex
	err     error
	once    sync.Once
	stop    func() bool
}

func OpenPSMux(ctx context.Context, conn net.Conn, opts PSMuxOptions) (*PSMux, error) {
	c := &PSMux{conn: conn, options: opts, frames: make(chan Screen, 1), done: make(chan struct{})}
	if !validSize(opts.Width, opts.Height) || opts.Pane < 0 || opts.Session == "" || strings.ContainsAny(opts.Key, "\r\n") || opts.Key == "" {
		_ = conn.Close()
		return nil, errors.New("invalid persistent terminal options")
	}
	c.scanner = bufio.NewScanner(conn)
	c.scanner.Buffer(make([]byte, 4096), MaxFrameBytes)
	c.stop = context.AfterFunc(ctx, func() { _ = conn.Close() })
	ok := false
	defer func() {
		if !ok {
			_ = c.Close()
		}
	}()
	deadline := time.Now().Add(3 * time.Second)
	if d, has := ctx.Deadline(); has && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)
	if _, err := fmt.Fprintf(conn, "AUTH %s\n", opts.Key); err != nil {
		return nil, errors.New("terminal authentication write failed")
	}
	if !c.scanner.Scan() || c.scanner.Text() != "OK" {
		return nil, errors.New("terminal authentication failed")
	}
	if _, err := fmt.Fprintf(conn, "PERSISTENT\nclient-attach\nclient-size %d %d\ndump-state\n", opts.Width, opts.Height); err != nil {
		return nil, errors.New("terminal handshake write failed")
	}
	frame, err := c.readScreen()
	if err != nil {
		return nil, err
	}
	c.options.Key = ""
	c.frames <- frame
	_ = conn.SetDeadline(time.Time{})
	ok = true
	go c.readLoop()
	return c, nil
}

// Next returns a recent complete screen. There is at most one pending screen,
// so a slow renderer never accumulates terminal history in the product.
func (c *PSMux) Next(ctx context.Context) (Screen, error) {
	if err := ctx.Err(); err != nil {
		return Screen{}, err
	}
	select {
	case <-c.done:
		return Screen{}, c.failure()
	default:
	}
	select {
	case <-ctx.Done():
		return Screen{}, ctx.Err()
	case <-c.done:
		return Screen{}, c.failure()
	case s := <-c.frames:
		select {
		case <-c.done:
			return Screen{}, c.failure()
		default:
		}
		return s, nil
	}
}

func (c *PSMux) Close() error {
	c.fail(errors.New("terminal connection closed"))
	return nil
}

func (c *PSMux) fail(err error) {
	c.once.Do(func() {
		c.errMu.Lock()
		c.err = err
		c.errMu.Unlock()
		if c.stop != nil {
			c.stop()
		}
		close(c.done)
		_ = c.conn.Close()
	})
}

func (c *PSMux) failure() error {
	c.errMu.Lock()
	defer c.errMu.Unlock()
	return c.err
}

func (c *PSMux) readLoop() {
	for {
		screen, err := c.readScreen()
		if err != nil {
			c.fail(err)
			return
		}
		select {
		case c.frames <- screen:
		default:
			select {
			case <-c.frames:
			default:
			}
			select {
			case c.frames <- screen:
			case <-c.done:
				return
			}
		}
	}
}
