package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestUINavigationV2RoundTripsWithoutLosingDensityOrFoldChoice(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.toml")
	text := "schema_version = 2\nsidebar_density = 'minimal'\nsidebar_mode = 'flat'\ndefault_group_collapsed = true\n"
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	settings, err := config.LoadUI(path)
	if err != nil {
		t.Fatalf("valid navigation preferences refused: %v", err)
	}
	if err = config.SaveUI(path, settings); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"schema_version = 6", `sidebar_density = "minimal"`, `sidebar_mode = "flat"`, `default_group_collapsed = true`} {
		if !strings.Contains(string(data), field) {
			t.Errorf("round-trip lost %s: %s", field, data)
		}
	}
}

func TestUILegacyDensityReadDoesNotRewriteAndUsesGroupedExpandedDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.toml")
	text := "schema_version=1\nsidebar_density='full'\n"
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := config.LoadUI(path)
	if err != nil || s.SidebarDensity != "full" || s.SidebarMode != "grouped" || s.DefaultGroupCollapsed {
		t.Fatalf("legacy defaults: %+v %v", s, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != text {
		t.Fatal("read migrated the legacy file")
	}
}

func TestUINavigationRefusesUnknownKeysAndWrongTypesWithoutOverwriting(t *testing.T) {
	for _, tail := range []string{"Sidebar_Mode='flat'", "default_group_collapsed='true'", "default_group_collapsed=1", "other='PRIVATE-SYNTHETIC'", "[navigation]\nsidebar_mode='flat'"} {
		path := filepath.Join(t.TempDir(), "ui.toml")
		text := "schema_version=2\nsidebar_density='full'\n" + tail + "\n"
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if err := config.SaveUI(path, config.UISettings{SidebarDensity: "minimal", SidebarMode: "grouped"}); err == nil || strings.Contains(err.Error(), "PRIVATE-SYNTHETIC") {
			t.Fatalf("bad v2 document accepted or exposed: %v", err)
		}
		after, err := os.ReadFile(path)
		if err != nil || string(after) != text {
			t.Fatal("refusal modified original")
		}
	}
}
