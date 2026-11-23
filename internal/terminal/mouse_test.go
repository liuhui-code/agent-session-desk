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

func TestPSMuxProgramMousePreservesRightButtonAndModifiers(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	received := make(chan string, 1)
	go func() {
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprintln(server, "OK")
		for range 4 {
			_, _ = r.ReadString('\n')
		}
		_, _ = fmt.Fprintln(server, firstFrame)
		var got string
		for range 6 {
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
	mouse, ok := any(live).(interface {
		Mouse(context.Context, string, string, uint8, int, int) error
	})
	if !ok {
		t.Fatal("program mouse has no button/modifier transport")
	}
	for _, action := range []string{"down", "drag", "up"} {
		if err := mouse.Mouse(ctx, action, "right", 5, 3, 1); err != nil {
			t.Fatal(err)
		}
	}
	if got := <-received; got != "pane-mouse 7 22 3 1 M\ndump-state\npane-mouse 7 54 3 1 M\ndump-state\npane-mouse 7 22 3 1 m\ndump-state\n" {
		t.Fatalf("right Shift+Ctrl gesture changed: %q", got)
	}
}

func TestPSMuxProgramMouseMatrixAndRejectedEventsNeverAlterWire(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	received := make(chan string, 1)
	go func() {
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprintln(server, "OK")
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	var want strings.Builder
	for base, button := range []string{"left", "middle", "right"} {
		for modifiers, wire := range []int{0, 4, 8, 12, 16, 20, 24, 28} {
			for _, action := range []string{"down", "drag", "up"} {
				if err = live.Mouse(ctx, action, button, uint8(modifiers), 3, 1); err != nil {
					t.Fatal(err)
				}
				code := base + wire
				suffix := "M"
				if action == "drag" {
					code += 32
				}
				if action == "up" {
					suffix = "m"
				}
				fmt.Fprintf(&want, "pane-mouse 7 %d 3 1 %s\ndump-state\n", code, suffix)
			}
		}
	}
	if err = live.Mouse(ctx, "move", "none", 0, 7, 1); err != nil {
		t.Fatal(err)
	}
	want.WriteString("pane-mouse 7 35 7 1 M\ndump-state\n")
	for _, v := range []struct {
		action, button string
		mods           uint8
		x, y           int
	}{
		{"down", "none", 0, 0, 0}, {"drag", "none", 0, 0, 0}, {"up", "none", 0, 0, 0}, {"move", "left", 0, 0, 0}, {"move", "none", 1, 0, 0},
		{"down\nsend-keys Enter", "left", 0, 0, 0}, {"down", "unknown", 0, 0, 0}, {"down", "right", 8, 0, 0}, {"down", "left", 0, -1, 0}, {"down", "left", 0, 0, -1}, {"up", "right", 0, terminal.MaxWidth, 0}, {"drag", "middle", 0, 0, terminal.MaxHeight},
	} {
		if err = live.Mouse(ctx, v.action, v.button, v.mods, v.x, v.y); err == nil {
			t.Fatalf("invalid event accepted: %+v", v)
		}
	}
	_ = live.Close()
	if got := <-received; got != want.String() {
		t.Fatalf("matrix or rejection changed native wire:\ngot %q\nwant %q", got, want.String())
	}
}
