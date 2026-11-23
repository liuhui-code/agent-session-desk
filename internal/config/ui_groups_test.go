package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestUIV3GroupFoldsPersistWithLegacyReadOnlyLoading(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.toml")
	legacy := "schema_version=2\nsidebar_density='full'\nsidebar_mode='grouped'\ndefault_group_collapsed=true\n"
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := config.LoadUI(path)
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != legacy {
		t.Fatal("load rewrote v2")
	}
	s.CollapsedGroups = strings.Repeat("b", 32) + "," + strings.Repeat("a", 32)
	if err = config.SaveUI(path, s); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.LoadUI(path)
	if err != nil || loaded.CollapsedGroups != strings.Repeat("a", 32)+","+strings.Repeat("b", 32) || !loaded.DefaultGroupCollapsed || loaded.SidebarDensity != "full" {
		t.Fatalf("lost folds/density: %+v %v", loaded, err)
	}
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "schema_version = 5") {
		t.Fatalf("not current v4: %s %v", data, err)
	}
	before := string(data)
	s.CollapsedGroups = "../bad"
	if err = config.SaveUI(path, s); err == nil {
		t.Fatal("non-identity fold accepted")
	}
	after, err = os.ReadFile(path)
	if err != nil || string(after) != before {
		t.Fatal("invalid save replaced valid prefs")
	}
}
