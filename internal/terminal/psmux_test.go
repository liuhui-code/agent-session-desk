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

const firstFrame = `{"session_name":"asd_test","layout":{"type":"leaf","id":7,"rows":2,"cols":8,"cursor_row":1,"cursor_col":3,"hide_cursor":false,"active":true,"rows_v2":[{"runs":[{"text":"PS> ","fg":"default","bg":"default","flags":0,"width":1}]},{"runs":[{"text":"中文","fg":"#7aa2f7","bg":"default","flags":2,"width":2}]}]}}`

func TestPSMuxConnectionReadsExistingNativeScreen(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	serverErr := make(chan error, 1)
	go func() {
		r := bufio.NewReader(server)
		line, err := r.ReadString('\n')
		if err != nil || line != "AUTH synthetic-key\n" {
			serverErr <- fmt.Errorf("auth: %q %v", line, err)
			return
		}
		if _, err = fmt.Fprint(server, "OK\n"); err != nil {
			serverErr <- err
			return
		}
		for _, want := range []string{"PERSISTENT\n", "client-attach\n", "client-size 8 2\n", "dump-state\n"} {
			line, err = r.ReadString('\n')
			if err != nil || line != want {
				serverErr <- fmt.Errorf("handshake: got %q want %q: %v", line, want, err)
				return
			}
		}
		_, err = fmt.Fprintln(server, firstFrame)
		serverErr <- err
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
	if screen.Width != 8 || screen.Height != 2 || screen.Cursor.X != 3 || screen.Cursor.Y != 1 || !screen.Cursor.Visible {
		t.Fatalf("screen geometry lost: %+v", screen)
	}
	if screen.Rows[1].Runs[0].Text != "中文" || screen.Rows[1].Runs[0].Foreground != "#7aa2f7" {
		t.Fatalf("native screen/style lost: %+v", screen.Rows)
	}
	if err := <-serverErr; err != nil {
		t.Fatal(err)
	}
}

func TestPSMuxTextIsLiteralUTF8AndCannotBecomeAControlCommand(t *testing.T) {
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
		line, _ := r.ReadString('\n')
		received <- line
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	if err := live.SendText(ctx, "中文;\nkill-session\\"); err != nil {
		t.Fatal(err)
	}
	if got := <-received; got != "send-keys -t %7 -H e4 b8 ad e6 96 87 3b 0a 6b 69 6c 6c 2d 73 65 73 73 69 6f 6e 5c\n" {
		t.Fatalf("literal input escaped its byte transport: %q", got)
	}
}

func TestPSMuxScreenUpdatesStayBoundedWhenViewIsSlow(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	sent := make(chan error, 1)
	go func() {
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprint(server, "OK\n")
		for range 4 {
			_, _ = r.ReadString('\n')
		}
		if _, err := fmt.Fprintln(server, firstFrame); err != nil {
			sent <- err
			return
		}
		for i := range 100 {
			frame := strings.Replace(firstFrame, "PS> ", fmt.Sprintf("%03d", i), 1)
			if _, err := fmt.Fprintln(server, frame); err != nil {
				sent <- err
				return
			}
		}
		sent <- nil
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	select {
	case err := <-sent:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("screen producer stalled behind the view")
	}
	for range 2 {
		screen, err := live.Next(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if screen.Rows[0].Runs[0].Text == "099" {
			return
		}
	}
	t.Fatal("old screens queued instead of retaining only the latest")
}

func TestPSMuxResizeChangesViewportWithoutInputOrRestart(t *testing.T) {
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
		a, _ := r.ReadString('\n')
		b, _ := r.ReadString('\n')
		received <- a + b
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	if err := live.Resize(ctx, 92, 24); err != nil {
		t.Fatal(err)
	}
	if got := <-received; got != "client-size 92 24\ndump-state\n" {
		t.Fatalf("resize had extra terminal effects: %q", got)
	}
	if err := live.Resize(ctx, 0, 24); err == nil {
		t.Fatal("invalid dimensions accepted")
	}
}

func TestPSMuxPasteDoesNotSubmitAndSpecialKeysRemainExplicit(t *testing.T) {
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
		a, _ := r.ReadString('\n')
		b, _ := r.ReadString('\n')
		received <- a + b
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	if err := live.Paste(ctx, "中\nx"); err != nil {
		t.Fatal(err)
	}
	if err := live.SendKey(ctx, "C-Left"); err != nil {
		t.Fatal(err)
	}
	if got := <-received; got != "send-paste 5LitCng= -t %7\nsend-keys -t %7 C-Left\n" {
		t.Fatalf("paste/submission boundary changed: %q", got)
	}
	if err := live.SendKey(ctx, "Enter\nkill-session"); err == nil {
		t.Fatal("injected key name accepted")
	}
}

func TestPSMuxNativeStyleRunMaySpanTheWholeRow(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	go func() {
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprint(server, "OK\n")
		for range 4 {
			_, _ = r.ReadString('\n')
		}
		frame := strings.Replace(firstFrame, `"text":"PS> ","fg":"default","bg":"default","flags":0,"width":1`, `"text":"PS>     ","fg":"default","bg":"default","flags":0,"width":8`, 1)
		_, _ = fmt.Fprintln(server, frame)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	s, err := live.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if s.Rows[0].Runs[0].Width != 8 {
		t.Fatalf("run width lost: %+v", s.Rows[0])
	}
}

func TestPSMuxRejectsChangedIdentityAndUnsafeNativeFrames(t *testing.T) {
	for name, frame := range map[string]string{
		"session":        strings.Replace(firstFrame, `"session_name":"asd_test"`, `"session_name":"another"`, 1),
		"pane":           strings.Replace(firstFrame, `"id":7`, `"id":8`, 1),
		"split":          strings.Replace(firstFrame, `"type":"leaf"`, `"type":"split"`, 1),
		"control text":   strings.Replace(firstFrame, `PS> `, `\u001b]52;unsafe`, 1),
		"row overflow":   strings.Replace(firstFrame, `"width":1`, `"width":9`, 1),
		"cursor outside": strings.Replace(firstFrame, `"cursor_col":3`, `"cursor_col":8`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			client, server := net.Pipe()
			defer server.Close()
			go func() {
				r := bufio.NewReader(server)
				_, _ = r.ReadString('\n')
				_, _ = fmt.Fprint(server, "OK\n")
				for range 4 {
					_, _ = r.ReadString('\n')
				}
				_, _ = fmt.Fprintln(server, frame)
			}()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
			if err == nil {
				live.Close()
				t.Fatal("unsafe native frame accepted")
			}
			if strings.Contains(err.Error(), "synthetic-key") || strings.Contains(err.Error(), "unsafe") {
				t.Fatalf("payload leaked into diagnostics: %v", err)
			}
		})
	}
}

func TestPSMuxRefreshGetsScreenWithoutResizingOrWritingInput(t *testing.T) {
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
		line, _ := r.ReadString('\n')
		received <- line
		_, _ = fmt.Fprintln(server, strings.Replace(firstFrame, "PS> ", "new ", 1))
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	if _, err := live.Next(ctx); err != nil {
		t.Fatal(err)
	}
	if err := live.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if got := <-received; got != "dump-state\n" {
		t.Fatalf("refresh changed terminal state: %q", got)
	}
	screen, err := live.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if screen.Rows[0].Runs[0].Text != "new " {
		t.Fatal("refresh did not deliver current screen")
	}
}
