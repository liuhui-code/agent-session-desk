package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/cli"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestSavedOutputCanBeReadOfflineOnlyInItsOriginalScope(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	state := filepath.Join(dir, "state")
	runtimeDir := filepath.Join(dir, "runtime")
	kind, location, err := backend.Location(backend.Options{RuntimeDir: runtimeDir, Namespace: "asd-output"})
	if err != nil {
		t.Fatal(err)
	}
	scope := kind + ":" + location
	if runtime.GOOS == "windows" {
		scope = strings.ToLower(scope)
	}
	db, err := store.Open(ctx, filepath.Join(state, "metadata.sqlite3"), scope)
	if err != nil {
		t.Fatal(err)
	}
	v := session.Session{ID: strings.Repeat("1", 32), RunID: strings.Repeat("2", 32), BackendID: "asd_original", Title: "任务", Command: "custom.exe"}
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	capture, err := db.RecordOutput(ctx, v, session.Preview{Text: "合成输出\x1b]52;c;SYNTHETIC\a"})
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	for _, ns := range []string{"asd-output", "asd-other"} {
		var out, stderr bytes.Buffer
		code := cli.Run(ctx, []string{"--backend", filepath.Join(dir, "absent-backend"), "--runtime-dir", runtimeDir, "--namespace", ns, "--state-dir", state, "output", v.RunID}, nil, &out, &stderr)
		if ns == "asd-other" {
			if code == 0 || strings.Contains(out.String(), capture.ID) {
				t.Fatal("output escaped its original scope")
			}
			continue
		}
		if code != 0 {
			t.Fatalf("offline output required backend: %s", &stderr)
		}
		var rows []session.OutputRecord
		if err = json.Unmarshal(out.Bytes(), &rows); err != nil || len(rows) != 1 || rows[0].ID != capture.ID || rows[0].Text != capture.Text || rows[0].Complete {
			t.Fatalf("offline record: %s %v", &out, err)
		}
		if strings.ContainsRune(out.String(), '\x1b') {
			t.Fatal("CLI replayed terminal controls instead of JSON data")
		}
	}
}

func TestRecordValidatesFullSessionTargetBeforeTouchingRuntimeFiles(t *testing.T) {
	dir := t.TempDir()
	state := filepath.Join(dir, "state")
	runtimeDir := filepath.Join(dir, "runtime")
	for _, tail := range [][]string{{"record"}, {"record", "0"}, {"record", strings.Repeat("a", 32), "extra"}} {
		var out, stderr bytes.Buffer
		args := append([]string{"--backend", filepath.Join(dir, "missing"), "--state-dir", state, "--runtime-dir", runtimeDir}, tail...)
		if code := cli.Run(context.Background(), args, nil, &out, &stderr); code == 0 || !strings.Contains(stderr.String(), "full session_id") {
			t.Fatalf("invalid record target: %d %s", code, &stderr)
		}
	}
	for _, path := range []string{state, runtimeDir} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("invalid capture request created runtime/state files")
		}
	}
}
