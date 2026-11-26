//go:build windows

package manager_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/cli"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

// Real native psmux + ordinary controlled CLI, with an independent input journal.
// It is not a model-agent, readiness protocol, desktop or long-soak test.
func TestWindowsNativeSimulatedAgentMessageContract(t *testing.T) {
	bundle, fixture := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_MESSAGING_FIXTURE")
	if bundle == "" || fixture == "" {
		t.Skip("explicit pinned bundle and fixture required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := t.TempDir()
	id, err := session.NewID()
	if err != nil {
		t.Fatal(err)
	}
	opts := backend.Options{TerminalBundle: bundle, RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-sim-" + id[:12]}
	b, err := backend.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	kind, location, err := backend.Location(opts)
	if err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(ctx, filepath.Join(root, "state", "metadata.sqlite3"), strings.ToLower(kind+":"+location))
	if err != nil {
		t.Fatal(err)
	}
	m := manager.New(b, db)
	globals := []string{"--terminal-bundle", bundle, "--runtime-dir", opts.RuntimeDir, "--namespace", opts.Namespace, "--state-dir", filepath.Join(root, "state")}
	call := func(args ...string) (session.Message, int) {
		var out, stderr bytes.Buffer
		code := cli.Run(ctx, append(append([]string(nil), globals...), args...), nil, &out, &stderr)
		var q session.Message
		if out.Len() > 0 {
			if err := json.Unmarshal(out.Bytes(), &q); err != nil {
				t.Errorf("invalid machine-readable message response: %v", err)
			}
		}
		return q, code
	}
	var originals []session.Session
	stopped := map[string]bool{}
	defer func() {
		for _, v := range originals {
			if stopped[v.RunID] {
				continue
			}
			c, stop := context.WithTimeout(context.Background(), 10*time.Second)
			if err := m.Stop(c, v); err != nil {
				t.Errorf("owned fixture cleanup: %v", err)
			}
			stop()
		}
		db.Close()
	}()
	start := func(name string) (session.Session, string) {
		journal := filepath.Join(root, name+".jsonl")
		v, err := m.Start(ctx, session.Launch{Name: name, Executable: fixture, Dir: root, Env: []string{"ASD_FIXTURE_REPORT_PATH=" + journal}})
		if err != nil {
			t.Fatal(err)
		}
		originals = append(originals, v)
		nativeWaitFor(t, ctx, func() bool { data, _ := os.ReadFile(journal); return bytes.Contains(data, []byte(`"event":"started"`)) })
		return v, journal
	}
	a, sourceLog := start("Source")
	bTarget, targetLog := start("Target")
	sendFile := func(target session.Session, body string, deliver bool) (session.Message, int, []string) {
		requestID, e := session.NewID()
		if e != nil {
			t.Fatal(e)
		}
		file := filepath.Join(root, requestID+".txt")
		if e = os.WriteFile(file, append([]byte{0xef, 0xbb, 0xbf}, []byte(body)...), 0600); e != nil {
			t.Fatal(e)
		}
		args := []string{"send", "--request-id", requestID, "--to", target.ID, "--run", target.RunID, "--body-file", file}
		if deliver {
			args = append(args, "--deliver")
		}
		q, code := call(args...)
		return q, code, args
	}
	sourceText := "SOURCE_" + id[:8] + "_中文"
	if q, code, _ := sendFile(a, sourceText+"\n", true); code != 0 || q.State != "input_written" {
		t.Fatal("source seed delivery failed")
	}
	nativeWaitFor(t, ctx, func() bool { lines := nativeInputLines(t, sourceLog); return len(lines) == 1 && lines[0] == sourceText })
	var record session.OutputRecord
	sourceLine := 0
	nativeWaitFor(t, ctx, func() bool {
		var out, stderr bytes.Buffer
		code := cli.Run(ctx, append(append([]string(nil), globals...), "read", "--lines", "60", a.ID), nil, &out, &stderr)
		if code != 0 {
			t.Fatal("native source read failed")
		}
		if err := json.Unmarshal(out.Bytes(), &record); err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(record.Text, "\n") {
			var event struct{ Event, Text string }
			if json.Unmarshal([]byte(line), &event) == nil && event.Event == "input" && event.Text == sourceText {
				sourceLine = i + 1
				return true
			}
		}
		return false
	})
	requestID, _ := session.NewID()
	sourceArgs := []string{"send", "--request-id", requestID, "--to", bTarget.ID, "--run", bTarget.RunID, "--source-run", a.RunID, "--record", record.ID, "--start-line", fmt.Sprint(sourceLine), "--end-line", fmt.Sprint(sourceLine)}
	queued, code := call(sourceArgs...)
	if code != 0 || queued.State != "queued" || len(nativeInputLines(t, targetLog)) != 0 {
		t.Fatal("creation wrote input or failed")
	}
	written, code := call("deliver", requestID)
	if code != 0 || written.State != "input_written" || written.Source == nil || written.Source.Complete {
		t.Fatal("snapshot delivery contract failed")
	}
	// Deliberately submit the fixture line with a separate explicit LF request.
	// Product delivery itself must not append a submission key.
	if q, code, _ := sendFile(bTarget, "\n", true); code != 0 || q.State != "input_written" {
		t.Fatal("explicit fixture submission failed")
	}
	nativeWaitFor(t, ctx, func() bool {
		lines := nativeInputLines(t, targetLog)
		return len(lines) == 1 && lines[0] == queued.Body
	})
	literal := "quotes ' \" tail\\ 中文🙂\nsecond é $() ;\n"
	q, code, retryArgs := sendFile(bTarget, literal, true)
	if code != 0 || q.State != "input_written" || q.Body != literal {
		t.Fatal("literal file delivery failed")
	}
	again, code := call(retryArgs...)
	if code != 0 || again.ID != q.ID || again.State != "input_written" {
		t.Fatal("idempotent CLI retry changed receipt")
	}
	nativeWaitFor(t, ctx, func() bool {
		lines := nativeInputLines(t, targetLog)
		return len(lines) == 3 && strings.Join(lines[1:], "\n")+"\n" == literal
	})
	// Two public CLI consumers race one queued request; only one can write it.
	q, code, _ = sendFile(bTarget, "RACE_"+id[:8]+"\n", false)
	if code != 0 {
		t.Fatal("race request creation failed")
	}
	startRace := make(chan struct{})
	results := make(chan int, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); <-startRace; _, code := call("deliver", q.ID); results <- code }()
	}
	close(startRace)
	wg.Wait()
	close(results)
	wins := 0
	for code := range results {
		if code == 0 {
			wins++
		}
	}
	if wins != 1 {
		t.Fatalf("concurrent CLI winners=%d", wins)
	}
	nativeWaitFor(t, ctx, func() bool {
		lines := nativeInputLines(t, targetLog)
		return len(lines) == 4 && lines[3] == "RACE_"+id[:8]
	})
	pending, code, _ := sendFile(bTarget, "DO_NOT_REPLAY\n", false)
	if code != 0 {
		t.Fatal("old-run request creation failed")
	}
	if err = m.Stop(ctx, bTarget); err != nil {
		t.Fatal(err)
	}
	stopped[bTarget.RunID] = true
	if _, code = call("deliver", pending.ID); code == 0 {
		t.Fatal("stopped Run accepted delivery")
	}
	status, code := call("status", pending.ID)
	if code != 0 || status.State != "queued" || status.Target.RunID != bTarget.RunID {
		t.Fatal("stopped-target failure lost durable request")
	}
	if lines := nativeInputLines(t, targetLog); len(lines) != 4 {
		t.Fatal("duplicate or stopped-target input observed")
	}
	// Existing queued/written records remain queryable without constructing a UI.
	t.Log("native controlled CLI: snapshot -> copied input; literal Unicode/multiline; idempotent retry; two CLI consumers/one write; stopped Run rejected; 4 target input lines")
}

func nativeWaitFor(t *testing.T, ctx context.Context, check func() bool) {
	t.Helper()
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		if check() {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("native simulation context expired")
		case <-timer.C:
			t.Fatal("independent native observation deadline")
		case <-ticker.C:
		}
	}
}
func nativeInputLines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("fixture observer unavailable")
	}
	var lines []string
	for _, raw := range bytes.Split(data, []byte{'\n'}) {
		if len(raw) == 0 {
			continue
		}
		var event struct{ Event, Text string }
		if json.Unmarshal(raw, &event) != nil {
			continue
		}
		if event.Event == "input" {
			lines = append(lines, event.Text)
		}
	}
	return lines
}
