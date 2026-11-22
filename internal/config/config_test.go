package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestLoadNamedTemplatePreservesLaunchData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agents.toml")
	text := `schema_version = 1
[templates.review]
name = "评审"
executable = "agent.exe"
args = ["", "a b", "中文", 'a"b', 'end\', '$(x);&', "--help"]
cwd = "project"
[templates.review.env]
ASD_TEST = "中文 value=a=b"
`
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	launch, err := config.LoadLaunch(path, "review")
	if err != nil {
		t.Fatal(err)
	}
	wantArgs := []string{"", "a b", "中文", `a"b`, `end\`, "$(x);&", "--help"}
	if launch.Name != "评审" || launch.Executable != "agent.exe" || !reflect.DeepEqual(launch.Args, wantArgs) || launch.Dir != filepath.Join(dir, "project") || !reflect.DeepEqual(launch.Env, []string{"ASD_TEST=中文 value=a=b"}) {
		t.Fatalf("launch data changed: %#v", launch)
	}
}

func TestUnsupportedSchemaDoesNotProduceLaunch(t *testing.T) {
	for _, input := range []string{"", "[templates.review]\nexecutable='agent'\n", "schema_version=2\n[templates.review]\nexecutable='agent'\n"} {
		path := filepath.Join(t.TempDir(), "agents.toml")
		if err := os.WriteFile(path, []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := config.LoadLaunch(path, "review"); err == nil || !strings.Contains(err.Error(), "schema_version") {
			t.Fatalf("unsupported schema accepted: %v", err)
		}
	}
}

func TestUnknownFieldsAreRejectedWithoutEchoingInput(t *testing.T) {
	const secret = "synthetic-sensitive-marker"
	for _, key := range []string{"unknown", "templates.review.argz"} {
		path := filepath.Join(t.TempDir(), "agents.toml")
		input := "schema_version=1\n" + key + "='" + secret + "'\n[templates.review]\nexecutable='agent'\n"
		if key != "unknown" {
			input = "schema_version=1\n[templates.review]\nexecutable='agent'\nargz='" + secret + "'\n"
		}
		if err := os.WriteFile(path, []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := config.LoadLaunch(path, "review")
		if err == nil || strings.Contains(err.Error(), secret) {
			t.Fatalf("unknown field was accepted or input leaked: %v", err)
		}
	}
}

func TestMalformedConfigDoesNotExposeSourceOrValues(t *testing.T) {
	for _, body := range []string{
		"schema_version=1\n\"synthetic-sensitive-marker\" = [}\n",
		"schema_version=1\n[templates.review]\nexecutable='agent'\nargs=['synthetic-sensitive-marker'\n",
		"schema_version=1\n[templates.review]\nexecutable='agent'\nargs='synthetic-sensitive-marker'\n",
		"schema_version=1\n[templates.review]\nexecutable='agent'\nexecutable='synthetic-sensitive-marker'\n",
	} {
		path := filepath.Join(t.TempDir(), "bad.toml")
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := config.LoadLaunch(path, "review")
		if err == nil || strings.Contains(err.Error(), "synthetic-sensitive-marker") {
			t.Fatalf("unsafe parse error: %v", err)
		}
	}
}

func TestMissingTemplateAndInvalidLaunchAreRejected(t *testing.T) {
	for _, tc := range []struct{ body, id string }{
		{"schema_version=1\n", "missing"},
		{"schema_version=1\n[templates.dev]\nexecutable='agent'\n", ""},
		{"schema_version=1\n[templates.dev]\nargs=['x']\n", "dev"},
		{"schema_version=1\n[templates.dev]\nexecutable='agent'\nargs=[\"\\u0000\"]\n", "dev"},
		{"schema_version=1\n[templates.dev]\nexecutable='agent'\n[templates.dev.env]\n'A=B'='value'\n", "dev"},
	} {
		path := filepath.Join(t.TempDir(), "config.toml")
		if err := os.WriteFile(path, []byte(tc.body), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := config.LoadLaunch(path, tc.id); err == nil {
			t.Fatal("invalid selection/launch accepted")
		}
	}
}

func TestSchemaKeysAreCaseSensitive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("schema_version=1\n[templates.dev]\nExecutable='agent'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.LoadLaunch(path, "dev"); err == nil {
		t.Fatal("case-misspelled key accepted")
	}
}

func TestConfigurationReadHasABound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	base := "schema_version=1\n[templates.dev]\nexecutable='agent'\n#"
	if err := os.WriteFile(path, []byte(base+strings.Repeat("x", (1<<20)-len(base))), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.LoadLaunch(path, "dev"); err != nil {
		t.Fatalf("at limit: %v", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("x")
	f.Close()
	if _, err := config.LoadLaunch(path, "dev"); err == nil {
		t.Fatal("oversized configuration accepted")
	}
}

func TestTemplateHasNoInterpolationOrSharedLaunchData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	body := "schema_version=1\n[templates.dev]\nexecutable='${AGENT}'\nargs=['literal $HOME', '']\n[templates.dev.env]\nZ='z'\nA='a'\n"
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	a, err := config.LoadLaunch(path, "dev")
	if err != nil {
		t.Fatal(err)
	}
	a.Args[0] = "changed"
	a.Env[0] = "changed"
	b, err := config.LoadLaunch(path, "dev")
	if err != nil {
		t.Fatal(err)
	}
	if b.Executable != "${AGENT}" || b.Args[0] != "literal $HOME" || b.Dir != filepath.Dir(path) || !reflect.DeepEqual(b.Env, []string{"A=a", "Z=z"}) {
		t.Fatalf("unexpected expansion or shared data: %#v", b)
	}
}
