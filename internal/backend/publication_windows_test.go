//go:build windows

package backend

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"golang.org/x/sys/windows"
)

func TestWindowsNativeDiscoveryBeforeAndAfterPublication(t *testing.T) {
	bundle, fixture := os.Getenv("ASD_TERMINAL_BUNDLE_TEST_DIR"), os.Getenv("ASD_FIXTURE_TEST_BINARY")
	if bundle == "" || fixture == "" {
		t.Skip("explicit pinned native bundle and fixture required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	id, _ := session.NewID()
	runID, _ := session.NewID()
	root := t.TempDir()
	b, err := New(Options{TerminalBundle: bundle, RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-publish-" + id[:12]})
	if err != nil {
		t.Fatal(err)
	}
	name := "asd_" + id + "_" + runID
	report := filepath.Join(root, "fixture.jsonl")
	l := session.Launch{Name: "Publication fixture", Executable: fixture, Dir: root, Env: []string{"ASD_FIXTURE_REPORT_PATH=" + report}}
	// This is StartRun's real construction window: process exists, marker absent.
	if _, err = b.run(ctx, maxControlBytes, psmuxStartArgs(name, l)...); err != nil {
		t.Fatal(b.failedStart(name, err))
	}
	var process windows.Handle
	t.Cleanup(func() {
		c, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if _, err := b.run(c, maxControlBytes, "kill-session", "-t", name); err != nil {
			t.Error("owned unpublished session cleanup", err)
		}
		if process != 0 {
			if state, err := windows.WaitForSingleObject(process, 10000); err != nil || state != windows.WAIT_OBJECT_0 {
				t.Error("original fixture did not end", state, err)
			}
			windows.CloseHandle(process)
		}
	})
	event := fixtureEvent(t, ctx, report, "started", "")
	process, err = windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(event.PID))
	if err != nil {
		t.Fatal(err)
	}
	rows, err := b.List(ctx)
	if err != nil || len(rows) != 0 {
		t.Fatalf("native unpublished session misclassified: %+v %v", rows, err)
	}
	pane, err := b.run(ctx, maxControlBytes, "list-panes", "-t", name, "-F", "#{pane_id}")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := encodeRecord(liveRecord{1, id, runID, l.Name, fixture, strings.TrimSpace(pane.text)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = b.run(ctx, maxControlBytes, b.metadataArgs(name, encoded)...); err != nil {
		t.Fatal(err)
	}
	rows, err = b.List(ctx)
	if err != nil || len(rows) != 1 || rows[0].ID != id || rows[0].RunID != runID {
		t.Fatalf("native published identity missing: %+v %v", rows, err)
	}
	if state, err := windows.WaitForSingleObject(process, 0); err != nil || state != uint32(windows.WAIT_TIMEOUT) {
		t.Fatal("publication changed original fixture lifetime", state, err)
	}
}
