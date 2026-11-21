package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestReportObservesExactPayloadAndExit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report 中文.jsonl")
	env := map[string]string{
		"ASD_FIXTURE_REPORT_PATH": path,
		"ASD_FIXTURE_ENV_KEYS":    "ASD_TEST",
		"ASD_TEST":                "synthetic value",
		"ASD_FIXTURE_EXIT_CODE":   "7",
		"NOT_SELECTED":            "not-in-report",
	}
	args := []string{"", "space value", `quote"tail\`, "中文"}
	var stdout, stderr bytes.Buffer
	code := run(strings.NewReader("你好\nexit\n"), &stdout, &stderr, args, func(k string) string { return env[k] })
	if code != 7 || stderr.Len() != 0 {
		t.Fatalf("fixture exit=%d stderr=%s", code, &stderr)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var events []map[string]json.RawMessage
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte{'\n'}) {
		var event map[string]json.RawMessage
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	if len(events) != 4 {
		t.Fatalf("expected started/input/input/exiting: %s", data)
	}
	var gotArgs []string
	var gotEnv map[string]string
	if err := json.Unmarshal(events[0]["args"], &gotArgs); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(events[0]["env"], &gotEnv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotArgs, args) || !reflect.DeepEqual(gotEnv, map[string]string{"ASD_TEST": "synthetic value"}) {
		t.Fatalf("payload changed or unselected environment leaked: %s", data)
	}
	if string(events[3]["code"]) != "7" || string(events[3]["event"]) != `"exiting"` {
		t.Fatalf("exit evidence missing: %s", data)
	}
}

func TestReportCannotReplaceExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing")
	if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run(strings.NewReader("exit\n"), &stdout, &stderr, nil, func(k string) string {
		if k == "ASD_FIXTURE_REPORT_PATH" {
			return path
		}
		return ""
	})
	data, err := os.ReadFile(path)
	if code != 2 || err != nil || string(data) != "keep" || stdout.Len() != 0 {
		t.Fatalf("existing evidence overwritten: code=%d contents=%q err=%v", code, data, err)
	}
}

func TestImmediateExitDoesNotReadInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(failReader{t}, &stdout, &stderr, nil, func(k string) string {
		if k == "ASD_FIXTURE_EXIT_AFTER_START" {
			return "1"
		}
		return ""
	})
	if code != 0 || !strings.Contains(stdout.String(), `"event":"exiting"`) {
		t.Fatalf("immediate exit missing: %d %s %s", code, &stdout, &stderr)
	}
}

type failReader struct{ t *testing.T }

func (r failReader) Read([]byte) (int, error) { r.t.Fatal("immediate exit read stdin"); return 0, nil }

func TestLoadCommandsRequireExplicitOptIn(t *testing.T) {
	for _, enabled := range []string{"", "0"} {
		var stdout, stderr bytes.Buffer
		code := run(strings.NewReader("load 100 256 100\nload-stop\nexit\n"), &stdout, &stderr, nil, func(k string) string {
			if k == "ASD_FIXTURE_ENABLE_LOAD" {
				return enabled
			}
			return ""
		})
		if code != 0 || stderr.Len() != 0 || strings.Contains(stdout.String(), "ASD-LOAD ") || strings.Contains(stdout.String(), `"event":"load-`) {
			t.Fatalf("disabled load mode changed ordinary echo: enable=%q code=%d output=%s stderr=%s", enabled, code, &stdout, &stderr)
		}
		if !strings.Contains(stdout.String(), `"text":"load 100 256 100"`) || !strings.Contains(stdout.String(), `"text":"load-stop"`) {
			t.Fatalf("disabled mode lost ordinary input evidence: %s", &stdout)
		}
	}
	var stdout, stderr bytes.Buffer
	code := run(strings.NewReader("exit\n"), &stdout, &stderr, nil, func(k string) string {
		if k == "ASD_FIXTURE_ENABLE_LOAD" {
			return "true"
		}
		return ""
	})
	if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "ASD_FIXTURE_ENABLE_LOAD") {
		t.Fatalf("invalid opt-in value silently enabled or accepted: code=%d stdout=%s stderr=%s", code, &stdout, &stderr)
	}
}

func TestLoadParameterBounds(t *testing.T) {
	for _, command := range []string{
		"load 100 256 25000", "load 1 32 1", "load 1000 32 60000",
		"load 256 4096 32000", // Exact 1 MiB/s and 32 MiB planned-byte caps.
	} {
		if _, err := parseLoad(strings.Fields(command)); err != nil {
			t.Errorf("valid bounded workload %q rejected: %v", command, err)
		}
	}
	for _, command := range []string{
		"", "load", "load 1 32", "load 1 32 1 extra", "other 1 32 1",
		"load x 32 1", "load 1.5 32 1", "load 99999999999999999999999 32 1",
		"load 0 32 1", "load -1 32 1", "load 1001 32 1",
		"load 1 31 1", "load 1 4097 1", "load 1 -32 1",
		"load 1 32 0", "load 1 32 -1", "load 1 32 60001",
		"load 257 4096 1000",  // One rate step over the byte-rate cap.
		"load 256 4096 32001", // Rounded final line exceeds the total-byte cap.
	} {
		if _, err := parseLoad(strings.Fields(command)); err == nil {
			t.Errorf("invalid or unbounded workload %q accepted", command)
		}
	}
}

type signalledWriter struct {
	data   bytes.Buffer
	writes chan []byte
}

func (w *signalledWriter) Write(data []byte) (int, error) {
	n, err := w.data.Write(data)
	w.writes <- bytes.Clone(data)
	return n, err
}

type pumpHarness struct {
	ticks    chan time.Time
	deadline chan time.Time
	cancel   context.CancelFunc
	done     chan struct{}
	result   loadResult
}

func startTestPump(t *testing.T, writer io.Writer, lineBytes int) *pumpHarness {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	h := &pumpHarness{ticks: make(chan time.Time), deadline: make(chan time.Time), cancel: cancel, done: make(chan struct{})}
	go func() {
		h.result = pumpLoad(ctx, writer, loadSpec{rate: 100, lineBytes: lineBytes, durationMS: 1000}, h.ticks, h.deadline)
		close(h.done)
	}()
	t.Cleanup(func() {
		cancel()
		h.wait(t)
	})
	return h
}

func (h *pumpHarness) tick(t *testing.T) {
	t.Helper()
	select {
	case h.ticks <- time.Now():
	case <-time.After(5 * time.Second):
		t.Fatal("output worker did not consume a controlled tick")
	}
}

func (h *pumpHarness) wait(t *testing.T) loadResult {
	t.Helper()
	select {
	case <-h.done:
		return h.result
	case <-time.After(5 * time.Second):
		t.Fatal("output worker did not finish or join")
		return loadResult{}
	}
}

func TestLoadOutputLengthAccountingAndDeadline(t *testing.T) {
	for _, size := range []int{32, 256, 4096} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			writer := &signalledWriter{writes: make(chan []byte, 3)}
			h := startTestPump(t, writer, size)
			for i := 0; i < 3; i++ {
				h.tick(t)
				select {
				case line := <-writer.writes:
					if len(line) != size || line[size-1] != '\n' || bytes.Count(line, []byte{'\n'}) != 1 || !bytes.HasPrefix(line, []byte("ASD-LOAD ")) {
						t.Fatalf("payload length or framing changed: %q", line)
					}
					for _, char := range line {
						if char >= 128 {
							t.Fatal("load payload is not deterministic ASCII")
						}
					}
				case <-time.After(5 * time.Second):
					t.Fatal("controlled tick produced no output")
				}
			}
			close(h.deadline)
			got := h.wait(t)
			if got.status != "completed" || got.err != nil || got.lines != 3 || got.bytes != int64(3*size) || writer.data.Len() != 3*size {
				t.Fatalf("incorrect finite workload accounting: %+v length=%d", got, writer.data.Len())
			}
		})
	}
}

func TestLoadCancellationJoinsWithoutAnotherTick(t *testing.T) {
	writer := &signalledWriter{writes: make(chan []byte, 1)}
	h := startTestPump(t, writer, 64)
	h.tick(t)
	select {
	case <-writer.writes:
	case <-time.After(5 * time.Second):
		t.Fatal("controlled workload did not start")
	}
	h.cancel()
	got := h.wait(t)
	if got.status != "cancelled" || got.lines != 1 || got.bytes != 64 || got.err != nil {
		t.Fatalf("cancellation lost already-written payload or continued output: %+v", got)
	}
}

type partialFixtureWriter struct{ err error }

func (w partialFixtureWriter) Write([]byte) (int, error) { return 7, w.err }

func TestLoadWriteErrorsReportOnlyActualBytes(t *testing.T) {
	for _, writeErr := range []error{nil, io.ErrClosedPipe} {
		h := startTestPump(t, partialFixtureWriter{writeErr}, 64)
		h.tick(t)
		got := h.wait(t)
		wantErr := writeErr
		if wantErr == nil {
			wantErr = io.ErrShortWrite
		}
		if got.status != "error" || got.lines != 0 || got.bytes != 7 || !errors.Is(got.err, wantErr) {
			t.Fatalf("partial output misreported: %+v, want %v", got, wantErr)
		}
	}
}

type observedFixtureEvent struct {
	Event  string `json:"event"`
	Text   string `json:"text"`
	Reason string `json:"reason"`
	loadEvent
}

// No mutex here: run must serialize the scanner and worker writes. Race tests
// will expose concurrent use. The test reads data only after run has joined.
type observedFixtureWriter struct {
	data   bytes.Buffer
	events chan observedFixtureEvent
}

func (w *observedFixtureWriter) Write(data []byte) (int, error) {
	n, err := w.data.Write(data)
	var event observedFixtureEvent
	if json.Unmarshal(data, &event) == nil && event.Event != "" {
		w.events <- event
	}
	return n, err
}

type runningFixture struct {
	stdin  *io.PipeWriter
	stdout *observedFixtureWriter
	stderr bytes.Buffer
	report string
	done   chan struct{}
	code   int
}

func startLoadFixture(t *testing.T) *runningFixture {
	t.Helper()
	reader, writer := io.Pipe()
	h := &runningFixture{
		stdin: writer, stdout: &observedFixtureWriter{events: make(chan observedFixtureEvent, 64)},
		report: filepath.Join(t.TempDir(), "load-report.jsonl"), done: make(chan struct{}),
	}
	go func() {
		defer reader.Close()
		h.code = run(reader, h.stdout, &h.stderr, nil, func(k string) string {
			switch k {
			case "ASD_FIXTURE_ENABLE_LOAD":
				return "1"
			case "ASD_FIXTURE_REPORT_PATH":
				return h.report
			}
			return ""
		})
		close(h.done)
	}()
	t.Cleanup(func() {
		writer.Close()
		h.wait(t)
	})
	h.event(t, "started")
	return h
}

func (h *runningFixture) send(t *testing.T, line string) {
	t.Helper()
	if _, err := io.WriteString(h.stdin, line+"\n"); err != nil {
		t.Fatal(err)
	}
}

func (h *runningFixture) event(t *testing.T, kind string) observedFixtureEvent {
	t.Helper()
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()
	for {
		select {
		case event := <-h.stdout.events:
			if event.Event == kind {
				return event
			}
		case <-timeout.C:
			t.Fatalf("fixture did not emit %s", kind)
			return observedFixtureEvent{}
		}
	}
}

func (h *runningFixture) wait(t *testing.T) {
	t.Helper()
	select {
	case <-h.done:
		if h.code != 0 || h.stderr.Len() != 0 {
			t.Errorf("fixture failed: code=%d stderr=%s", h.code, &h.stderr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("fixture did not cancel/join its workload on return")
	}
}

func TestLoadFiniteRunReportsTotalsAndReturnsToIdle(t *testing.T) {
	h := startLoadFixture(t)
	h.send(t, "load 100 256 100")
	started := h.event(t, "load-start")
	finished := h.event(t, "load-finished")
	if started.LoadID != 1 || started.Status != "running" || started.LineBytes != 256 || started.LinesPerSecond != 100 || started.DurationMS != 100 || started.LinesWritten != 0 || started.BytesWritten != 0 || started.StartedUTC.IsZero() || started.FinishedUTC != nil {
		t.Fatalf("invalid workload start report: %+v", started)
	}
	if finished.LoadID != started.LoadID || finished.Status != "completed" || finished.BytesWritten != finished.LinesWritten*256 || finished.FinishedUTC == nil || finished.FinishedUTC.Before(started.StartedUTC) {
		t.Fatalf("invalid workload completion report: %+v", finished)
	}
	h.send(t, "idle-marker")
	if event := h.event(t, "input"); event.Text != "idle-marker" {
		t.Fatalf("fixture did not return to ordinary input handling: %+v", event)
	}
	h.send(t, "exit")
	h.wait(t)
	data, err := os.ReadFile(h.report)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("ASD-LOAD ")) {
		t.Fatal("workload payload was duplicated into the observation report")
	}
	var kinds []string
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte{'\n'}) {
		var event observedFixtureEvent
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, event.Event)
	}
	if want := []string{"started", "input", "load-start", "load-finished", "input", "input", "exiting"}; !reflect.DeepEqual(kinds, want) {
		t.Fatalf("observation records grew with output lines or lost order: %v", kinds)
	}
	payloadLines := bytes.Count(h.stdout.data.Bytes(), []byte("ASD-LOAD "))
	if int64(payloadLines) != finished.LinesWritten {
		t.Fatalf("reported count differs from actual stdout: report=%d stdout=%d", finished.LinesWritten, payloadLines)
	}
}

func TestLoadRejectsInvalidAndConcurrentJobsThenCanRestart(t *testing.T) {
	h := startLoadFixture(t)
	h.send(t, "load-stop")
	if event := h.event(t, "load-rejected"); event.Reason != "no_active_load" {
		t.Fatalf("idle stop rejection missing: %+v", event)
	}
	h.send(t, "load 1001 32 10")
	if event := h.event(t, "load-rejected"); event.Reason != "invalid_parameters" {
		t.Fatalf("invalid workload rejection missing: %+v", event)
	}
	h.send(t, "load 10 32 60000")
	first := h.event(t, "load-start")
	h.send(t, "load 10 32 60000")
	if event := h.event(t, "load-rejected"); event.Reason != "load_active" {
		t.Fatalf("concurrent workload was not rejected: %+v", event)
	}
	h.send(t, "load-stop")
	if event := h.event(t, "load-finished"); event.Status != "cancelled" || event.LoadID != first.LoadID {
		t.Fatalf("explicit stop did not cancel the original workload: %+v", event)
	}
	h.send(t, "load 10 32 60000")
	second := h.event(t, "load-start")
	if second.LoadID != first.LoadID+1 {
		t.Fatalf("rejected commands consumed IDs or restart lost identity: first=%+v second=%+v", first, second)
	}
	h.send(t, "exit")
	if event := h.event(t, "load-finished"); event.Status != "cancelled" || event.LoadID != second.LoadID {
		t.Fatalf("exit failed to cancel the restarted workload: %+v", event)
	}
	h.event(t, "exiting")
	h.wait(t)
}

func TestLoadEOFJoinsBeforeExitRecord(t *testing.T) {
	h := startLoadFixture(t)
	h.send(t, "load 10 32 60000")
	h.event(t, "load-start")
	if err := h.stdin.Close(); err != nil {
		t.Fatal(err)
	}
	if event := h.event(t, "load-finished"); event.Status != "cancelled" {
		t.Fatalf("EOF did not cancel the output worker: %+v", event)
	}
	h.event(t, "exiting")
	h.wait(t)
	data, err := os.ReadFile(h.report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(data, []byte("{\"event\":\"exiting\",\"code\":0}\n")) {
		t.Fatalf("worker appended evidence after process exit: %s", data)
	}
}
