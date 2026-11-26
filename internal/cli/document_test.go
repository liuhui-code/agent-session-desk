package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDocumentAppRejectsUnsupportedUseBeforeState(t *testing.T) {
	state := filepath.Join(t.TempDir(), "state")
	tests := [][]string{{"--state-dir", state, "--document-app", "shell"}, {"--state-dir", state, "--document-app", "system-notepad", "list"}}
	if runtime.GOOS != "windows" {
		tests = append(tests, []string{"--state-dir", state, "--document-app", "system-notepad"})
	}
	for _, args := range tests {
		var out, stderr bytes.Buffer
		if Run(context.Background(), args, nil, &out, &stderr) == 0 {
			t.Fatal("unsupported editor use accepted")
		}
	}
	if _, err := os.Stat(state); !os.IsNotExist(err) {
		t.Fatal("invalid editor option touched state")
	}
}
