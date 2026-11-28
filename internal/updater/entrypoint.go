package updater

import (
	"context"
	"path/filepath"
)

// Entrypoint resolves the currently committed program for the fixed launcher.
// It shares the update lock and validation, but never repairs or runs anything.
func Entrypoint(ctx context.Context, root string) (string, error) {
	var target string
	err := (Engine{Root: root}).locked(ctx, func(f *files) error {
		if err := noIntent(f); err != nil {
			return err
		}
		a, err := f.installation()
		if err == nil {
			target = filepath.Join(f.path, a.entry)
		}
		return err
	})
	return target, err
}
