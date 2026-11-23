package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestUILaunchOptionsCarryAndSaveManualSplitWithoutChangingOtherState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ui.toml")
	settings := config.UISettings{Theme: "light", SidebarDensity: "minimal", SidebarMode: "flat", SidebarPercent: 45, DefaultGroupCollapsed: true}
	if e := config.SaveUI(path, settings); e != nil {
		t.Fatal(e)
	}
	marker := filepath.Join(dir, "launch.toml")
	if e := os.WriteFile(marker, []byte("synthetic launch sentinel"), 0600); e != nil {
		t.Fatal(e)
	}
	opts := loadUIOptions(backend.Options{RuntimeDir: t.TempDir(), Namespace: "asd-ui-prefs"}, dir)
	if opts.Theme != "light" || opts.SidebarPreferenceError != nil || opts.SidebarPercent != 45 || opts.SidebarDensity != "minimal" || opts.SidebarMode != "flat" || !opts.DefaultGroupCollapsed {
		t.Fatalf("launch lost prefs: %+v", opts)
	}
	settings.SidebarPercent = 0
	if e := opts.SaveUISettings(context.Background(), settings); e != nil {
		t.Fatal(e)
	}
	got, e := config.LoadUI(path)
	if e != nil || got != settings {
		t.Fatalf("auto roundtrip: %+v %v", got, e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	settings.SidebarPercent = 60
	if e = opts.SaveUISettings(ctx, settings); e == nil {
		t.Fatal("cancelled writer changed preferences")
	}
	got, e = config.LoadUI(path)
	if e != nil || got.SidebarPercent != 0 {
		t.Fatal("cancelled save changed auto")
	}
	b, e := os.ReadFile(marker)
	if e != nil || string(b) != "synthetic launch sentinel" {
		t.Fatal("UI save touched launch config")
	}
}
