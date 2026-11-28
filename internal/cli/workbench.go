package cli

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/diagnostics"
	"github.com/liuhui-code/agent-session-desk/internal/hotkey"
	"github.com/liuhui-code/agent-session-desk/internal/keybindings"
	"github.com/liuhui-code/agent-session-desk/internal/release"
	"io"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/document"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/pet"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

func runWorkbench(ctx context.Context, service *manager.Manager, info backend.Info, options backend.Options, stateDir, configPath string, desktopPet bool, documentApp string, focus func() bool, input io.Reader, output io.Writer) error {
	metadata, _, e := metadataPath(options, stateDir)
	if e != nil {
		return e
	}
	keys, e := keybindings.Load(filepath.Join(filepath.Dir(metadata), "keybindings.toml"))
	if e != nil {
		return e
	}
	settings := withQuickLaunch(loadUIOptions(options, stateDir), options, stateDir, configPath)
	settings.DiagnosticRuntime = diagnostics.CurrentRuntime(info.Name, info.Version)
	settings.DiagnosticRuntime.ProductVersion = release.Version
	settings.ExportDiagnostics = func(ctx context.Context, report diagnostics.Report) (string, error) {
		return diagnostics.Export(ctx, filepath.Join(filepath.Dir(metadata), "diagnostics"), report)
	}
	settings = withRecoveryControls(settings, service, options, stateDir)
	settings.Keybindings = keys
	settings.AttentionFocus = focus
	settings.OpenDocument = document.Open
	if documentApp == "system-notepad" {
		settings.OpenDocument = document.OpenInSystemNotepad
	}
	petCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var companion *pet.Companion
	requests := make(chan pet.Activation, 1)
	if desktopPet {
		companion = pet.NewCompanion(func(a pet.Activation) {
			select {
			case requests <- a:
			default:
			}
		})
		settings.AttentionObserver = companion.Publish
	}
	model := ui.NewWithOptions(ctx, service, settings)
	defer model.Close()
	program := tea.NewProgram(model, tea.WithInput(input), tea.WithOutput(output), tea.WithContext(ctx))
	go func() {
		if e := hotkey.Run(petCtx, keys.Wake(), func() { program.Send(ui.GlobalActivate{}) }); e != nil {
			program.Send(ui.HotkeyFailure{})
		}
	}()
	if companion != nil {
		done := make(chan struct{})
		go func() {
			defer close(done)
			if err := pet.Run(petCtx, pet.Options{Live: companion}); err != nil && petCtx.Err() == nil {
				program.Send(ui.AttentionCompanionFailure{Err: err})
			}
		}()
		go func() {
			for {
				select {
				case <-petCtx.Done():
					return
				case a := <-requests:
					if a.ID == "" {
						program.Send(ui.WorkbenchActivate{})
					} else {
						program.Send(ui.AttentionActivate{ID: a.ID, Version: a.Version})
					}
				}
			}
		}()
		defer func() {
			cancel()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
			}
		}()
	}
	_, err := program.Run()
	return err
}
