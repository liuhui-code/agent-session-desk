package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestUIThemeV5LoadsAndRoundTripsWithoutLosingOtherPreferences(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ui.toml")
	text := "schema_version=5\nsidebar_density='full'\nsidebar_mode='flat'\nsidebar_percent=35\ndefault_group_collapsed=true\ncollapsed_groups='aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'\ntheme='light'\n"
	if e := os.WriteFile(p, []byte(text), 0600); e != nil {
		t.Fatal(e)
	}
	s, e := config.LoadUI(p)
	if e != nil {
		t.Fatalf("valid light theme unavailable: %v", e)
	}
	if s.Theme != "light" {
		t.Fatal("loaded theme was discarded")
	}
	if e = config.SaveUI(p, s); e != nil {
		t.Fatal(e)
	}
	after, e := config.LoadUI(p)
	if e != nil || after != s {
		t.Fatalf("theme/settings roundtrip: %+v %v", after, e)
	}
	b, e := os.ReadFile(p)
	if e != nil || !strings.Contains(string(b), `theme = "light"`) {
		t.Fatal("save lost theme")
	}
}

func TestUIThemeLegacyLoadsReadOnlyAndUpgradesOnlyOnExplicitSave(t *testing.T) {
	for _, version := range []string{"1", "2", "3", "4"} {
		t.Run(version, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "ui.toml")
			original := "schema_version=" + version + "\nsidebar_density='minimal'\n"
			if e := os.WriteFile(p, []byte(original), 0600); e != nil {
				t.Fatal(e)
			}
			s, e := config.LoadUI(p)
			if e != nil || s.Theme != "" {
				t.Fatalf("legacy default: %+v %v", s, e)
			}
			b, e := os.ReadFile(p)
			if e != nil || string(b) != original {
				t.Fatal("loading rewrote old preference file")
			}
			s.Theme = "light"
			if e = config.SaveUI(p, s); e != nil {
				t.Fatal(e)
			}
			b, e = os.ReadFile(p)
			if e != nil || !strings.Contains(string(b), "schema_version = 5") || !strings.Contains(string(b), `theme = "light"`) {
				t.Fatal("explicit save did not upgrade theme")
			}
		})
	}
}

func TestUIThemeRejectsInvalidFilesWithoutReplacingOrExposingContent(t *testing.T) {
	for _, text := range []string{"schema_version=1\ntheme='light'", "schema_version=2\ntheme='light'", "schema_version=3\ntheme='light'", "schema_version=4\ntheme='light'", "schema_version=6\ntheme='light'", "schema_version=5\ntheme=''", "schema_version=5\ntheme='system'", "schema_version=5\ntheme='LIGHT'", "schema_version=5\ntheme=true", "schema_version=5\ntheme=['light']", "schema_version=5\ntheme='PRIVATE-SYNTHETIC'"} {
		t.Run(text, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "ui.toml")
			if e := os.WriteFile(p, []byte(text), 0600); e != nil {
				t.Fatal(e)
			}
			_, e := config.LoadUI(p)
			if e == nil || strings.Contains(e.Error(), "PRIVATE-SYNTHETIC") {
				t.Fatal("invalid theme accepted or private content exposed")
			}
			if e = config.SaveUI(p, config.UISettings{SidebarDensity: "compact", Theme: "light"}); e == nil {
				t.Fatal("invalid source replaced")
			}
			b, e := os.ReadFile(p)
			if e != nil || string(b) != text {
				t.Fatal("invalid source modified")
			}
		})
	}
}

func TestUIThemeDefaultAndExplicitDarkRoundTripAndRejectBadWrites(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ui.toml")
	for _, theme := range []string{"", "dark", "light"} {
		s := config.UISettings{SidebarDensity: "compact", SidebarMode: "flat", Theme: theme}
		if e := config.SaveUI(p, s); e != nil {
			t.Fatal(e)
		}
		after, e := config.LoadUI(p)
		if e != nil || after != s {
			t.Fatalf("theme roundtrip: %+v %v", after, e)
		}
	}
	before, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	for _, theme := range []string{"system", "LIGHT", " light", "private-synthetic"} {
		if e := config.SaveUI(p, config.UISettings{SidebarDensity: "compact", Theme: theme}); e == nil {
			t.Fatal("invalid write accepted")
		}
	}
	after, e := os.ReadFile(p)
	if e != nil || string(after) != string(before) {
		t.Fatal("invalid write altered saved theme")
	}
	entries, e := os.ReadDir(filepath.Dir(p))
	if e != nil || len(entries) != 1 {
		t.Fatal("theme write left temporary files")
	}
}
