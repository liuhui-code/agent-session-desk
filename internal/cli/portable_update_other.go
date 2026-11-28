//go:build !windows

package cli

import (
	"context"
	"fmt"
)

func offlinePortableUpdate(context.Context, string, func() error) error {
	return fmt.Errorf("portable installation changes require native Windows")
}
