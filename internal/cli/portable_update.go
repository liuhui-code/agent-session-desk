package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/store"
	"github.com/liuhui-code/agent-session-desk/internal/updater"
)

func portableUpdate(ctx context.Context, args []string, out, diag io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: asd update probe|package|status|download|apply|rollback|repair|discard")
	}
	action := args[0]
	if action == "probe" {
		if len(args) != 1 {
			return fmt.Errorf("update probe takes no arguments")
		}
		return writeJSON(out, updater.CurrentProbe())
	}
	if action == "package" {
		f := flag.NewFlagSet("update package", flag.ContinueOnError)
		f.SetOutput(diag)
		var program, tag, output string
		f.StringVar(&program, "program", "", "verified release executable")
		f.StringVar(&tag, "tag", "", "release tag")
		f.StringVar(&output, "output", "", "new asd-update.json path")
		if err := f.Parse(args[1:]); err != nil {
			return err
		}
		if f.NArg() != 0 || program == "" || output == "" {
			return fmt.Errorf("update package requires --program, --tag and --output")
		}
		return updater.PackageContract(ctx, program, tag, output, probeUpdateProgram)
	}
	f := flag.NewFlagSet("update "+action, flag.ContinueOnError)
	f.SetOutput(diag)
	var root, tag string
	var confirmed bool
	f.StringVar(&root, "install-dir", "", "managed portable directory (default executable directory)")
	if action == "download" {
		f.StringVar(&tag, "tag", "", "explicit published release tag")
	}
	switch action {
	case "apply", "rollback", "repair":
		f.BoolVar(&confirmed, "confirm-closed", false, "all workbenches using this installation are closed")
	case "status", "download", "discard":
	default:
		return fmt.Errorf("unknown update action")
	}
	if err := f.Parse(args[1:]); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected update arguments")
	}
	if root == "" {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		root = filepath.Dir(exe)
	}
	if (action == "apply" || action == "rollback" || action == "repair") && !confirmed {
		return fmt.Errorf("close all workbenches and pass --confirm-closed")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	root = abs
	engine := newUpdateEngine(root, confirmed)
	switch action {
	case "status":
		s, e := engine.Status(ctx)
		if e != nil {
			return e
		}
		return writeJSON(out, s)
	case "download":
		c, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		if err = engine.Download(c, tag, updater.Client{}); err != nil {
			return err
		}
		fmt.Fprintln(out, "Verified update staged. Close the workbench, then run update apply --confirm-closed.")
		return nil
	case "discard":
		if err = engine.Discard(ctx); err != nil {
			return err
		}
		fmt.Fprintln(out, "Staged download discarded; current program unchanged.")
		return nil
	}
	var receipt updater.Receipt
	switch action {
	case "apply":
		receipt, err = engine.Apply(ctx)
	case "rollback":
		receipt, err = engine.Rollback(ctx)
	case "repair":
		receipt, err = engine.Repair(ctx)
	}
	if err != nil {
		return err
	}
	return writeJSON(out, receipt)
}
func newUpdateEngine(root string, confirmed bool) updater.Engine {
	return updater.Engine{Root: root, Offline: func(ctx context.Context, fn func() error) error {
		if !confirmed {
			return fmt.Errorf("close all workbenches for this installation and pass --confirm-closed; no window will be closed automatically")
		}
		return offlinePortableUpdate(ctx, root, fn)
	}, Probe: probeUpdateProgram, Check: func(ctx context.Context, path string) (int, error) {
		r, e := store.Check(ctx, path)
		return r.SchemaVersion, e
	}, Backup: func(ctx context.Context, src, dst string) error {
		return store.SnapshotFixed(ctx, src, dst, filepath.Join(filepath.Dir(dst), "backup.capture.tmp"))
	}}
}
