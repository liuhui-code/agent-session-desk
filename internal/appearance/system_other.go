//go:build !windows

package appearance

import (
	"context"
	"errors"
)

func openSystem(context.Context) (Connection, error) {
	return nil, errors.New("native system appearance is unavailable on this platform")
}
