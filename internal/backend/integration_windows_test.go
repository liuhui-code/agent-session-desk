//go:build windows

package backend

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// These tests execute a real psmux and a real controlled CLI on Windows.
// They do not run on a cross-compiler or substitute the fakeBackend helper.
// A headless runner does not establish Windows Terminal rendering, IME, focus,
// interactive attach/detach, or a published Windows support matrix.
//
// Opt in with ASD_PSMUX_TEST_BINARY and ASD_FIXTURE_TEST_BINARY. The CLI-process
// scenario additionally uses ASD_PRODUCT_TEST_BINARY. CI verifies the official
// release artifact before setting these paths; an optional expected executable
// digest is checked again using ASD_PSMUX_TEST_SHA256.

type nativeFixtureEvent struct {
	Event string            `json:"event"`
	PID   int               `json:"pid"`
	Args  []string          `json:"args"`
	Cwd   string            `json:"cwd"`
	Env   map[string]string `json:"env"`
	Text  string            `json:"text"`
	Code  int               `json:"code"`
}

type nativeHarness struct {
	t       *testing.T
	ctx     context.Context
	b       *Backend
	opts    Options
	fixture string
	dir     string
	owned   map[string]session.Session
	reports []string
}

func newNativeHarness(t *testing.T) *nativeHarness {
	t.Helper()
	psmux := os.Getenv("ASD_PSMUX_TEST_BINARY")
	if psmux == "" {
		t.Skip("native Windows psmux integration is opt-in: set ASD_PSMUX_TEST_BINARY and ASD_FIXTURE_TEST_BINARY")
	}
	fixture := os.Getenv("ASD_FIXTURE_TEST_BINARY")
	if fixture == "" {
		t.Fatal("ASD_FIXTURE_TEST_BINARY is required when native psmux integration is enabled")
	}
	for _, path := range []string{psmux, fixture} {
		if !filepath.IsAbs(path) {
			t.Fatalf("native test executable must use an absolute path: %q", path)
		}
		st, err := os.Stat(path)
		if err != nil || st.IsDir() {
			t.Fatalf("native test executable is unavailable: %q: %v", path, err)
		}
	}
	digest := nativeFileSHA256(t, psmux)
	if expected := os.Getenv("ASD_PSMUX_TEST_SHA256"); expected != "" && !strings.EqualFold(expected, digest) {
		t.Fatalf("psmux executable SHA-256 differs from the selected artifact: got %s, want %s", digest, expected)
	}
	dir := t.TempDir()
	id, err := session.NewID()
	if err != nil {
		t.Fatal(err)
	}
	opts := Options{Binary: psmux, RuntimeDir: filepath.Join(dir, "运行 data"), Namespace: "asd-native-" + id[:12]}
	b, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	h := &nativeHarness{t: t, ctx: ctx, b: b, opts: opts, dir: dir, fixture: fixture, owned: make(map[string]session.Session)}
	t.Cleanup(h.cleanup)
	info, err := b.Probe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("native Windows backend=%s version=%s sha256=%s namespace=%s", info.Name, info.Version, digest, opts.Namespace)
	if info.Name != "psmux" || info.Version != psmuxBaseline {
		t.Fatalf("unexpected native backend: %+v", info)
	}
	return h
}

func nativeFileSHA256(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (h *nativeHarness) launch(name string, args ...string) (session.Launch, string) {
	h.t.Helper()
	dir := filepath.Join(h.dir, name+" 工作区")
	if err := os.MkdirAll(dir, 0700); err != nil {
		h.t.Fatal(err)
	}
	report := filepath.Join(dir, "fixture-events.jsonl")
	h.reports = append(h.reports, report)
	return session.Launch{
		Name: name, Executable: h.fixture, Args: args, Dir: dir,
		Env: []string{
			"ASD_FIXTURE_REPORT_PATH=" + report,
			"ASD_FIXTURE_ENV_KEYS=ASD_NATIVE_VALUE,ASD_NATIVE_EMPTY",
			"ASD_FIXTURE_EXIT_AFTER_START=0", "ASD_FIXTURE_EXIT_CODE=0",
			"ASD_NATIVE_VALUE=synthetic 中文 value", "ASD_NATIVE_EMPTY=",
		},
	}, report
}

func (h *nativeHarness) start(launch session.Launch) session.Session {
	h.t.Helper()
	s, err := h.b.Start(h.ctx, launch)
	// A failed post-create operation can still leave an owned terminal. Keep
	// the exact generated target for bounded cleanup, never a wildcard/name kill.
	if s.BackendID != "" {
		h.owned[s.BackendID] = s
	}
	if err != nil {
		h.t.Fatalf("native start: %v", err)
	}
	if s.ID == s.RunID || s.ID == "" || s.RunID == "" || s.BackendID == "" {
		h.t.Fatalf("Session/Run/backend identity not established: %+v", s)
	}
	return s
}

func (h *nativeHarness) fresh() *Backend {
	h.t.Helper()
	b, err := New(h.opts)
	if err != nil {
		h.t.Fatal(err)
	}
	return b
}

func (h *nativeHarness) stop(s session.Session) {
	h.t.Helper()
	if err := h.b.Stop(h.ctx, s); err != nil {
		h.t.Fatalf("stop owned session %s: %v", s.ID, err)
	}
	delete(h.owned, s.BackendID)
}

// waitNative retries observation only. It never retries a launch, key write or
// stop, so a timeout cannot duplicate an operation whose result is uncertain.
func waitNative(t *testing.T, parent context.Context, label string, check func(context.Context) error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	var last error
	for {
		if last = check(ctx); last == nil {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("waiting for %s: %v (last observation: %v)", label, ctx.Err(), last)
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func readNativeEvents(path string) ([]nativeFixtureEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) > 256*1024 {
		return nil, errors.New("synthetic fixture report exceeded 256 KiB")
	}
	// A concurrent writer may have started its next record. Only complete
	// newline-terminated records are observations; never parse a partial tail.
	end := bytes.LastIndexByte(data, '\n')
	if end < 0 {
		return nil, errors.New("no complete fixture record yet")
	}
	var events []nativeFixtureEvent
	for _, line := range bytes.Split(data[:end], []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var event nativeFixtureEvent
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, fmt.Errorf("invalid complete fixture record: %w", err)
		}
		events = append(events, event)
	}
	return events, nil
}

func fixtureEvent(t *testing.T, ctx context.Context, path, kind, text string) nativeFixtureEvent {
	t.Helper()
	var found nativeFixtureEvent
	waitNative(t, ctx, "fixture event "+kind, func(context.Context) error {
		events, err := readNativeEvents(path)
		if err != nil {
			return err
		}
		for _, event := range events {
			if event.Event == kind && (text == "" || event.Text == text) {
				found = event
				return nil
			}
		}
		return fmt.Errorf("%s event not observed", kind)
	})
	return found
}

func findNativeSession(items []session.Session, expected session.Session) (session.Session, error) {
	for _, item := range items {
		if item.ID == expected.ID {
			if item.RunID != expected.RunID || item.BackendID != expected.BackendID {
				return item, fmt.Errorf("session identity changed: %+v", item)
			}
			return item, nil
		}
	}
	return session.Session{}, fmt.Errorf("original session %s not found", expected.ID)
}

func (h *nativeHarness) waitState(b *Backend, s session.Session, state session.TerminalState) {
	h.t.Helper()
	waitNative(h.t, h.ctx, "terminal state "+string(state), func(ctx context.Context) error {
		items, err := b.List(ctx)
		if err != nil {
			return err
		}
		item, err := findNativeSession(items, s)
		if err != nil {
			return err
		}
		if item.TerminalState != state {
			return fmt.Errorf("terminal is %q, want %q", item.TerminalState, state)
		}
		return nil
	})
}

func (h *nativeHarness) paneSize(s session.Session) string {
	h.t.Helper()
	r, err := h.b.validateTarget(h.ctx, s)
	if err != nil {
		h.t.Fatal(err)
	}
	out, err := h.b.run(h.ctx, 128, "display-message", "-p", "-t", s.BackendID+":."+r.PaneID, "#{pane_width}x#{pane_height}")
	if err != nil {
		h.t.Fatal(err)
	}
	value := strings.TrimSpace(out.text)
	var width, height int
	if _, err := fmt.Sscanf(value, "%dx%d", &width, &height); err != nil || width < 1 || height < 1 {
		h.t.Fatalf("invalid pane size response %q", value)
	}
	return value
}

func (h *nativeHarness) sendLine(s session.Session, line string) {
	h.t.Helper()
	r, err := h.b.validateTarget(h.ctx, s)
	if err != nil {
		h.t.Fatal(err)
	}
	// In stable 3.4.0 send-keys drops a second -t after the global route.
	// Use one complete session:.pane target, then submit separately. These
	// synthetic line writes are test input, not the future product send API.
	target := s.BackendID + ":." + r.PaneID
	if _, err := h.b.run(h.ctx, 1024, "send-keys", "-t", target, "-l", line); err != nil {
		h.t.Fatalf("write synthetic fixture input: %v", err)
	}
	if _, err := h.b.run(h.ctx, 1024, "send-keys", "-t", target, "Enter"); err != nil {
		h.t.Fatalf("submit synthetic fixture input: %v", err)
	}
}

func (h *nativeHarness) waitPreview(s session.Session, marker string) session.Preview {
	h.t.Helper()
	var preview session.Preview
	waitNative(h.t, h.ctx, "screen marker", func(ctx context.Context) error {
		var err error
		preview, err = h.b.Preview(ctx, s, 60)
		if err != nil {
			return err
		}
		if !strings.Contains(preview.Text, marker) {
			return fmt.Errorf("screen has not shown marker %q", marker)
		}
		return nil
	})
	if !utf8.ValidString(preview.Text) || len(preview.Text) > MaxPreviewBytes {
		h.t.Fatal("preview is invalid UTF-8 or exceeds its byte bound")
	}
	return preview
}

func (h *nativeHarness) cleanup() {
	// Preserve the fixture's synthetic observations before t.TempDir removal.
	// No system environment, registry keys, auth files or unrelated logs are
	// exported. CI may upload this directory even when a test fails.
	if destination := os.Getenv("ASD_NATIVE_EVIDENCE_DIR"); destination != "" {
		dir := filepath.Join(destination, strings.ReplaceAll(h.t.Name(), "/", "_"), h.opts.Namespace)
		if err := os.MkdirAll(dir, 0700); err != nil {
			h.t.Errorf("create native evidence directory: %v", err)
		} else {
			for i, report := range h.reports {
				data, err := os.ReadFile(report)
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				if err != nil || len(data) > 256*1024 {
					h.t.Errorf("read bounded fixture evidence: %v", err)
					continue
				}
				if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("fixture-%d.jsonl", i)), data, 0600); err != nil {
					h.t.Errorf("save fixture evidence: %v", err)
				}
			}
		}
	}
	for name := range h.owned {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		_, killErr := h.b.run(ctx, 4096, "kill-session", "-t", name)
		cancel()
		// Confirm absence even after a successful command: server shutdown can
		// finish after the command returns. This also handles an already-gone
		// exact target without hiding an unrelated list/connection failure.
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		var lastErr error
		confirmed := false
		for ctx.Err() == nil {
			out, listErr := h.b.run(ctx, maxControlBytes, "list-sessions", "-F", "#{session_name}")
			if noServer(listErr) || listErr == nil && !strings.Contains(out.text, name) {
				confirmed = true
				break
			}
			lastErr = listErr
			select {
			case <-ctx.Done():
			case <-time.After(100 * time.Millisecond):
			}
		}
		cancel()
		if !confirmed {
			h.t.Errorf("cleanup of exact owned target %s unconfirmed: kill=%v list=%v", name, killErr, lastErr)
		}
	}
}

// Partial native evidence for TC-LCH-001/002/003/005/009, TC-WIN-007,
// terminal capture and reconnection contracts. No form or interactive UI claim.
func TestWindowsPsmuxNativeLaunchLifecycle(t *testing.T) {
	h := newNativeHarness(t)
	binDir := filepath.Join(h.dir, "工具 box")
	if err := os.Mkdir(binDir, 0700); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(h.fixture)
	if err != nil {
		t.Fatal(err)
	}
	h.fixture = filepath.Join(binDir, "fixture 中文.exe")
	if err := os.WriteFile(h.fixture, data, 0700); err != nil {
		t.Fatal(err)
	}
	wantArgs := []string{"", "with spaces", "中文", `quote"inside`, `tail\`, "a;b&literal", "--model=value", "-x="}
	launch, report := h.launch("参数验证", wantArgs...)
	launch.Env = append(launch.Env, "asd_native_value=最后覆盖 value")
	s := h.start(launch)
	started := fixtureEvent(t, h.ctx, report, "started", "")
	if !reflect.DeepEqual(started.Args, wantArgs) {
		t.Fatalf("native argv changed: got %#v, want %#v", started.Args, wantArgs)
	}
	if !strings.EqualFold(filepath.Clean(started.Cwd), filepath.Clean(launch.Dir)) {
		t.Fatalf("native cwd changed: got %q, want %q", started.Cwd, launch.Dir)
	}
	wantEnv := map[string]string{"ASD_NATIVE_VALUE": "最后覆盖 value", "ASD_NATIVE_EMPTY": ""}
	if !reflect.DeepEqual(started.Env, wantEnv) {
		t.Fatalf("native synthetic environment changed: got %#v, want %#v", started.Env, wantEnv)
	}
	fresh := h.fresh()
	h.waitState(fresh, s, session.TerminalAvailable)
	size := h.paneSize(s)
	h.waitPreview(s, "CLI fixture")
	if after := h.paneSize(s); after != size {
		t.Fatalf("passive preview resized the pane: %s -> %s", size, after)
	}
	cmd, err := fresh.Attach(h.ctx, s)
	if err != nil || cmd == nil || cmd.Process != nil {
		t.Fatalf("prepare attach to original pane (without executing it): cmd=%v err=%v", cmd, err)
	}
	for _, arg := range cmd.Args {
		if arg == "-d" {
			t.Fatal("attach preparation would detach an existing client")
		}
	}
	stale := s
	stale.RunID = strings.Repeat("0", 32)
	if _, err := fresh.Preview(h.ctx, stale, 10); err == nil {
		t.Fatal("stale Run could read the current pane")
	}
	if _, err := fresh.Attach(h.ctx, stale); err == nil {
		t.Fatal("stale Run could prepare an attach")
	}
	if err := fresh.Stop(h.ctx, stale); err == nil {
		t.Fatal("stale Run could stop the current pane")
	}
	marker := "native-input 中文"
	h.sendLine(s, marker)
	fixtureEvent(t, h.ctx, report, "input", marker)
	h.sendLine(s, "exit")
	ended := fixtureEvent(t, h.ctx, report, "exiting", "")
	if ended.Code != 0 {
		t.Fatalf("fixture reported unexpected explicit exit: %d", ended.Code)
	}
	h.waitState(fresh, s, session.TerminalExited)
	h.waitPreview(s, "exiting")
	if _, err := fresh.Attach(h.ctx, s); err == nil {
		t.Fatal("exited fixture was offered as an interactive terminal")
	}
	h.stop(s)
	items, err := fresh.List(h.ctx)
	if err != nil || len(items) != 0 {
		t.Fatalf("explicit stop left a discovered session: %#v, %v", items, err)
	}
}

func TestWindowsPsmuxNativeIsolationAndRejectedLaunches(t *testing.T) {
	h := newNativeHarness(t)
	other := newNativeHarness(t)
	firstLaunch, firstReport := h.launch("first")
	first := h.start(firstLaunch)
	fixtureEvent(t, h.ctx, firstReport, "started", "")
	secondLaunch, secondReport := h.launch("second")
	second := h.start(secondLaunch)
	fixtureEvent(t, h.ctx, secondReport, "started", "")
	otherLaunch, otherReport := other.launch("separate-namespace")
	otherSession := other.start(otherLaunch)
	fixtureEvent(t, other.ctx, otherReport, "started", "")
	otherConfig, err := os.ReadFile(other.b.config)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*session.Launch)
	}{
		{"missing-executable", func(l *session.Launch) { l.Executable = filepath.Join(h.dir, "absent.exe") }},
		{"missing-directory", func(l *session.Launch) { l.Dir = filepath.Join(h.dir, "absent-directory") }},
		{"short-option-rewrite", func(l *session.Launch) { l.Args = []string{"-x=value"} }},
		{"separator", func(l *session.Launch) { l.Args = []string{";", "kill-server"} }},
		{"escaped-separator", func(l *session.Launch) { l.Args = []string{`\;`} }},
		{"environment-trim", func(l *session.Launch) { l.Env = []string{"ASD_NATIVE_VALUE=trailing "} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			launch := firstLaunch
			tc.change(&launch)
			if s, err := h.b.Start(h.ctx, launch); err == nil || s.BackendID != "" {
				t.Fatalf("invalid launch created or reported a backend target: %+v, %v", s, err)
			}
		})
	}
	items, err := h.b.List(h.ctx)
	if err != nil || len(items) != 2 {
		t.Fatalf("rejected launches changed live-session set: %#v, %v", items, err)
	}
	h.stop(first)
	h.waitState(h.fresh(), second, session.TerminalAvailable)
	other.waitState(other.fresh(), otherSession, session.TerminalAvailable)
	if after, err := os.ReadFile(other.b.config); err != nil || !bytes.Equal(after, otherConfig) {
		t.Fatalf("another namespace config was modified: %v", err)
	}
	h.sendLine(second, "second-only")
	fixtureEvent(t, h.ctx, secondReport, "input", "second-only")
	events, err := readNativeEvents(otherReport)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Event == "input" {
			t.Fatalf("test input crossed into another namespace: %+v", event)
		}
	}
	// Only fixture reports and private config are inspected. This test does
	// not claim that a user's existing global setup was exercised in CI.
}

func TestWindowsPsmuxNativeRapidExitRetainsScreen(t *testing.T) {
	h := newNativeHarness(t)
	launch, report := h.launch("rapid-exit", "rapid-marker")
	launch.Env = append(launch.Env, "ASD_FIXTURE_EXIT_AFTER_START=1", "ASD_FIXTURE_EXIT_CODE=7")
	s := h.start(launch)
	started := fixtureEvent(t, h.ctx, report, "started", "")
	if !reflect.DeepEqual(started.Args, []string{"rapid-marker"}) {
		t.Fatalf("rapid process reported wrong arguments: %#v", started.Args)
	}
	ended := fixtureEvent(t, h.ctx, report, "exiting", "")
	if ended.Code != 7 {
		t.Fatalf("controlled nonzero exit not observed: %+v", ended)
	}
	h.waitState(h.fresh(), s, session.TerminalExited)
	h.waitPreview(s, "rapid-marker")
	if _, err := h.b.Attach(h.ctx, s); err == nil {
		t.Fatal("rapidly exited process was offered as live attach")
	}
	// Code 7 comes from the independent fixture record. The product model
	// intentionally does not derive an exit code from pane_dead_status.
}

func TestWindowsPsmuxNativeMissingOriginalPaneNeverFallsBack(t *testing.T) {
	h := newNativeHarness(t)
	launch, report := h.launch("original", "original-pane-marker")
	s := h.start(launch)
	fixtureEvent(t, h.ctx, report, "started", "")
	original, err := h.b.readRecord(h.ctx, s.BackendID)
	if err != nil {
		t.Fatal(err)
	}
	next, nextReport := h.launch("sentinel", "other-pane-sentinel")
	args := []string{"new-window", "-t", s.BackendID, "-P", "-F", "#{pane_id}", "-n", "sentinel", "-c", next.Dir}
	for _, value := range next.Env {
		args = append(args, "-e", value)
	}
	// Stable new-window preserves direct argv when it has more than one
	// command token. The controlled fixture and marker deliberately meet that
	// contract, and the fresh report path cannot replace the original report.
	args = append(args, "--", next.Executable, "other-pane-sentinel")
	out, err := h.b.run(h.ctx, maxControlBytes, args...)
	if err != nil {
		t.Fatalf("create owned alternate window: %v", err)
	}
	otherPane := strings.TrimSpace(out.text)
	if !panePattern.MatchString(otherPane) || otherPane == original.PaneID {
		t.Fatalf("alternate pane identity not returned: %q", otherPane)
	}
	fixtureEvent(t, h.ctx, nextReport, "started", "")
	preview := h.waitPreview(s, "original-pane-marker")
	if strings.Contains(preview.Text, "other-pane-sentinel") {
		t.Fatal("preview followed another active pane while the original still existed")
	}
	if _, err := h.b.Attach(h.ctx, s); err == nil {
		t.Fatal("attach would enter the alternate active pane")
	}
	if _, err := h.b.run(h.ctx, 4096, "kill-pane", "-t", s.BackendID+":."+original.PaneID); err != nil {
		t.Fatalf("remove exact original test pane: %v", err)
	}
	waitNative(t, h.ctx, "original pane removal", func(ctx context.Context) error {
		out, err := h.b.run(ctx, maxControlBytes, "list-panes", "-a", "-t", s.BackendID, "-F", "#{pane_id}")
		if err != nil {
			return err
		}
		foundOther := false
		for _, pane := range strings.Fields(out.text) {
			if pane == original.PaneID {
				return errors.New("original pane still exists")
			}
			foundOther = foundOther || pane == otherPane
		}
		if !foundOther {
			return errors.New("alternate pane also disappeared")
		}
		return nil
	})
	if preview, err := h.b.Preview(h.ctx, s, 60); err == nil || strings.Contains(preview.Text, "other-pane-sentinel") {
		t.Fatalf("missing original pane fell back to another pane: preview=%q error=%v", preview.Text, err)
	}
	if _, err := h.b.Attach(h.ctx, s); err == nil {
		t.Fatal("missing original pane could prepare an attach to another window")
	}
}

func TestWindowsPsmuxNativeCLIProcessExitAndReconnect(t *testing.T) {
	if os.Getenv("ASD_PSMUX_TEST_BINARY") == "" {
		t.Skip("native Windows psmux integration is opt-in")
	}
	product := os.Getenv("ASD_PRODUCT_TEST_BINARY")
	if product == "" {
		t.Skip("set ASD_PRODUCT_TEST_BINARY to check independent product processes")
	}
	if !filepath.IsAbs(product) {
		t.Fatal("ASD_PRODUCT_TEST_BINARY must be an absolute path")
	}
	h := newNativeHarness(t)
	launch, report := h.launch("separate-cli-process", "cli-marker")
	startArgs := []string{"start", "--name", launch.Name, "--cwd", launch.Dir}
	for _, value := range launch.Env {
		startArgs = append(startArgs, "--env", value)
	}
	startArgs = append(startArgs, "--", launch.Executable)
	startArgs = append(startArgs, launch.Args...)
	out, err := h.productCommand(product, startArgs...)
	if err != nil {
		t.Fatalf("CLI start process failed: %v", err)
	}
	var created session.Session
	if err := json.Unmarshal(out, &created); err != nil || !namePattern.MatchString(created.BackendID) {
		t.Fatalf("CLI returned invalid created identity: %q, %v", out, err)
	}
	h.owned[created.BackendID] = created
	started := fixtureEvent(t, h.ctx, report, "started", "")
	if started.PID <= 0 {
		t.Fatalf("fixture did not report its process identity: %+v", started)
	}
	handle, err := syscall.OpenProcess(syscall.SYNCHRONIZE|syscall.PROCESS_TERMINATE, false, uint32(started.PID))
	if err != nil {
		t.Fatalf("open owned fixture process after CLI exit: %v", err)
	}
	defer syscall.CloseHandle(handle)
	defer func() {
		// A stop regression must still leave no synthetic child behind. This
		// handle identifies only the exact fixture observed above, not a name
		// or a PID that could be reused. Cleanup never turns a failed stop into
		// a successful assertion; it runs after the test's observation.
		if state, err := syscall.WaitForSingleObject(handle, 0); err == nil && state == syscall.WAIT_TIMEOUT {
			t.Log("cleanup: terminating the exact still-running test fixture after the scenario")
			if err := syscall.TerminateProcess(handle, 125); err != nil {
				t.Errorf("cleanup of owned fixture handle failed: %v", err)
			}
		}
	}()
	state, err := syscall.WaitForSingleObject(handle, 0)
	if err != nil || state != syscall.WAIT_TIMEOUT {
		t.Fatalf("fixture did not survive the completed CLI process: wait=%d, %v", state, err)
	}
	out, err = h.productCommand(product, "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var items []session.Session
	if err := json.Unmarshal(out, &items); err != nil {
		t.Fatal(err)
	}
	if _, err := findNativeSession(items, created); err != nil {
		t.Fatal(err)
	}
	h.waitPreview(created, "CLI fixture")
	out, err = h.productCommand(product, "preview", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		SessionID string          `json:"session_id"`
		RunID     string          `json:"run_id"`
		Kind      string          `json:"kind"`
		Snapshot  session.Preview `json:"snapshot"`
	}
	if err := json.Unmarshal(out, &snapshot); err != nil || snapshot.SessionID != created.ID || snapshot.RunID != created.RunID || snapshot.Kind != "screen_snapshot" {
		t.Fatalf("reopened CLI snapshot lost provenance: %q, %v", out, err)
	}
	if _, err := h.productCommand(product, "stop", created.ID); err != nil {
		t.Fatal(err)
	}
	delete(h.owned, created.BackendID)
	waitNative(t, h.ctx, "owned fixture process exit after explicit stop", func(context.Context) error {
		state, err := syscall.WaitForSingleObject(handle, 0)
		if err != nil {
			return err
		}
		if state != syscall.WAIT_OBJECT_0 {
			return errors.New("owned fixture process is still alive")
		}
		return nil
	})
	out, err = h.productCommand(product, "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &items); err != nil || len(items) != 0 {
		t.Fatalf("stopped CLI session was rediscovered: %q, %v", out, err)
	}
}

func (h *nativeHarness) productCommand(binary string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(h.ctx, 40*time.Second)
	defer cancel()
	all := []string{"--backend", h.opts.Binary, "--runtime-dir", h.opts.RuntimeDir, "--namespace", h.opts.Namespace}
	cmd := exec.CommandContext(ctx, binary, append(all, args...)...)
	configureCommand(cmd)
	cmd.WaitDelay = time.Second
	stdout, stderr := &tailBuffer{limit: 128 * 1024}, &tailBuffer{limit: 8 * 1024}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Run(); err != nil {
		return stdout.data, fmt.Errorf("product command %s: %w: %s", args[0], err, strings.TrimSpace(string(stderr.data)))
	}
	if stdout.truncated {
		return stdout.data, errors.New("product command output exceeded native test limit")
	}
	return stdout.data, nil
}
