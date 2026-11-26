//go:build windows

package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/instance"
	"github.com/liuhui-code/agent-session-desk/internal/pet"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"io"
	"os"
	"time"
)

func prepareWorkbench(ctx context.Context, args []string, inside bool, token string, options backend.Options, stateDir string, input io.Reader, output, errorOutput io.Writer) (*workbenchHost, bool, error) {
	if input != os.Stdin || output != os.Stdout {
		return &workbenchHost{}, false, nil
	}
	file, _, err := metadataPath(options, stateDir)
	if err != nil {
		return nil, false, err
	}
	kind, location, err := backend.Location(options)
	if err != nil {
		return nil, false, err
	}
	key, err := instance.ScopeKey(kind, location, file)
	if err != nil {
		return nil, false, err
	}
	short, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	notify := func() (bool, error) {
		found, activated, e := instance.Notify(short, key)
		if found {
			if e != nil || !activated {
				fmt.Fprintln(errorOutput, "Workbench already running; could not restore its window automatically. Switch to the original window.")
			}
			return true, nil
		}
		return false, e
	}
	if !inside {
		err = instance.WithLaunchGate(short, key, func() error {
			found, e := notify()
			if found || e != nil {
				return e
			}
			launchToken, e := session.NewID()
			if e != nil {
				return e
			}
			handled, e := openTerminalHost(ctx, args, false, launchToken, input, output)
			if e != nil {
				return e
			}
			if !handled {
				return fmt.Errorf("Windows workbench handoff unavailable")
			}
			if e = instance.WaitReady(short, key); e != nil {
				return fmt.Errorf("workbench startup acknowledgement unavailable; do not repeat automatically: %w", e)
			}
			return nil
		})
		return nil, true, err
	}
	var owner *pet.Owner
	if token != "" {
		bindCtx, stop := context.WithTimeout(ctx, 4*time.Second)
		owner, err = pet.CaptureWindow(bindCtx, token)
		stop()
	} else {
		owner, err = pet.CaptureOwner()
	}
	if err != nil {
		return nil, false, err
	}
	listener, err := instance.Listen(ctx, key, owner.Activate)
	if errors.Is(err, instance.ErrAlreadyRunning) {
		defer owner.Close()
		_, err = notify()
		return nil, true, err
	}
	if err != nil {
		owner.Close()
		return nil, false, err
	}
	if token != "" {
		titleCtx, stop := context.WithTimeout(ctx, 3*time.Second)
		err = owner.Present(titleCtx)
		stop()
		if err != nil {
			listener.Close()
			owner.Close()
			return nil, false, err
		}
	}
	if err = listener.BindWindow(owner.WindowHandle()); err != nil {
		listener.Close()
		owner.Close()
		return nil, false, err
	}
	return &workbenchHost{focus: owner.Activate, close: func() { listener.Close(); owner.Close() }}, false, nil
}
