package cli

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestStartKeepsUserArgumentBoundaries(t *testing.T) {
	want := []string{"", "a b", "中文", `a"b`, `end\`, "$(x);&", "--help"}
	args := append([]string{"--name", "开发", "--env", "KEY=a=b", "--", "agent"}, want...)
	l, err := parseLaunch(args, io.Discard, io.Discard)
	if err != nil || l.Executable != "agent" || l.Name != "开发" || !reflect.DeepEqual(l.Args, want) || !reflect.DeepEqual(l.Env, []string{"KEY=a=b"}) {
		t.Fatalf("parse changed launch boundaries: %#v %v", l, err)
	}
}

func TestInvalidCommandsAndHelpDoNotNeedBackend(t *testing.T) {
	missingBackend := filepath.Join(t.TempDir(), "missing-backend")
	for _, tc := range []struct {
		args []string
		code int
		text string
	}{
		{[]string{"--help"}, 0, "M0 source migration"},
		{[]string{"version"}, 0, Version},
		{[]string{"start", "--help"}, 0, "Start options"},
		{[]string{"unknown"}, 1, "unknown command"},
		{[]string{"attach", "0"}, 1, "full session_id"},
		{[]string{"stop"}, 1, "full session_id"},
		{[]string{"start"}, 1, "usage:"},
		{[]string{"list", "--unknown"}, 1, "usage:"},
	} {
		var out, errOut bytes.Buffer
		code := Run(context.Background(), append([]string{"--backend", missingBackend}, tc.args...), strings.NewReader(""), &out, &errOut)
		text := out.String() + errOut.String()
		if code != tc.code || !strings.Contains(text, tc.text) || strings.Contains(text, "executable not found") {
			t.Fatalf("%v -> %d, %q", tc.args, code, text)
		}
	}
}

func TestSessionLookupUsesFullStableIDAndRejectsAmbiguity(t *testing.T) {
	id := strings.Repeat("a", 32)
	a := session.Session{ID: id, RunID: "run-a", Title: "same name"}
	b := session.Session{ID: strings.Repeat("b", 32), RunID: "run-b", Title: "same name"}
	got, err := findSession([]session.Session{b, a}, id)
	if err != nil || got.RunID != "run-a" {
		t.Fatalf("selection used position/name instead of ID: %#v, %v", got, err)
	}
	for _, id := range []string{"same name", "0", "aaaa"} {
		if _, err := findSession([]session.Session{a, b}, id); err == nil {
			t.Fatalf("ambiguous/non-id target accepted: %q", id)
		}
	}
	if _, err := findSession([]session.Session{a, a}, a.ID); err == nil {
		t.Fatal("duplicate backend identity was accepted")
	}
}

func TestCLIErrorCannotExecuteTerminalEscapes(t *testing.T) {
	err := safeError("bad\x1b]52;c;clipboard\a\r\npath")
	if strings.ContainsAny(err, "\x1b\a\r\n") {
		t.Fatalf("unsafe terminal control in error: %q", err)
	}
}
