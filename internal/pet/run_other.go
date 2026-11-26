//go:build !windows

package pet

import (
	"context"
	"fmt"
)

func runNative(context.Context, Options) error { return fmt.Errorf("pet-demo requires native Windows") }
