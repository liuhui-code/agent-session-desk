//go:build !windows

package hotkey

import "context"

func Run(ctx context.Context, key string, activate func()) error { return nil }
