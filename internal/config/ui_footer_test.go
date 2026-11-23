package config_test

import (
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUIFooterV7LoadsAndRoundTripsModeWithoutChangingOtherPreferences(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ui.toml")
	original := "schema_version=7\nsidebar_density='minimal'\nsidebar_mode='flat'\nsidebar_percent=45\ntheme='system'\nfooter='curated'\n"
	if err := os.WriteFile(p, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := config.LoadUI(p)
	if err != nil {
		t.Fatalf("valid v7 footer unavailable: %v", err)
	}
	if err = config.SaveUI(p, s); err != nil {
		t.Fatal(err)
	}
	after, err := config.LoadUI(p)
	if err != nil || after != s {
		t.Fatalf("footer/settings roundtrip lost: %+v %v", after, err)
	}
	data, err := os.ReadFile(p)
	if err != nil || !strings.Contains(string(data), `footer = "curated"`) {
		t.Fatal("saved configuration lost curated mode")
	}
}

func TestUIFooterNormalizesLikeSourceAndRetainsDefaultZero(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{{"", ""}, {"full", "full"}, {"curated", "curated"}, {"compact", "compact"}, {"minimal", "minimal"}, {" MINIMAL ", "minimal"}, {"unknown", "full"}} {
		t.Run(tc.raw, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "ui.toml")
			if err := os.WriteFile(p, []byte(fmt.Sprintf("schema_version=7\nsidebar_density='compact'\nfooter=%q\n", tc.raw)), 0600); err != nil {
				t.Fatal(err)
			}
			s, err := config.LoadUI(p)
			if err != nil || s.Footer != tc.want {
				t.Fatalf("load normalization: %+v %v", s, err)
			}
			s.Footer = tc.raw
			if err = config.SaveUI(p, s); err != nil {
				t.Fatal(err)
			}
			s, err = config.LoadUI(p)
			if err != nil || s.Footer != tc.want {
				t.Fatalf("save normalization: %+v %v", s, err)
			}
		})
	}
}

func TestUIFooterLegacyLoadRemainsReadOnlyUntilExplicitUpgrade(t *testing.T) {
	for version := 1; version <= 6; version++ {
		p := filepath.Join(t.TempDir(), "ui.toml")
		original := fmt.Sprintf("schema_version=%d\nsidebar_density='full'\n", version)
		if err := os.WriteFile(p, []byte(original), 0600); err != nil {
			t.Fatal(err)
		}
		s, err := config.LoadUI(p)
		if err != nil || s.Footer != "" {
			t.Fatalf("v%d load: %+v %v", version, s, err)
		}
		if b, err := os.ReadFile(p); err != nil || string(b) != original {
			t.Fatal("reading old preferences rewrote source")
		}
		s.Footer = "curated"
		if err = config.SaveUI(p, s); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(p)
		if err != nil || !strings.Contains(string(b), "schema_version = 7") {
			t.Fatal("explicit save did not upgrade")
		}
	}
}

func TestUIFooterRejectsVersionOrTypeMismatchWithoutOverwritingSource(t *testing.T) {
	texts := []string{"schema_version=8\nfooter='curated'", "schema_version=7\nfooter=true", "schema_version=7\nfooter=['PRIVATE-SYNTHETIC']"}
	for version := 1; version <= 6; version++ {
		texts = append(texts, fmt.Sprintf("schema_version=%d\nfooter='curated'", version))
	}
	for _, text := range texts {
		p := filepath.Join(t.TempDir(), "ui.toml")
		if err := os.WriteFile(p, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := config.LoadUI(p); err == nil || strings.Contains(err.Error(), "PRIVATE-SYNTHETIC") {
			t.Fatal("unsupported footer accepted or source exposed")
		}
		if err := config.SaveUI(p, config.UISettings{SidebarDensity: "compact", Footer: "full"}); err == nil {
			t.Fatal("invalid file replaced")
		}
		if b, err := os.ReadFile(p); err != nil || string(b) != text {
			t.Fatal("invalid source was modified")
		}
	}
}
