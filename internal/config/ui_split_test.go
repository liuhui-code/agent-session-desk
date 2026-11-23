package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestUISplitV4RoundTripsManualAndAutoWithOtherPreferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.toml")
	for _, pct := range []string{"10", "35", "90", "0"} {
		text := "schema_version=4\nsidebar_density='full'\nsidebar_mode='flat'\ndefault_group_collapsed=true\ncollapsed_groups='aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'\nsidebar_percent=" + pct + "\n"
		if e := os.WriteFile(path, []byte(text), 0600); e != nil {
			t.Fatal(e)
		}
		s, e := config.LoadUI(path)
		if e != nil {
			t.Fatalf("valid split cannot load: %v", e)
		}
		if e = config.SaveUI(path, s); e != nil {
			t.Fatal(e)
		}
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		for _, value := range []string{"schema_version = 4", "sidebar_percent = " + pct, `sidebar_density = "full"`, `sidebar_mode = "flat"`, `default_group_collapsed = true`, `collapsed_groups = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`} {
			if !strings.Contains(string(b), value) {
				t.Fatalf("lost preference %s: %s", value, b)
			}
		}
	}
}

func TestUISplitRejectsInvalidOrVersionMismatchedValuesWithoutReplacing(t *testing.T) {
	for _, entry := range []struct{ version, value string }{{"4", "-1"}, {"4", "1"}, {"4", "9"}, {"4", "91"}, {"4", "100"}, {"4", "1.5"}, {"4", "true"}, {"4", "'40'"}, {"4", "9223372036854775808"}, {"3", "40"}, {"2", "0"}, {"1", "40"}, {"5", "40"}} {
		t.Run(entry.version+"/"+entry.value, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "ui.toml")
			original := "schema_version=" + entry.version + "\nsidebar_density='minimal'\nsidebar_percent=" + entry.value + "\n"
			if e := os.WriteFile(path, []byte(original), 0600); e != nil {
				t.Fatal(e)
			}
			if _, e := config.LoadUI(path); e == nil {
				t.Fatal("invalid split loaded")
			}
			if e := config.SaveUI(path, config.UISettings{SidebarDensity: "full", SidebarPercent: 40}); e == nil {
				t.Fatal("invalid original file replaced")
			}
			b, e := os.ReadFile(path)
			if e != nil || string(b) != original {
				t.Fatal("rejection modified original file")
			}
		})
	}
}

func TestUISplitLegacyLoadRemainsReadOnlyUntilExplicitSave(t *testing.T) {
	for _, version := range []string{"1", "2", "3"} {
		path := filepath.Join(t.TempDir(), "ui.toml")
		original := "schema_version=" + version + "\nsidebar_density='minimal'\n"
		if e := os.WriteFile(path, []byte(original), 0600); e != nil {
			t.Fatal(e)
		}
		s, e := config.LoadUI(path)
		if e != nil || s.SidebarPercent != 0 {
			t.Fatalf("legacy auto default: %+v %v", s, e)
		}
		b, e := os.ReadFile(path)
		if e != nil || string(b) != original {
			t.Fatal("load rewrote legacy")
		}
		s.SidebarPercent = 55
		if e = config.SaveUI(path, s); e != nil {
			t.Fatal(e)
		}
		loaded, e := config.LoadUI(path)
		if e != nil || loaded.SidebarPercent != 55 || loaded.SidebarDensity != "minimal" {
			t.Fatalf("legacy explicit save: %+v %v", loaded, e)
		}
	}
}

func TestUISplitInvalidSavePreservesExistingAndLeavesNoTemporaryFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.toml")
	valid := config.UISettings{SidebarDensity: "minimal", SidebarPercent: 35}
	if e := config.SaveUI(path, valid); e != nil {
		t.Fatal(e)
	}
	before, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	for _, pct := range []int{-1, 1, 9, 91, 100} {
		invalid := valid
		invalid.SidebarPercent = pct
		if e = config.SaveUI(path, invalid); e == nil {
			t.Fatal("invalid save accepted")
		}
		after, e := os.ReadFile(path)
		if e != nil || string(after) != string(before) {
			t.Fatal("invalid save changed valid preferences")
		}
	}
	entries, e := os.ReadDir(filepath.Dir(path))
	if e != nil || len(entries) != 1 {
		t.Fatal("temporary artifacts remained")
	}
}
