package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// This child executable records OS-delivered argv and serves a deliberately
// narrow fake multiplexer. It does not emulate ConPTY or establish Windows
// compatibility. Every test uses an isolated state and command log.
func TestMain(m *testing.M) {
	if os.Getenv("ASD_TEST_BACKEND_HELPER") == "1" {
		os.Exit(fakeBackend(os.Args[1:]))
	}
	os.Exit(m.Run())
}

type fakeState struct {
	Name     string
	Metadata string
	Dead     string
	Active   string
	Missing  bool
}

func fakeBackend(args []string) int {
	log, err := os.OpenFile(os.Getenv("ASD_TEST_BACKEND_LOG"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return 90
	}
	_ = json.NewEncoder(log).Encode(args)
	_ = log.Close()
	if os.Getenv("ASD_TEST_BACKEND_DELAY") != "" {
		time.Sleep(3 * time.Second)
	}
	var route string
	for len(args) >= 2 && (args[0] == "-L" || args[0] == "-f" || args[0] == "-t") {
		if args[0] == "-t" {
			route = args[1]
		}
		args = args[2:]
	}
	if len(args) == 0 {
		return 91
	}
	if strings.Contains("|"+os.Getenv("ASD_TEST_BACKEND_FAIL")+"|", "|"+args[0]+"|") {
		fmt.Fprint(os.Stderr, "synthetic backend refusal")
		return 1
	}
	statePath := os.Getenv("ASD_TEST_BACKEND_STATE")
	var state fakeState
	data, _ := os.ReadFile(statePath)
	_ = json.Unmarshal(data, &state)
	flag := func(name string) string {
		for i := 0; i+1 < len(args); i++ {
			if args[i] == name {
				return args[i+1]
			}
		}
		return ""
	}
	save := func() { data, _ := json.Marshal(state); _ = os.WriteFile(statePath, data, 0600) }
	switch args[0] {
	case "-V":
		fmt.Print("tmux 3.4\npsmux 3.4.0\n")
	case "new-session":
		state = fakeState{Name: flag("-s"), Dead: "0"}
		save()
	case "list-panes":
		fmt.Println("%7")
	case "set-option":
		if route != state.Name {
			return 92
		}
		state.Metadata = args[len(args)-1]
		save()
	case "show-options":
		if route != state.Name {
			return 93
		}
		fmt.Println(state.Metadata)
	case "list-sessions":
		if state.Name != "" {
			fmt.Println(state.Name)
		}
	case "display-message":
		if args[len(args)-1] == "#{pane_id}" && flag("-t") == state.Name {
			if state.Active == "" {
				fmt.Println("%7")
			} else {
				fmt.Println(state.Active)
			}
			break
		}
		if state.Missing || route != "" || flag("-t") != state.Name+":.%7" {
			return 94
		}
		fmt.Println(state.Dead)
	case "capture-pane":
		if route != "" || flag("-t") != state.Name+":.%7" {
			return 95
		}
		if path := os.Getenv("ASD_TEST_CAPTURE_FILE"); path != "" {
			data, _ := os.ReadFile(path)
			_, _ = os.Stdout.Write(data)
		} else {
			fmt.Print("fixture output\n")
		}
		if os.Getenv("ASD_TEST_PANE_REMOVED_DURING_CAPTURE") == "1" {
			state.Missing = true
			save()
		}
	case "kill-session":
		if flag("-t") != state.Name {
			return 96
		}
		state = fakeState{}
		save()
	default:
		fmt.Fprint(os.Stderr, "unexpected command")
		return 97
	}
	return 0
}

func newFake(t *testing.T) *Backend {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("ASD_TEST_BACKEND_HELPER", "1")
	t.Setenv("ASD_TEST_BACKEND_LOG", filepath.Join(dir, "commands.jsonl"))
	t.Setenv("ASD_TEST_BACKEND_STATE", filepath.Join(dir, "state.json"))
	// Child test executables have no pending work on exit. Avoid the race
	// runtime's one-second exit grace period on every contract invocation.
	t.Setenv("GORACE", os.Getenv("GORACE")+" atexit_sleep_ms=0")
	bin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return &Backend{bin: bin, kind: "psmux", dataDir: dir, namespace: "asd-test", config: filepath.Join(dir, "test.conf")}
}

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

func TestPsmuxVersionRequiresOwnExactStableLine(t *testing.T) {
	for _, tc := range []struct {
		output string
		valid  bool
	}{
		{"tmux 3.4\r\npsmux 3.4.0\r\n", true},
		{"tmux 3.4", false}, {"psmux 3.3.9", false}, {"psmux 3.4.0-dev", false},
		{"psmux 3.5.0", false}, {"psmux 3.4.0 extra", false}, {"psmux -3.4.0", false},
	} {
		_, err := parsePsmuxVersion(tc.output)
		if (err == nil) != tc.valid {
			t.Errorf("parsePsmuxVersion(%q)=%v", tc.output, err)
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
