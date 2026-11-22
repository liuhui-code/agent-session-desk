package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func metadataCommand(ctx context.Context, args []string, opts backend.Options, dir string, output io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: asd metadata check [FILE] | backup FILE | restore FILE")
	}
	action := args[0]
	switch action {
	case "check":
		if len(args) > 2 {
			return fmt.Errorf("usage: asd metadata check [FILE]")
		}
	case "backup", "restore":
		if len(args) != 2 || args[1] == "" {
			return fmt.Errorf("usage: asd metadata %s FILE", action)
		}
	default:
		return fmt.Errorf("unknown metadata operation; use check, backup or restore")
	}
	if action == "check" && len(args) == 2 {
		result, err := store.Check(ctx, args[1])
		if err != nil {
			return err
		}
		return writeJSON(output, result)
	}
	path, _, err := metadataPath(opts, dir)
	if err != nil {
		return err
	}
	if action == "check" {
		result, err := store.Check(ctx, path)
		if err != nil {
			return err
		}
		return writeJSON(output, result)
	}
	source, target := path, args[1]
	if action == "restore" {
		source, target = args[1], path
		if _, err = store.Check(ctx, source); err != nil {
			return err
		}
		if err = os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return err
		}
	}
	if err = store.Snapshot(ctx, source, target); err != nil {
		return err
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return err
	}
	return writeJSON(output, struct {
		Operation string `json:"operation"`
		Path      string `json:"path"`
		Scope     string `json:"scope"`
	}{action, target, "all stored scopes; metadata only"})
}
