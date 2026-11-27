package backend

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func commandLog(t *testing.T) [][]string {
	t.Helper()
	data, err := os.ReadFile(os.Getenv("ASD_TEST_BACKEND_LOG"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var result [][]string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var args []string
		if err := json.Unmarshal([]byte(line), &args); err != nil {
			t.Fatal(err)
		}
		result = append(result, args)
	}
	return result
}

func startFake(t *testing.T, b *Backend) session.Session {
	t.Helper()
	s, err := b.Start(context.Background(), session.Launch{Name: "开发 agent", Executable: b.bin, Dir: t.TempDir(), Args: []string{"", `a b`, `quote"inside`, `D:\dir\`, `a&b`, `a;b`, `--name=hello world`, `-literal`}, Env: []string{"ASD_SYNTHETIC=not-a-real-token"}})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestStartPreservesCommandBoundaryAndPrivateMetadata(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	if s.ID == s.RunID || s.TerminalState != session.TerminalAvailable {
		t.Fatalf("unexpected identity/state: %+v", s)
	}
	var argv []string
	for _, args := range commandLog(t) {
		for i, arg := range args {
			if arg == "--" {
				argv = args[i+1:]
			}
		}
	}
	want := []string{b.bin, "", `a b`, `quote"inside`, `D:\dir\`, `a&b`, `a;b`, `--name=hello world`, `-literal`}
	if !reflect.DeepEqual(argv, want) {
		t.Fatalf("OS-delivered backend argv=%q, want %q", argv, want)
	}
	data, err := os.ReadFile(os.Getenv("ASD_TEST_BACKEND_STATE"))
	if err != nil {
		t.Fatal(err)
	}
	var state fakeState
	_ = json.Unmarshal(data, &state)
	r, err := decodeRecord(s.BackendID, state.Metadata)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(r)
	if strings.Contains(string(encoded), "not-a-real-token") || strings.Contains(string(encoded), "quote") {
		t.Fatalf("argv/environment leaked into metadata: %s", encoded)
	}
	if r.PaneID != "%7" {
		t.Fatalf("did not pin initial pane: %q", r.PaneID)
	}
	listed, err := b.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0] != s {
		t.Fatalf("rediscovery changed identity: %+v", listed)
	}
}

func TestChangedRunCannotCaptureAttachOrStop(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	s.RunID = strings.Repeat("f", 32)
	before := len(commandLog(t))
	if _, err := b.Preview(context.Background(), s, 20); err == nil {
		t.Fatal("stale Run captured output")
	}
	if _, err := b.Attach(context.Background(), s); err == nil {
		t.Fatal("stale Run attached")
	}
	if err := b.Stop(context.Background(), s); err == nil {
		t.Fatal("stale Run stopped target")
	}
	for _, args := range commandLog(t)[before:] {
		for _, arg := range args {
			if arg == "capture-pane" || arg == "attach-session" || arg == "kill-session" {
				t.Fatalf("stale target invoked %q", args)
			}
		}
	}
}

func TestPreviewIsBoundedAndNeverResizesOrWrites(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	path := filepath.Join(t.TempDir(), "capture.txt")
	if err := os.WriteFile(path, []byte(strings.Repeat("中文", MaxPreviewBytes)), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ASD_TEST_CAPTURE_FILE", path)
	before := len(commandLog(t))
	preview, err := b.Preview(context.Background(), s, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Truncated || len(preview.Text) > MaxPreviewBytes || !utf8.ValidString(preview.Text) {
		t.Fatalf("invalid preview bound/UTF-8: bytes=%d truncated=%v", len(preview.Text), preview.Truncated)
	}
	for _, args := range commandLog(t)[before:] {
		joined := strings.Join(args, " ")
		for _, forbidden := range []string{"resize", "send-keys", "paste-buffer", "select-pane"} {
			if strings.Contains(joined, forbidden) {
				t.Fatalf("passive preview mutated terminal: %q", args)
			}
		}
		if strings.Contains(joined, "capture-pane") && !strings.Contains(joined, "-S -2000") {
			t.Fatalf("unbounded line request: %q", args)
		}
		if strings.Contains(joined, "capture-pane") {
			var targets []string
			for i := 0; i+1 < len(args); i++ {
				if args[i] == "-t" {
					targets = append(targets, args[i+1])
				}
			}
			if !reflect.DeepEqual(targets, []string{s.BackendID + ":.%7"}) {
				t.Fatalf("psmux could validate a session while capturing an unvalidated pane: %q", args)
			}
		}
	}
}

func TestPreviewDiscardsCaptureWhenOriginalPaneDisappears(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	t.Setenv("ASD_TEST_PANE_REMOVED_DURING_CAPTURE", "1")
	preview, err := b.Preview(context.Background(), s, 20)
	if err == nil || !strings.Contains(err.Error(), "snapshot discarded") || preview.Text != "" {
		t.Fatalf("unverified pane output escaped after capture: %+v, %v", preview, err)
	}
}

func TestAttachDoesNotStealClientOrInheritProbeDeadline(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	ctx, cancel := context.WithCancel(context.Background())
	cmd, err := b.Attach(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if cmd.Cancel != nil {
		t.Fatal("attach lifetime tied to probe context")
	}
	for _, arg := range cmd.Args {
		if arg == "-d" {
			t.Fatal("attach forcibly detaches another client")
		}
	}
	if cmd.SysProcAttr != nil {
		t.Fatal("attach hides its interactive terminal")
	}
	if cmd.Args[len(cmd.Args)-1] != s.BackendID {
		t.Fatalf("wrong attach target: %q", cmd.Args)
	}
}

func TestMetadataFailureCleansOnlyNewTarget(t *testing.T) {
	b := newFake(t)
	t.Setenv("ASD_TEST_BACKEND_FAIL", "set-option")
	_, err := b.Start(context.Background(), session.Launch{Executable: b.bin, Dir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "removed") {
		t.Fatalf("unexpected error: %v", err)
	}
	var created, killed string
	for _, args := range commandLog(t) {
		for i, arg := range args {
			if arg == "-s" {
				created = args[i+1]
			}
			if arg == "kill-session" {
				killed = args[i+2]
			}
			if arg == "kill-server" {
				t.Fatal("cleanup killed whole server")
			}
		}
	}
	if created == "" || killed != created {
		t.Fatalf("cleanup target %q != newly created %q", killed, created)
	}
}

func TestFailedCleanupPreservesUncertainCreationResult(t *testing.T) {
	b := newFake(t)
	t.Setenv("ASD_TEST_BACKEND_FAIL", "set-option|kill-session")
	_, err := b.Start(context.Background(), session.Launch{Executable: b.bin, Dir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "cleanup") || !strings.Contains(err.Error(), "not confirmed") {
		t.Fatalf("uncertain cleanup falsely confirmed: %v", err)
	}
}

func TestListRefusalIsNotAnEmptySuccess(t *testing.T) {
	b := newFake(t)
	t.Setenv("ASD_TEST_BACKEND_FAIL", "list-sessions")
	if _, err := b.List(context.Background()); err == nil {
		t.Fatal("backend refusal swallowed as empty list")
	}
}

func TestCommandDeadlineStopsProbe(t *testing.T) {
	b := newFake(t)
	t.Setenv("ASD_TEST_BACKEND_DELAY", "1")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := b.Probe(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline not preserved: %v", err)
	}
}

func TestDeadOriginalPaneDoesNotAttach(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	path := os.Getenv("ASD_TEST_BACKEND_STATE")
	data, _ := os.ReadFile(path)
	var state fakeState
	_ = json.Unmarshal(data, &state)
	state.Dead = "1"
	data, _ = json.Marshal(state)
	_ = os.WriteFile(path, data, 0600)
	listed, err := b.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].TerminalState != session.TerminalExited {
		t.Fatalf("dead pane falsely available: %+v", listed)
	}
	if _, err := b.Attach(context.Background(), s); err == nil {
		t.Fatal("attached exited original pane")
	}
	if _, err := b.Preview(context.Background(), s, 20); err != nil {
		t.Fatalf("retained dead screen unreadable: %v", err)
	}
}

func TestDifferentActivePaneDoesNotAttachOrChangeFocus(t *testing.T) {
	b := newFake(t)
	s := startFake(t, b)
	path := os.Getenv("ASD_TEST_BACKEND_STATE")
	data, _ := os.ReadFile(path)
	var state fakeState
	_ = json.Unmarshal(data, &state)
	state.Active = "%8"
	data, _ = json.Marshal(state)
	_ = os.WriteFile(path, data, 0600)
	before := len(commandLog(t))
	if _, err := b.Attach(context.Background(), s); err == nil {
		t.Fatal("attached a different active pane")
	}
	for _, args := range commandLog(t)[before:] {
		if strings.Contains(strings.Join(args, " "), "select-pane") {
			t.Fatal("attach preparation changed another client's focus")
		}
	}
}

func TestMetadataLimitAllowsProtocolNewlineButRejectsExtraPayload(t *testing.T) {
	ids := strings.Repeat("a", 32)
	name := "asd_" + ids + "_" + ids
	r := liveRecord{Schema: 1, SessionID: ids, RunID: ids, PaneID: "%7"}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	r.Title = strings.Repeat("x", 3072-len(data))
	value, err := encodeRecord(r)
	if err != nil || len(value) != maxMetadataBytes {
		t.Fatalf("boundary setup bytes=%d err=%v", len(value), err)
	}
	if _, err = decodeRecord(name, value+"\r\n"); err != nil {
		t.Fatalf("valid metadata rejected with protocol newline: %v", err)
	}
	r.Title += "x"
	if _, err = encodeRecord(r); err == nil {
		t.Fatal("oversize metadata accepted")
	}
}

func TestPsmuxVersionRequiresOwnExactStableRelease(t *testing.T) {
	for _, tc := range []struct {
		output string
		valid  bool
	}{
		// Stable 20aa1d7 src/cli.rs::print_version and
		// src/types.rs::build_version_string, including CRLF from Windows.
		{"tmux 3.4.0\r\npsmux 3.4.0 (20aa1d7 2026-10-13)\r\n", true},
		{"psmux 3.4.0 (20aa1d7)", true},
		{"psmux 3.4.0 (20aa1d7, dirty)", true},
		{"psmux 3.4.0 (20aa1d7 2026-10-13, dirty)", true},
		{"psmux 3.4.0 (unknown commit)", true},
		{"tmux 3.4\r\npsmux 3.4.0\r\n", true},
		{"tmux 3.4", false}, {"tmux 3.4.0", false}, {"psmux 3.3.9", false}, {"psmux 3.4.0-dev", false},
		{"psmux 3.5.0", false}, {"psmux 3.4.0 extra", false}, {"psmux -3.4.0", false},
		{"psmux 3.5.0 (20aa1d7 2026-10-13)", false},
		{"psmux 3.4.0-dev (20aa1d7 2026-10-13)", false},
		{"psmux 3.4.0 (arbitrary extra words)", false},
		{"psmux 3.4.0 (20aa1d7 2026-10-13) extra", false},
		{"psmux 3.4.0 (20aa1d7 2026-10-13", false},
	} {
		version, err := parsePsmuxVersion(tc.output)
		if (err == nil) != tc.valid {
			t.Errorf("parsePsmuxVersion(%q)=%v", tc.output, err)
		}
		if tc.valid && version != psmuxBaseline {
			t.Errorf("parsePsmuxVersion(%q) version=%q, want %q", tc.output, version, psmuxBaseline)
		}
	}
}

func TestKnownMultiplexerRewritesFailBeforeLaunch(t *testing.T) {
	b := newFake(t)
	for _, arg := range []string{"-x=value", "-T=中文", ";", `\;`} {
		_, err := b.Start(context.Background(), session.Launch{Executable: b.bin, Args: []string{arg}})
		if err == nil {
			t.Errorf("known rewritten arg %q accepted", arg)
		}
	}
	if len(commandLog(t)) != 0 {
		t.Fatal("invalid transport started or probed backend")
	}
	for _, arg := range []string{"--name=value", "-x=", "a;b", "x;", "-tname", ""} {
		if err := validateTransport("psmux", session.Launch{Args: []string{arg}}); err != nil {
			t.Errorf("unaffected argument %q rejected: %v", arg, err)
		}
	}
	if err := validateTransport("tmux", session.Launch{Args: []string{"end;"}}); err == nil {
		t.Fatal("tmux trailing command separator accepted")
	}
	for _, launch := range []session.Launch{
		{Executable: b.bin, Env: []string{"-x=value"}},
		{Executable: b.bin, Env: []string{"VALUE=trailing "}},
		{Executable: b.bin, Env: []string{"中文=value"}},
		{Executable: ";"}, {Executable: b.bin, Dir: ";"},
	} {
		if _, err := b.Start(context.Background(), launch); err == nil {
			t.Fatal("unsafe payload reached backend")
		}
	}
	if len(commandLog(t)) != 0 {
		t.Fatal("unsafe payload started or probed backend")
	}
	for _, launch := range []session.Launch{
		{Executable: "program;"}, {Dir: "directory;"}, {Env: []string{"KEY=value;"}},
	} {
		if err := validateTransport("tmux", launch); err == nil {
			t.Fatal("tmux command separator in non-argv payload accepted")
		}
	}
}

func TestEnvironmentIsolationDoesNotMutateCaller(t *testing.T) {
	base := []string{"PATH=normal", "token=synthetic", "PSMUX_TARGET_SESSION=foreign", "psmux_data_dir=foreign", "TMUX=/foreign", "TMUX_TMPDIR=/foreign", "PSMUX_NO_WARM=0"}
	original := append([]string(nil), base...)
	env := commandEnv(base, "psmux", "/private/data", "/private/config")
	if !reflect.DeepEqual(base, original) {
		t.Fatal("caller environment changed")
	}
	joined := strings.Join(env, "\n")
	for _, want := range []string{"PATH=normal", "token=synthetic", "PSMUX_NO_WARM=1", "PSMUX_DATA_DIR=/private/data", "PSMUX_CONFIG_FILE=/private/config"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %s", want)
		}
	}
	if strings.Contains(joined, "foreign") {
		t.Fatalf("foreign backend route retained: %s", joined)
	}
	if err := validateTransport("psmux", session.Launch{Env: []string{"psmux_data_dir=foreign"}}); err == nil {
		t.Fatal("launch overrides private backend environment")
	}
}

func TestConfigDoesNotOverwriteUnknownExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private.conf")
	if err := os.WriteFile(path, []byte("existing content"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ensureConfig(path, "new content"); err == nil {
		t.Fatal("unknown config overwritten")
	}
	got, _ := os.ReadFile(path)
	if string(got) != "existing content" {
		t.Fatalf("config modified: %q", got)
	}
}

func TestStartRunUsesReservedIdentityAndRejectsAliases(t *testing.T) {
	b := newFake(t)
	ctx := context.Background()
	sid := "11111111111111111111111111111111"
	rid := "33333333333333333333333333333333"
	v, e := b.StartRun(ctx, session.Launch{Executable: b.bin, Dir: t.TempDir()}, sid, rid)
	if e != nil || v.ID != sid || v.RunID != rid || v.BackendID != "asd_"+sid+"_"+rid {
		t.Fatalf("reserved start: %#v %v", v, e)
	}
	before := len(commandLog(t))
	if _, e = b.StartRun(ctx, session.Launch{Executable: b.bin}, "alias", rid); e == nil {
		t.Fatal("invalid reserved identity accepted")
	}
	if len(commandLog(t)) != before {
		t.Fatal("invalid reservation touched backend")
	}
}
