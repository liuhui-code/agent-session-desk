// asd-fixture is a controlled CLI for the M0 launch/attach checks. It is not an
// agent and does not emit readiness, completion, or attention protocol events.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, os.Args[1:], os.Getenv))
}

// The optional JSONL report is an independent observation channel for native
// integration tests: terminal wrapping must not become the argv test oracle.
// It is enabled only by an explicit test-owned path and never replaces a file.
func run(stdin io.Reader, stdout, stderr io.Writer, args []string, getenv func(string) string) int {
	var report *os.File
	if path := getenv("ASD_FIXTURE_REPORT_PATH"); path != "" {
		var err error
		report, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			fmt.Fprintln(stderr, "fixture report:", err)
			return 2
		}
		defer report.Close()
	}
	exitCode := 0
	if value := getenv("ASD_FIXTURE_EXIT_CODE"); value != "" {
		var err error
		exitCode, err = strconv.Atoi(value)
		if err != nil || exitCode < 0 || exitCode > 255 {
			fmt.Fprintln(stderr, "ASD_FIXTURE_EXIT_CODE must be an integer from 0 to 255")
			return 2
		}
	}
	loadEnabled := getenv("ASD_FIXTURE_ENABLE_LOAD")
	if loadEnabled != "" && loadEnabled != "0" && loadEnabled != "1" {
		fmt.Fprintln(stderr, "ASD_FIXTURE_ENABLE_LOAD must be 0 or 1")
		return 2
	}
	sink := &fixtureSink{stdout: stdout}
	if report != nil {
		sink.report = report
	}
	loads := &loadController{enabled: loadEnabled == "1", sink: sink}
	// The input scanner stays on this goroutine. Only a finite output job gets
	// a worker, and every return joins it before the report is closed. Cancelling
	// a context cannot interrupt an arbitrary blocked Writer; native experiment
	// harnesses still need an outer process deadline for blocked terminal I/O.
	defer loads.stop()
	fail := func(err error) int {
		fmt.Fprintln(stderr, "fixture:", err)
		return 2
	}
	finish := func(code int) int {
		if err := loads.stop(); err != nil {
			return fail(err)
		}
		if err := sink.emit(struct {
			Event string `json:"event"`
			Code  int    `json:"code"`
		}{"exiting", code}); err != nil {
			return fail(err)
		}
		return code
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	env := map[string]string{}
	// Only explicitly selected synthetic keys are printed, never the full env.
	for _, key := range strings.Split(getenv("ASD_FIXTURE_ENV_KEYS"), ",") {
		if key != "" {
			env[key] = getenv(key)
		}
	}
	if err := sink.emit(struct {
		Event string            `json:"event"`
		PID   int               `json:"pid"`
		Args  []string          `json:"args"`
		Cwd   string            `json:"cwd"`
		Env   map[string]string `json:"env"`
	}{"started", os.Getpid(), args, cwd, env}); err != nil {
		return fail(err)
	}
	if getenv("ASD_FIXTURE_EXIT_AFTER_START") == "1" {
		return finish(exitCode)
	}
	if _, err := fmt.Fprintln(sink, "CLI fixture · 中文 · e\u0301 · 🧪 — type exit to end"); err != nil {
		return fail(err)
	}
	input := bufio.NewScanner(stdin)
	input.Buffer(make([]byte, 4096), 64*1024)
	for input.Scan() {
		line := input.Text()
		if err := sink.emit(struct {
			Event string `json:"event"`
			Text  string `json:"text"`
		}{"input", line}); err != nil {
			return fail(err)
		}
		if line == "exit" {
			return finish(exitCode)
		}
		if err := loads.handle(line); err != nil {
			return fail(err)
		}
	}
	if err := input.Err(); err != nil {
		fmt.Fprintln(stderr, "fixture input:", err)
		return finish(2)
	}
	return finish(exitCode)
}

// fixtureSink serializes complete records and payload writes. Only observation
// records reach the JSONL report; the synthetic output workload goes to stdout.
// There is no output queue or buffered report writer.
type fixtureSink struct {
	mu     sync.Mutex
	stdout io.Writer
	report io.Writer
}

func (s *fixtureSink) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeChecked(s.stdout, data)
}

func (s *fixtureSink) emit(value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.report != nil {
		// Direct writes expose each completed record while stdin is waiting.
		if _, err := writeChecked(s.report, data); err != nil {
			return err
		}
	}
	_, err = writeChecked(s.stdout, data)
	return err
}

func writeChecked(w io.Writer, data []byte) (int, error) {
	n, err := w.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	return n, err
}

// These bounds limit this opt-in test fixture, not the product's memory or
// throughput budget. BYTES includes one LF and output is deterministic ASCII.
const (
	loadMaxRate         = 1000
	loadMinLineBytes    = 32
	loadMaxLineBytes    = 4096
	loadMaxDurationMS   = 60000
	loadMaxBytesPerSec  = 1 << 20
	loadMaxPlannedBytes = 32 << 20
)

type loadSpec struct {
	rate, lineBytes, durationMS int
}

func parseLoad(fields []string) (loadSpec, error) {
	var spec loadSpec
	if len(fields) != 4 || fields[0] != "load" {
		return spec, errors.New("use load RATE BYTES DURATION_MS")
	}
	values := []*int{&spec.rate, &spec.lineBytes, &spec.durationMS}
	for i, target := range values {
		value, err := strconv.ParseInt(fields[i+1], 10, 32)
		if err != nil {
			return spec, errors.New("load parameters must be decimal integers")
		}
		*target = int(value)
	}
	if spec.rate < 1 || spec.rate > loadMaxRate {
		return spec, fmt.Errorf("load RATE must be from 1 to %d", loadMaxRate)
	}
	if spec.lineBytes < loadMinLineBytes || spec.lineBytes > loadMaxLineBytes {
		return spec, fmt.Errorf("load BYTES must be from %d to %d, including LF", loadMinLineBytes, loadMaxLineBytes)
	}
	if spec.durationMS < 1 || spec.durationMS > loadMaxDurationMS {
		return spec, fmt.Errorf("load DURATION_MS must be from 1 to %d", loadMaxDurationMS)
	}
	bytesPerSecond := int64(spec.rate) * int64(spec.lineBytes)
	if bytesPerSecond > loadMaxBytesPerSec {
		return spec, fmt.Errorf("load rate exceeds the fixture limit of %d bytes/second", loadMaxBytesPerSec)
	}
	// Round up conservatively, so even a fractional final interval fits the
	// finite workload bound. Widen before multiplying to avoid int overflow.
	plannedLines := (int64(spec.rate)*int64(spec.durationMS) + 999) / 1000
	if plannedLines*int64(spec.lineBytes) > loadMaxPlannedBytes {
		return spec, fmt.Errorf("load exceeds the fixture limit of %d planned bytes", loadMaxPlannedBytes)
	}
	return spec, nil
}

type loadEvent struct {
	Event          string     `json:"event"`
	LoadID         uint64     `json:"load_id"`
	LineBytes      int        `json:"line_bytes"`
	LinesPerSecond int        `json:"lines_per_second"`
	DurationMS     int        `json:"duration_ms"`
	LinesWritten   int64      `json:"lines_written"`
	BytesWritten   int64      `json:"bytes_written"`
	Status         string     `json:"status"`
	StartedUTC     time.Time  `json:"started_utc"`
	FinishedUTC    *time.Time `json:"finished_utc,omitempty"`
	Error          string     `json:"error,omitempty"`
}

type loadJob struct {
	cancel context.CancelFunc
	done   chan struct{}
	err    error // Read only after done closes.
}

type loadController struct {
	enabled bool
	sink    *fixtureSink
	nextID  uint64
	active  *loadJob // Only the stdin goroutine reads or replaces this pointer.
}

func (c *loadController) reject(reason string, err error) error {
	return c.sink.emit(struct {
		Event  string `json:"event"`
		Status string `json:"status"`
		Reason string `json:"reason"`
		Error  string `json:"error"`
	}{"load-rejected", "rejected", reason, err.Error()})
}

func (c *loadController) handle(line string) error {
	if !c.enabled {
		return nil
	}
	fields := strings.Fields(line)
	if len(fields) == 0 || fields[0] != "load" && fields[0] != "load-stop" {
		return nil
	}
	if c.active != nil {
		select {
		case <-c.active.done:
			if err := c.stop(); err != nil {
				return err
			}
		default:
		}
	}
	if fields[0] == "load-stop" {
		if len(fields) != 1 {
			return c.reject("invalid_parameters", errors.New("load-stop takes no parameters"))
		}
		if c.active == nil {
			return c.reject("no_active_load", errors.New("no output workload is active"))
		}
		return c.stop()
	}
	spec, err := parseLoad(fields)
	if err != nil {
		return c.reject("invalid_parameters", err)
	}
	if c.active != nil {
		return c.reject("load_active", errors.New("an output workload is already active"))
	}
	c.nextID++
	startedAt := time.Now()
	started := loadEvent{
		Event: "load-start", LoadID: c.nextID, Status: "running",
		LineBytes: spec.lineBytes, LinesPerSecond: spec.rate, DurationMS: spec.durationMS,
		StartedUTC: startedAt.UTC(),
	}
	if err := c.sink.emit(started); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	job := &loadJob{cancel: cancel, done: make(chan struct{})}
	c.active = job
	go func() {
		defer close(job.done)
		defer cancel()
		ticker := time.NewTicker(time.Second / time.Duration(spec.rate))
		deadline := time.NewTimer(time.Until(startedAt.Add(time.Duration(spec.durationMS) * time.Millisecond)))
		defer ticker.Stop()
		defer deadline.Stop()
		result := pumpLoad(ctx, c.sink, spec, ticker.C, deadline.C)
		finished := started
		finished.Event, finished.Status = "load-finished", result.status
		finished.LinesWritten, finished.BytesWritten = result.lines, result.bytes
		now := time.Now().UTC()
		finished.FinishedUTC = &now
		if result.err != nil {
			finished.Error = result.err.Error()
		}
		job.err = errors.Join(result.err, c.sink.emit(finished))
	}()
	return nil
}

// stop joins the sole output worker. A finished job is reaped without creating
// another event, so exit cannot race a late load-finished record or report close.
func (c *loadController) stop() error {
	if c.active == nil {
		return nil
	}
	job := c.active
	job.cancel()
	<-job.done
	c.active = nil
	return job.err
}

type loadResult struct {
	status       string
	lines, bytes int64
	err          error
}

// Slow writes reduce the achieved rate; missed ticks are not queued or replayed.
// This bounded loop allocates one payload buffer, independently of line count.
// The channels allow tests to verify pacing, cancellation and accounting without
// relying on scheduler timing or inventing a throughput acceptance threshold.
func pumpLoad(ctx context.Context, stdout io.Writer, spec loadSpec, ticks, deadline <-chan time.Time) loadResult {
	line := make([]byte, spec.lineBytes)
	for i := range line {
		line[i] = 'x'
	}
	copy(line, "ASD-LOAD ")
	line[len(line)-1] = '\n'
	result := loadResult{status: "completed"}
	for {
		// Give already-observed stop/deadline signals priority over another tick.
		select {
		case <-ctx.Done():
			result.status = "cancelled"
			return result
		default:
		}
		select {
		case <-deadline:
			return result
		default:
		}
		select {
		case <-ctx.Done():
			result.status = "cancelled"
			return result
		case <-deadline:
			return result
		case <-ticks:
			n, err := writeChecked(stdout, line)
			result.bytes += int64(n)
			if n == len(line) {
				result.lines++
			}
			if err != nil {
				result.status, result.err = "error", err
				return result
			}
		}
	}
}
