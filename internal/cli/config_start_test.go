package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/cli"
)

// The observer is a real subprocess reached through the public CLI. It records
// argv before rejecting creation; it is not a terminal/ConPTY substitute.
func TestMain(m *testing.M) {
	if path := os.Getenv("ASD_CONFIG_OBSERVER"); path != "" {
		args := os.Args[1:]
		if slices.Contains(args, "-V") {
			if runtime.GOOS == "windows" {
				fmt.Println("psmux 3.4.0 (20aa1d7 2026-10-13)")
			} else {
				fmt.Println("tmux 3.8")
			}
			os.Exit(0)
		}
		if slices.Contains(args, "new-session") {
			data, _ := json.Marshal(args)
			if os.WriteFile(path, data, 0600) != nil {
				os.Exit(72)
			}
			os.Exit(71)
		}
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestCLIStartsExplicitNamedTemplate(t *testing.T) {
	testTemplate(t, false)
}

func TestCLIOverridesTemplateWorkingDirectoryAndEnvironment(t *testing.T) {
	testTemplate(t, true)
}

func testTemplate(t *testing.T, override bool) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	cwd := filepath.Join(dir, "project")
	if err := os.Mkdir(cwd, 0700); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, "agents.toml")
	// JSON's quoted strings are valid TOML basic strings for these path samples.
	quoted, _ := json.Marshal(exe)
	body := "schema_version=1\n[templates.review]\nname='评审'\nexecutable=" + string(quoted) + "\ncwd='project'\nargs=['', 'a b', '中文', 'end\\']\n[templates.review.env]\nASD_TEST='template value'\n"
	if err := os.WriteFile(cfg, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	trace := filepath.Join(dir, "trace.json")
	t.Setenv("ASD_CONFIG_OBSERVER", trace)
	var out, stderr bytes.Buffer
	argsIn := []string{"--backend", exe, "--runtime-dir", filepath.Join(dir, "runtime"), "--namespace", "asd-config-test", "--config", cfg, "start", "--template", "review"}
	wantEnv := "ASD_TEST=template value"
	if override {
		cwd = dir
		wantEnv = "ASD_TEST=CLI override"
		argsIn = append(argsIn, "--cwd", dir, "--name", "覆盖", "--env", wantEnv)
	}
	code := cli.Run(context.Background(), argsIn, strings.NewReader(""), &out, &stderr)
	if code == 0 {
		t.Fatal("observer should reject creation after recording argv")
	}
	data, err := os.ReadFile(trace)
	if err != nil {
		t.Fatalf("template did not reach backend: %v; %s", err, stderr.String())
	}
	var args []string
	if err := json.Unmarshal(data, &args); err != nil {
		t.Fatal(err)
	}
	want := []string{exe, "", "a b", "中文", `end\`}
	if len(args) < len(want) || !reflect.DeepEqual(args[len(args)-len(want):], want) {
		t.Fatalf("argument boundaries changed: %#v", args)
	}
	if p := slices.Index(args, "-c"); p < 0 || args[p+1] != cwd {
		t.Fatalf("cwd changed: %#v", args)
	}
	if !slices.Contains(args, wantEnv) || (override && slices.Contains(args, "ASD_TEST=template value")) {
		t.Fatalf("template env missing: %#v", args)
	}
}
