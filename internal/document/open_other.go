//go:build !windows

package document

import (
	"context"
	"fmt"
)

func Open(context.Context, string) error {
	return fmt.Errorf("external Markdown opening currently requires Windows")
}
