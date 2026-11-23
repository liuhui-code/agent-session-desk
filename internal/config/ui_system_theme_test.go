package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestUISystemThemeV6UpgradesWithoutResolvingPreference(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ui.toml")
	text := "schema_version=6\nsidebar_density='minimal'\nsidebar_mode='flat'\nsidebar_percent=45\ntheme='system'\n"
	if e := os.WriteFile(p, []byte(text), 0600); e != nil {
		t.Fatal(e)
	}
	s, e := config.LoadUI(p)
	if e != nil || s.Theme != "system" {
		t.Fatalf("System preference unavailable: %+v %v", s, e)
	}
	if e = config.SaveUI(p, s); e != nil {
		t.Fatal(e)
	}
	after, e := config.LoadUI(p)
	if e != nil || after != s {
		t.Fatalf("System preference was resolved or lost: %+v %v", after, e)
	}
	b, e := os.ReadFile(p)
	if e != nil || !strings.Contains(string(b), "schema_version = 7") || !strings.Contains(string(b), `theme = "system"`) {
		t.Fatal("System save did not retain v6 choice")
	}
}
