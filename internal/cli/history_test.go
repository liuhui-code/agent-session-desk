package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/cli"
)

func TestHistoryWorksWithoutBackendAndSeparatesScope(t *testing.T) {
	dir := t.TempDir()
	state := filepath.Join(dir, "state 中文 #%")
	missing := filepath.Join(dir, "missing-backend")
	for _, ns := range []string{"asd-a", "asd-b"} {
		var out, stderr bytes.Buffer
		code := cli.Run(context.Background(), []string{"--backend", missing, "--runtime-dir", filepath.Join(dir, "runtime"), "--state-dir", state, "--namespace", ns, "history"}, nil, &out, &stderr)
		if code != 0 {
			t.Fatalf("history required live backend: %s", &stderr)
		}
		var rows []any
		if err := json.Unmarshal(out.Bytes(), &rows); err != nil || len(rows) != 0 {
			t.Fatalf("history: %s %v", &out, err)
		}
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatal("unexpected backend mutation")
	}
}
