//go:build windows

// Windows notification handling selectively adapted from dark-mode-go v0.0.2,
// Copyright (c) 2025 Thiago Kenji Okada. SPDX-License-Identifier: MIT
// Commit cfb25c8db9bff8e2e1f993e90f26326d4e68ed44. Modified: bounded DWORD
// reads, thread-agnostic registration, latest observation and joined teardown.
package appearance

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const windowsThemePath = `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`

func openSystem(ctx context.Context) (Connection, error) { return openWindowsAt(ctx, windowsThemePath) }

// The path parameter is only for isolated native tests; production is fixed HKCU.
func openWindowsAt(parent context.Context, path string) (Connection, error) {
	if e := parent.Err(); e != nil {
		return nil, e
	}
	k, e := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE|registry.NOTIFY)
	if e != nil {
		return nil, e
	}
	changed, e := windows.CreateEvent(nil, 0, 0, nil)
	if e != nil {
		_ = k.Close()
		return nil, e
	}
	stopped, e := windows.CreateEvent(nil, 1, 0, nil)
	if e != nil {
		_ = windows.CloseHandle(changed)
		_ = k.Close()
		return nil, e
	}
	ctx, cancel := context.WithCancel(parent)
	s := &stream{events: make(chan Observation, 1), done: make(chan struct{}), stop: cancel}
	go func() {
		defer close(s.done)
		defer func() {
			s.closeErr = errors.Join(k.Close(), windows.CloseHandle(changed), windows.CloseHandle(stopped))
		}()
		defer cancel()
		stopDone := make(chan struct{})
		stopSignal := context.AfterFunc(ctx, func() {
			defer close(stopDone)
			_ = windows.SetEvent(stopped)
		})
		defer func() {
			if !stopSignal() {
				<-stopDone // Never signal a closed/reused native event handle.
			}
		}()
		arm := func() error {
			return windows.RegNotifyChangeKeyValue(windows.Handle(k), false, windows.REG_NOTIFY_CHANGE_LAST_SET|windows.REG_NOTIFY_THREAD_AGNOSTIC, changed, true)
		}
		fail := func(e error) { s.publish(Observation{Source: "windows", Err: e}) }
		if ctx.Err() != nil {
			return
		}
		if e := arm(); e != nil {
			fail(e)
			return
		}
		s.publish(windowsTheme(k))
		for {
			if ctx.Err() != nil {
				return
			}
			which, e := windows.WaitForMultipleObjects([]windows.Handle{changed, stopped}, false, windows.INFINITE)
			if ctx.Err() != nil {
				return
			}
			if e != nil {
				fail(e)
				return
			}
			switch which {
			case windows.WAIT_OBJECT_0:
				if e := arm(); e != nil {
					fail(e)
					return
				}
				s.publish(windowsTheme(k))
			case windows.WAIT_OBJECT_0 + 1:
				return
			default:
				fail(fmt.Errorf("unexpected system appearance wait status"))
				return
			}
		}
	}()
	return s, nil
}

func windowsTheme(k registry.Key) Observation {
	var data [4]byte
	n, typ, e := k.GetValue("AppsUseLightTheme", data[:])
	v := Observation{Source: "windows"}
	if e != nil || n != len(data) || typ != registry.DWORD {
		v.Err = errors.New("Windows appearance value unavailable")
		return v
	}
	switch binary.LittleEndian.Uint32(data[:]) {
	case 0:
		v.Theme = "dark"
	case 1:
		v.Theme = "light"
	default:
		v.Err = errors.New("Windows appearance value unrecognized")
	}
	return v
}
