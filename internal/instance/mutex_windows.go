//go:build windows

package instance

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"runtime"
	"strings"
	"time"
)

func currentSID(process windows.Handle) (string, error) {
	var token windows.Token
	if err := windows.OpenProcessToken(process, windows.TOKEN_QUERY, &token); err != nil {
		return "", err
	}
	defer token.Close()
	user, err := token.GetTokenUser()
	if err != nil {
		return "", err
	}
	return user.User.Sid.String(), nil
}
func objectName(key string) (string, error) {
	if len(key) != 64 || strings.IndexFunc(key, func(r rune) bool { return !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') }) >= 0 {
		return "", fmt.Errorf("invalid workbench scope key")
	}
	sid, err := currentSID(windows.CurrentProcess())
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(sid + ":" + key))
	return fmt.Sprintf("ASD.Workbench.%x", sum), nil
}
func openMutex(name string) (windows.Handle, error) {
	h, err := windows.CreateMutex(nil, false, wide(`Local\`+name))
	if h != 0 && errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return h, nil
	}
	return h, err
}

// Callers hold an OS thread until releasing an acquired Windows mutex.
func waitMutex(ctx context.Context, h windows.Handle, wait bool) (bool, error) {
	for {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		ms := uint32(0)
		if wait {
			ms = 50
		}
		status, err := windows.WaitForSingleObject(h, ms)
		if err != nil {
			return false, err
		}
		switch status {
		case windows.WAIT_OBJECT_0, windows.WAIT_ABANDONED:
			return true, nil
		case uint32(windows.WAIT_TIMEOUT):
			if !wait {
				return false, nil
			}
		default:
			return false, fmt.Errorf("unexpected workbench mutex state")
		}
	}
}
func WithLaunchGate(ctx context.Context, key string, fn func() error) error {
	name, err := objectName(key)
	if err != nil {
		return err
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	h, err := openMutex(name + ".Launch")
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	if _, err = waitMutex(ctx, h, true); err != nil {
		return err
	}
	defer windows.ReleaseMutex(h)
	return fn()
}
func owned(ctx context.Context, name string) (bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	h, err := openMutex(name)
	if err != nil {
		return false, err
	}
	defer windows.CloseHandle(h)
	acquired, err := waitMutex(ctx, h, false)
	if err != nil {
		return false, err
	}
	if acquired {
		windows.ReleaseMutex(h)
	}
	return !acquired, err
}
func pause(ctx context.Context) error {
	timer := time.NewTimer(25 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
