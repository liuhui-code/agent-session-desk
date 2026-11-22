package terminal_test

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestPSMuxEnterCopyModeTargetsOriginalPaneWithoutTyping(t *testing.T) {
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
	copyMode, ok := any(live).(interface{ EnterCopyMode(context.Context) error })
	if !ok {
		t.Fatal("native terminal has no semantic copy-mode entry")
	}
	if err := copyMode.EnterCopyMode(ctx); err != nil {
		t.Fatal(err)
	}
	if got := <-received; got != "copy-mode -t %7\ndump-state\n" {
		t.Fatalf("copy mode changed target or became child input: %q", got)
	}
}

func TestPSMuxClipboardBoundsAndMalformedDataDoNotBreakScreen(t *testing.T) {
	for _, tc := range []struct {
		name, payload string
		want          bool
	}{
		{"unicode", base64.StdEncoding.EncodeToString([]byte("中文\nselection")), true},
		{"at-limit", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", terminal.MaxClipboardBytes))), true},
		{"over-limit", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", terminal.MaxClipboardBytes+1))), false},
		{"query", "?", false},
		{"bad-base64", "!!", false},
		{"invalid-utf8", base64.StdEncoding.EncodeToString([]byte{255}), false},
		{"nul", base64.StdEncoding.EncodeToString([]byte("one\x00two")), false},
		{"empty", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, server := net.Pipe()
			defer server.Close()
			go func() {
				r := bufio.NewReader(server)
				_, _ = r.ReadString('\n')
				_, _ = fmt.Fprint(server, "OK\n")
				for range 4 {
					_, _ = r.ReadString('\n')
				}
				frame := strings.TrimSuffix(firstFrame, "}") + `,"clipboard_osc52":"` + tc.payload + `"}`
				_, _ = fmt.Fprintln(server, frame)
			}()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
			if err != nil {
				t.Fatal(err)
			}
			defer live.Close()
			if _, err = live.Next(ctx); err != nil {
				t.Fatal("clipboard metadata broke the terminal screen:", err)
			}
			deadline, stop := context.WithTimeout(ctx, 20*time.Millisecond)
			defer stop()
			text, err := live.NextClipboard(deadline)
			if tc.want {
				decoded, _ := base64.StdEncoding.DecodeString(tc.payload)
				if err != nil || text != string(decoded) {
					t.Fatalf("valid clipboard text lost: %v", err)
				}
			} else if !errors.Is(err, context.DeadlineExceeded) || text != "" {
				t.Fatal("invalid clipboard payload reached the host path")
			}
		})
	}
}

func TestPSMuxClipboardSurvivesScreenCoalescingAndKeepsLatest(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	sent := make(chan struct{})
	go func() {
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprint(server, "OK\n")
		for range 4 {
			_, _ = r.ReadString('\n')
		}
		for _, text := range []string{"older", "中文 selection"} {
			frame := strings.TrimSuffix(firstFrame, "}") + `,"clipboard_osc52":"` + base64.StdEncoding.EncodeToString([]byte(text)) + `"}`
			_, _ = fmt.Fprintln(server, frame)
		}
		for range 20 {
			_, _ = fmt.Fprintln(server, firstFrame)
		}
		close(sent)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	source, ok := any(live).(interface {
		NextClipboard(context.Context) (string, error)
	})
	if !ok {
		t.Fatal("clipboard selections disappear with the coalesced screen")
	}
	<-sent
	text, err := source.NextClipboard(ctx)
	if err != nil || text != "中文 selection" {
		t.Fatalf("latest clipboard selection lost: %q %v", text, err)
	}
}

func TestPSMuxSearchPromptRetainsValidatedUTF8ByteCursor(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	go func() {
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprint(server, "OK\n")
		for range 4 {
			_, _ = r.ReadString('\n')
		}
		frame := strings.Replace(firstFrame, `"active":true`, `"active":true,"copy_mode":true`, 1)
		frame = strings.TrimSuffix(frame, "}") + `,"status_message":"(search) 中文","status_label":9,"status_cursor":15}`
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
	if screen.CopyPrompt == nil || screen.CopyPrompt.Text != "(search) 中文" || screen.CopyPrompt.LabelBytes != 9 || screen.CopyPrompt.CursorBytes != 15 || screen.Cursor.Visible {
		t.Fatalf("native UTF-8 copy prompt lost or confused with child cursor: %+v", screen)
	}
}

func TestPSMuxRetainsNativeSelectionAndSearchHighlights(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	go func() {
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprint(server, "OK\n")
		for range 4 {
			_, _ = r.ReadString('\n')
		}
		frame := strings.Replace(firstFrame, `"active":true`, `"active":true,"copy_mode":true,"sel_start_row":0,"sel_start_col":1,"sel_end_row":0,"sel_end_col":3,"sel_mode":"line","copy_cursor_row":0,"copy_cursor_col":2,"copy_hl":[[1,0,1,1]]`, 1)
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
	if screen.Selection == nil || screen.Selection.Mode != "line" || screen.Selection.StartCol != 1 || screen.CopyCursor == nil || screen.CopyCursor.X != 2 || len(screen.CopyHighlights) != 1 || screen.CopyHighlights[0].Kind != 1 {
		t.Fatalf("native copy decorations were discarded: %+v", screen)
	}
}

func TestPSMuxResizeCanHideStaleCopyCursorWithoutDiscardingSearch(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	go func() {
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprint(server, "OK\n")
		for range 4 {
			_, _ = r.ReadString('\n')
		}
		frame := strings.Replace(firstFrame, `"active":true`, `"active":true,"copy_mode":true,"copy_cursor_row":24,"copy_cursor_col":40`, 1)
		frame = strings.TrimSuffix(frame, "}") + `,"status_message":"(search) needle","status_label":9,"status_cursor":15}`
		_, _ = fmt.Fprintln(server, frame)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	live, err := terminal.OpenPSMux(ctx, client, terminal.PSMuxOptions{Key: "synthetic-key", Session: "asd_test", Pane: 7, Width: 8, Height: 2})
	if err != nil {
		t.Fatal("a pre-resize copy cursor discarded an otherwise valid search frame:", err)
	}
	defer live.Close()
	screen, err := live.Next(ctx)
	if err != nil || screen.CopyCursor != nil || screen.CopyPrompt == nil || screen.CopyPrompt.Text != "(search) needle" {
		t.Fatalf("stale copy caret was displayed or search prompt disappeared: %+v %v", screen, err)
	}
}
