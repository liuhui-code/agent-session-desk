//go:build !windows

package hotkey

import "context"

func Run(ctx context.Context, activate func()) error { return nil }
