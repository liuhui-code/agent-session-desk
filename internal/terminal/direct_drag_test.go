package terminal_test

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestPSMuxDirectDragBeginsWithOriginalPaneAndBothEndpoints(t *testing.T) {
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
		for {
			line, e := r.ReadString('\n')
			got += line
			if e != nil {
				break
			}
		}
		received <- got
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, e := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if e != nil {
		t.Fatal(e)
	}
	defer live.Close()
	drag, ok := any(live).(interface {
		BeginCopyDrag(context.Context, int, int, int, int) error
	})
	if !ok {
		t.Fatal("original pane cannot begin direct native selection")
	}
	for _, points := range [][4]int{{-1, 0, 0, 0}, {terminal.MaxWidth, 0, 0, 0}, {0, -1, 0, 0}, {0, terminal.MaxHeight, 0, 0}, {0, 0, terminal.MaxWidth, 0}, {0, 0, -terminal.MaxWidth - 1, 0}, {0, 0, 0, terminal.MaxHeight}, {0, 0, 0, -terminal.MaxHeight - 1}} {
		if e := drag.BeginCopyDrag(ctx, points[0], points[1], points[2], points[3]); e == nil {
			t.Fatalf("invalid endpoints accepted: %v", points)
		}
	}
	if e := drag.BeginCopyDrag(ctx, 3, 1, 6, 1); e != nil {
		t.Fatal(e)
	}
	if e := drag.BeginCopyDrag(ctx, 0, 0, -terminal.MaxWidth, -terminal.MaxHeight); e != nil {
		t.Fatal(e)
	}
	if e := drag.BeginCopyDrag(ctx, terminal.MaxWidth-1, terminal.MaxHeight-1, terminal.MaxWidth-1, terminal.MaxHeight-1); e != nil {
		t.Fatal(e)
	}
	_ = live.Close()
	want := fmt.Sprintf("copy-drag-begin 7 3 1 6 1\ndump-state\ncopy-drag-begin 7 0 0 %d %d\ndump-state\ncopy-drag-begin 7 %d %d %d %d\ndump-state\n", -terminal.MaxWidth, -terminal.MaxHeight, terminal.MaxWidth-1, terminal.MaxHeight-1, terminal.MaxWidth-1, terminal.MaxHeight-1)
	if got := <-received; got != want {
		t.Fatalf("direct drag lost anchor/target or became input: %q", got)
	}
}
