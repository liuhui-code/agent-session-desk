package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/cli"
)

func TestTerminalBundleRejectsInvalidSelectionBeforeRuntimeCreation(t *testing.T) {
	dir := t.TempDir()
	runtimeDir := filepath.Join(dir, "runtime")
	var out, stderr bytes.Buffer
	code := cli.Run(context.Background(), []string{"--terminal-bundle", filepath.Join(dir, "missing"), "--runtime-dir", runtimeDir, "doctor"}, nil, &out, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "terminal bundle") || strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("invalid bundle must be rejected explicitly: code=%d stderr=%q", code, stderr.String())
	}
	if _, err := os.Stat(runtimeDir); !os.IsNotExist(err) {
		t.Fatalf("invalid bundle created runtime files: %v", err)
	}
}
