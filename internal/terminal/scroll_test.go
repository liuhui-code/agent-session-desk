package terminal_test

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestPSMuxScrollTargetsOriginalPaneWithoutTextOrArrowKeys(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	received := make(chan string, 1)
	go func() {
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprint(server, "OK\n")
		for range 4 {
			_, _ = r.ReadString('\n')
		}
		_, _ = fmt.Fprintln(server, firstFrame)
		var got string
		for range 4 {
			line, _ := r.ReadString('\n')
			got += line
		}
		received <- got
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	scroll, ok := any(live).(interface {
		Scroll(context.Context, int, int, bool) error
	})
	if !ok {
		t.Fatal("native terminal cannot scroll its original pane")
	}
	if err = scroll.Scroll(ctx, 3, 1, true); err != nil {
		t.Fatal(err)
	}
	if err = scroll.Scroll(ctx, 3, 1, false); err != nil {
		t.Fatal(err)
	}
	if got := <-received; got != "pane-scroll 7 up 3 1\ndump-state\npane-scroll 7 down 3 1\ndump-state\n" {
		t.Fatalf("scroll changed target or became keyboard input: %q", got)
	}
}

func TestPSMuxCopyModeHidesLiveCursorAndCarriesOffset(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	go func() {
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprint(server, "OK\n")
		for range 4 {
			_, _ = r.ReadString('\n')
		}
		frame := strings.Replace(firstFrame, `"active":true`, `"active":true,"copy_mode":true,"scroll_offset":9,"view_offset":9`, 1)
		_, _ = fmt.Fprintln(server, frame)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	screen, err := live.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !screen.CopyMode || screen.ScrollbackOffset != 9 {
		t.Fatal("native scroll mode/offset lost")
	}
	if screen.Cursor.Visible {
		t.Fatal("scrolled history still advertises a live input cursor")
	}
}
