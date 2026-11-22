package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/cli"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestMetadataCommandsBackupCheckAndRestoreWithoutBackend(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	state := filepath.Join(dir, "source")
	backup := filepath.Join(dir, "backup 中文 #%.sqlite3")
	restored := filepath.Join(dir, "restored")
	db, err := store.Open(ctx, filepath.Join(state, "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	v := session.Session{ID: "11111111111111111111111111111111", RunID: "22222222222222222222222222222222", BackendID: "asd_test", Title: "工作", Command: "agent.exe"}
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	db.Close()
	missing := filepath.Join(dir, "absent-backend.exe")
	for _, tc := range []struct {
		state string
		args  []string
	}{
		{state, []string{"metadata", "check"}},
		{state, []string{"metadata", "backup", backup}},
		{restored, []string{"metadata", "restore", backup}},
		{restored, []string{"metadata", "check"}},
	} {
		var out, stderr bytes.Buffer
		args := append([]string{"--backend", missing, "--state-dir", tc.state}, tc.args...)
		if code := cli.Run(ctx, args, nil, &out, &stderr); code != 0 {
			t.Fatalf("%v: %s", tc.args, &stderr)
		}
		var result map[string]any
		if err = json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatalf("response: %s %v", &out, err)
		}
	}
	db, err = store.Open(ctx, filepath.Join(restored, "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.History(ctx, 100)
	if err != nil || len(rows) != 1 || rows[0].Session.RunID != v.RunID || rows[0].Session.TerminalState != session.TerminalUnknown {
		t.Fatalf("restored evidence: %#v %v", rows, err)
	}
	if _, err = os.Stat(missing); !os.IsNotExist(err) {
		t.Fatal("metadata command created backend")
	}
}
