package cli

import (
	"context"
	"path/filepath"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

func loadUIOptions(options backend.Options, dir string) ui.Options {
	metadata, _, err := metadataPath(options, dir)
	if err != nil {
		return ui.Options{SidebarPreferenceError: err}
	}
	path := filepath.Join(filepath.Dir(metadata), "ui.toml")
	settings, err := config.LoadUI(path)
	return ui.Options{SidebarDensity: settings.SidebarDensity, SidebarMode: settings.SidebarMode, DefaultGroupCollapsed: settings.DefaultGroupCollapsed, CollapsedGroups: settings.CollapsedGroups, SidebarPreferenceError: err, SaveUISettings: func(ctx context.Context, settings config.UISettings) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return config.SaveUI(path, settings)
	}}
}
