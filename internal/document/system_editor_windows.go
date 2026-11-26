//go:build windows

package document

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// OpenInSystemNotepad is an explicit per-workbench choice. It neither writes
// associations nor changes how the normal Open function delegates to Windows.
func OpenInSystemNotepad(ctx context.Context, path string) error {
	return openSystemEditor(ctx, path, systemNotepad)
}
func openSystemEditor(ctx context.Context, path string, launch func(string) error) error {
	resolved, err := Resolve(path, "")
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return launch(resolved)
}
func systemNotepad(path string) error {
	directory, err := windows.GetSystemDirectory()
	if err != nil {
		return err
	}
	cmd := exec.Command(filepath.Join(directory, "notepad.exe"), path)
	cmd.Dir = filepath.Dir(path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not open system Notepad: %w", err)
	}
	// An external editor's lifetime belongs to the user after launch. Releasing
	// this handle must not bind it to a UI request timeout or view cancellation.
	return cmd.Process.Release()
}
