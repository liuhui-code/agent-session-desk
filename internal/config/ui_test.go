package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestUIPreferencesPersistAcrossLoadsAndReplaceExistingChoice(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.toml")
	initial, err := config.LoadUI(path)
	if err != nil || initial.SidebarDensity != "compact" {
		t.Fatalf("missing preference must use compact: %+v %v", initial, err)
	}
	for _, density := range []string{"full", "minimal", "auto", "compact"} {
		if err := config.SaveUI(path, config.UISettings{SidebarDensity: density}); err != nil {
			t.Fatal(err)
		}
		got, err := config.LoadUI(path)
		if err != nil || got.SidebarDensity != density {
			t.Fatalf("saved choice not restored: %+v %v", got, err)
		}
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary preference files remain: %v %v", entries, err)
	}
}

func TestUIPreferencesRefuseToOverwriteUnrecognizedOrInvalidFiles(t *testing.T) {
	for _, contents := range []string{
		"schema_version=7\nsidebar_density='full'\n",
		"schema_version=1\nother='PRIVATE-SYNTHETIC-VALUE'\n",
		"schema_version=1\nsidebar_density=42\n",
		"schema_version=1\nsidebar_density='unterminated",
		strings.Repeat("x", 4097),
	} {
		path := filepath.Join(t.TempDir(), "ui.toml")
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
		err := config.SaveUI(path, config.UISettings{SidebarDensity: "minimal"})
		if err == nil || strings.Contains(err.Error(), "PRIVATE-SYNTHETIC-VALUE") {
			t.Fatalf("invalid preference accepted or content leaked: %v", err)
		}
		got, e := os.ReadFile(path)
		if e != nil || string(got) != contents {
			t.Fatal("existing preference changed on refusal")
		}
	}
}

func TestUIPreferencesNormalizeUpstreamDensityValues(t *testing.T) {
	for _, value := range []string{"AUTO", " compact ", "unrecognized", ""} {
		path := filepath.Join(t.TempDir(), "ui.toml")
		contents := fmt.Sprintf("schema_version=1\nsidebar_density=%q\n", value)
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := config.LoadUI(path)
		want := "compact"
		if value == "AUTO" {
			want = "auto"
		}
		if err != nil || got.SidebarDensity != want {
			t.Fatalf("%q: %+v %v", value, got, err)
		}
	}
}

func TestUIPreferencesRejectCaseVariantSchemaKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui.toml")
	contents := "Schema_Version=1\nSidebar_Density='full'\n"
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.LoadUI(path); err == nil {
		t.Fatal("case-variant keys were accepted as the current schema")
	}
}
