package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

func withQuickLaunch(settings ui.Options, opts backend.Options, stateDir, explicit string) ui.Options {
	metadata, _, pathErr := metadataPath(opts, stateDir)
	dir := filepath.Dir(metadata)
	defaults := explicit
	if defaults == "" {
		defaults = filepath.Join(dir, "launch.toml")
	}
	settings.LoadProject = func(ctx context.Context) ([]session.Launch, error) { return config.LoadWorkspace(defaults, "default") }
	last := filepath.Join(dir, "last-launch.toml")
	settings.LoadQuickLaunch = func(ctx context.Context) (session.Launch, error) {
		if err := ctx.Err(); err != nil {
			return session.Launch{}, err
		}
		if pathErr != nil {
			return session.Launch{}, pathErr
		}
		if explicit == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return session.Launch{}, err
			}
			l := session.Launch{Executable: "sh", Dir: home}
			if runtime.GOOS == "windows" {
				l.Executable = "powershell.exe"
				l.Args = []string{"-NoLogo", "-NoProfile", "-NoExit"}
			}
			if err := config.EnsureDefaultLaunch(defaults, l); err != nil {
				return session.Launch{}, err
			}
		}
		return config.LoadQuickLaunch(defaults, last)
	}
	settings.RememberLaunch = func(ctx context.Context, l session.Launch) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if pathErr != nil {
			return pathErr
		}
		return config.SaveLastLaunch(last, l)
	}
	return settings
}

func rememberCLIStart(ctx context.Context, opts backend.Options, stateDir string, l session.Launch) error {
	settings := withQuickLaunch(ui.Options{}, opts, stateDir, "")
	if err := settings.RememberLaunch(ctx, l); err != nil {
		return fmt.Errorf("terminal started; last launch settings could not be saved")
	}
	return nil
}
