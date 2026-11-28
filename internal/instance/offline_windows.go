//go:build windows

package instance

import (
	"context"
	"fmt"
	"golang.org/x/sys/windows"
	"runtime"
)

// WithStopped excludes first-launch handoff and holds the existing workbench
// lifetime mutex for the callback. It never notifies, closes or kills a UI.
func WithStopped(ctx context.Context, key string, fn func() error) error {
	return WithLaunchGate(ctx, key, func() error {
		name, err := objectName(key)
		if err != nil {
			return err
		}
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		h, err := openMutex(name)
		if err != nil {
			return err
		}
		defer windows.CloseHandle(h)
		acquired, err := waitMutex(ctx, h, false)
		if err != nil {
			return err
		}
		if !acquired {
			return fmt.Errorf("close the original workbench before applying or rolling back an update")
		}
		defer windows.ReleaseMutex(h)
		return fn()
	})
}
