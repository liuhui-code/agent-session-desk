package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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
