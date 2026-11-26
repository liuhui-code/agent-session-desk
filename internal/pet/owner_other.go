//go:build !windows

package pet

import (
	"context"
	"fmt"
)

type Owner struct{}

func CaptureOwner() (*Owner, error) { return nil, fmt.Errorf("desktop reminder requires Windows") }
func (*Owner) Activate() bool       { return false }
func (*Owner) Close()               {}

func CaptureWindow(context.Context, string) (*Owner, error) {
	return nil, fmt.Errorf("desktop reminder requires Windows")
}

func (*Owner) Present(context.Context) error { return fmt.Errorf("desktop reminder requires Windows") }

func (*Owner) WindowHandle() uintptr { return 0 }
