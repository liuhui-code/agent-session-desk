package backend

import (
	"context"
	"encoding/json"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type visibleCapturer interface {
	CaptureVisible(context.Context, session.Session) (session.Preview, error)
}

func TestVisibleCaptureUsesFreshOriginalPaneWithoutHistoryOrTerminalEffects(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	n := len(commandLog(t))
	c, ok := any(b).(visibleCapturer)
	if !ok {
		t.Fatal("backend has no fresh visible-page capture")
	}
	p, e := c.CaptureVisible(context.Background(), s)
	if e != nil || p.Text != "fixture output\n" || p.Truncated {
		t.Fatalf("visible capture: %+v %v", p, e)
	}
	captures := 0
	for _, args := range commandLog(t)[n:] {
		for _, a := range args {
			switch a {
			case "-S", "-E", "-J", "send-keys", "resize-window", "attach-session", "new-session", "kill-session":
				t.Fatalf("visible copy changed input/lifetime or included history: %v", args)
			}
		}
		for _, a := range args {
			if a == "capture-pane" {
				captures++
				if !hasArgPair(args, "-t", s.BackendID+":.%7") {
					t.Fatal("capture changed original pane")
				}
			}
		}
	}
	if captures != 1 {
		t.Fatalf("captures=%d", captures)
	}
}
func hasArgPair(args []string, k, v string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == k && args[i+1] == v {
			return true
		}
	}
	return false
}

func TestVisibleCaptureRejectsIncompleteOrChangedOriginalWithoutCopying(t *testing.T) {
	for _, mode := range []string{"stale-run", "changed-command-before", "changed-tag", "removed-pane", "oversize", "invalid-utf8", "backend-failure"} {
		t.Run(mode, func(t *testing.T) {
			b := newFake(t)
			s := startFake(t, b)
			switch mode {
			case "stale-run":
				s.RunID = strings.Repeat("f", 32)
			case "changed-command-before":
				s.Command += ".different"
			case "changed-tag":
				raw, e := os.ReadFile(os.Getenv("ASD_TEST_BACKEND_STATE"))
				if e != nil {
					t.Fatal(e)
				}
				var state fakeState
				if e = json.Unmarshal(raw, &state); e != nil {
					t.Fatal(e)
				}
				record, e := decodeRecord(s.BackendID, state.Metadata)
				if e != nil {
					t.Fatal(e)
				}
				record.Command += ".changed"
				encoded, e := encodeRecord(record)
				if e != nil {
					t.Fatal(e)
				}
				t.Setenv("ASD_TEST_METADATA_AFTER_CAPTURE", encoded)
			case "removed-pane":
				t.Setenv("ASD_TEST_PANE_REMOVED_DURING_CAPTURE", "1")
			case "oversize", "invalid-utf8":
				data := []byte(strings.Repeat("x", terminal.MaxFrameBytes+1))
				if mode == "invalid-utf8" {
					data = []byte{0xff}
				}
				p := filepath.Join(t.TempDir(), "capture.txt")
				if e := os.WriteFile(p, data, 0600); e != nil {
					t.Fatal(e)
				}
				t.Setenv("ASD_TEST_CAPTURE_FILE", p)
			case "backend-failure":
				t.Setenv("ASD_TEST_BACKEND_FAIL", "capture-pane")
			}
			p, e := b.CaptureVisible(context.Background(), s)
			if e == nil || p.Text != "" {
				t.Fatalf("invalid visible capture escaped: %+v %v", p, e)
			}
		})
	}
}

func TestVisibleCaptureAcceptsC1VTControlsWithoutTreatingThemAsTextEncodingDamage(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	raw := []byte("\x9b31mred\x9b0m\x90synthetic\x9cvisible")
	p := filepath.Join(t.TempDir(), "capture-vt.txt")
	if e := os.WriteFile(p, raw, 0600); e != nil {
		t.Fatal(e)
	}
	t.Setenv("ASD_TEST_CAPTURE_FILE", p)
	got, e := b.CaptureVisible(context.Background(), s)
	if e != nil || got.Text != string(raw) {
		t.Fatalf("source-supported VT control form refused: %v", e)
	}
}
