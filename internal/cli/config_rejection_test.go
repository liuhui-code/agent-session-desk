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

func TestInvalidConfigurationFailsBeforeBackendInitialization(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		args       []string
	}{
		{"future schema", "schema_version=2", []string{"--template", "dev"}},
		{"unknown field", "schema_version=1\n[templates.dev]\nexecutable='agent'\nargz='private value'", []string{"--template", "dev"}},
		{"missing selection", "schema_version=1", nil},
		{"missing template", "schema_version=1", []string{"--template", "absent"}},
		{"empty selection", "schema_version=1", []string{"--template", ""}},
		{"mixed executable", "schema_version=1", []string{"--template", "dev", "--", "agent"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.toml")
			runtimeDir := filepath.Join(dir, "runtime")
			if err := os.WriteFile(path, []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"--runtime-dir", runtimeDir, "--backend", filepath.Join(dir, "absent-backend"), "--config", path, "start"}
			var out, stderr bytes.Buffer
			if cli.Run(context.Background(), append(args, tc.args...), strings.NewReader(""), &out, &stderr) == 0 {
				t.Fatal("invalid configuration accepted")
			}
			if _, err := os.Stat(runtimeDir); !os.IsNotExist(err) {
				t.Fatalf("backend initialized: %v", err)
			}
			if strings.Contains(stderr.String(), "private value") {
				t.Fatal("configuration values leaked")
			}
		})
	}
	var out, stderr bytes.Buffer
	if cli.Run(context.Background(), []string{"start", "--template", "dev"}, strings.NewReader(""), &out, &stderr) == 0 || !strings.Contains(stderr.String(), "requires --config") {
		t.Fatalf("implicit config selection: %s", &stderr)
	}
}
