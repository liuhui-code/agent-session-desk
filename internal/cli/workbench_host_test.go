package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowTokenRejectsWrongRoleBeforeStateOrBackendInitialization(t *testing.T) {
	token := strings.Repeat("a", 32)
	for _, tail := range [][]string{
		{"--window-token", token},
		{"--inside-terminal", "--window-token", token, "list"},
		{"--inside-terminal", "--window-token", "not-a-launch-token"},
		{"--inside-terminal", "--window-token", strings.Repeat("a", 31) + ";"},
	} {
		root := t.TempDir()
		state := filepath.Join(root, "state")
		args := append([]string{"--state-dir", state, "--backend", filepath.Join(root, "missing.exe")}, tail...)
		var out, errout bytes.Buffer
		if Run(context.Background(), args, strings.NewReader(""), &out, &errout) == 0 {
			t.Fatal("invalid internal handoff accepted")
		}
		if _, err := os.Stat(state); !os.IsNotExist(err) {
			t.Fatal("invalid handoff initialized persistent state")
		}
	}
}
