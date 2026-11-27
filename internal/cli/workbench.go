package cli

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/hotkey"
	"io"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/document"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/pet"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

func runWorkbench(ctx context.Context, service *manager.Manager, options backend.Options, stateDir, configPath string, desktopPet bool, documentApp string, focus func() bool, input io.Reader, output io.Writer) error {
	settings := withQuickLaunch(loadUIOptions(options, stateDir), options, stateDir, configPath)
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
		if e := hotkey.Run(petCtx, func() { program.Send(ui.GlobalActivate{}) }); e != nil {
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
