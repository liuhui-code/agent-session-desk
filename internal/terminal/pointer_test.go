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

func TestPSMuxLeftPointerTargetsOriginalPaneInOrder(t *testing.T) {
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
		for range 6 {
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
	pointer, ok := any(live).(interface {
		Pointer(context.Context, string, int, int) error
	})
	if !ok {
		t.Fatal("original pane has no native left-pointer transport")
	}
	for _, action := range []string{"down", "drag", "up"} {
		if err := pointer.Pointer(ctx, action, 3, 1); err != nil {
			t.Fatal(err)
		}
	}
	if got := <-received; got != "pane-mouse 7 0 3 1 M\ndump-state\npane-mouse 7 32 3 1 M\ndump-state\npane-mouse 7 0 3 1 m\ndump-state\n" {
		t.Fatalf("pointer changed original pane, order or became text input: %q", got)
	}
}

func TestPSMuxPointerRejectsInvalidActionsAndBoundsWithoutWritingThem(t *testing.T) {
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
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	for _, tc := range []struct {
		action string
		x, y   int
	}{
		{"down\nsend-keys Enter", 0, 0}, {"", 0, 0}, {"up", terminal.MaxWidth, 0}, {"down", -terminal.MaxWidth - 1, 0}, {"drag", 0, terminal.MaxHeight}, {"up", 0, -terminal.MaxHeight - 1},
	} {
		if err := live.Pointer(ctx, tc.action, tc.x, tc.y); err == nil {
			t.Fatalf("invalid event accepted: %+v", tc)
		}
	}
	if err := live.Pointer(ctx, "drag", -terminal.MaxWidth, -terminal.MaxHeight); err != nil {
		t.Fatal(err)
	}
	if err := live.Pointer(ctx, "up", terminal.MaxWidth-1, terminal.MaxHeight-1); err != nil {
		t.Fatal(err)
	}
	_ = live.Close()
	want := fmt.Sprintf("pane-mouse 7 32 %d %d M\ndump-state\npane-mouse 7 0 %d %d m\ndump-state\n", -terminal.MaxWidth, -terminal.MaxHeight, terminal.MaxWidth-1, terminal.MaxHeight-1)
	if got := <-received; got != want {
		t.Fatalf("invalid pointer altered stream: %q", got)
	}
}
