package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Four fixed publication/temporary slots bound disk use, including interrupted
// exports. Occupied or interrupted slots are never overwritten or auto-pruned.
const ExportSlots = 4

var ErrSlotsFull = errors.New("diagnostic export slots occupied; remove an old report or inspect interrupted temporary files")

func Export(ctx context.Context, dir string, report Report) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	data, err := report.Encode()
	if err != nil {
		return "", err
	}
	before, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		if err = os.Mkdir(dir, 0700); err != nil && !errors.Is(err, os.ErrExist) {
			return "", err
		}
		before, err = os.Lstat(dir)
	}
	if err != nil || !safeDirectory(before) {
		return "", errors.New("diagnostic directory must be a regular directory")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", err
	}
	defer root.Close()
	info, err := root.Stat(".")
	if err != nil || !os.SameFile(before, info) {
		return "", errors.New("diagnostic directory changed")
	}
	for slot := 1; slot <= ExportSlots; slot++ {
		name := fmt.Sprintf("diagnostic-%d.json", slot)
		if _, err = root.Lstat(name); !errors.Is(err, os.ErrNotExist) {
			if err != nil {
				return "", err
			}
			continue
		}
		temporary := fmt.Sprintf(".diagnostic-%d.tmp", slot)
		f, err := root.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		original, statErr := f.Stat()
		_, writeErr := f.Write(data)
		syncErr := f.Sync()
		closeErr := f.Close()
		removeOwn := func() {
			current, e := root.Lstat(temporary)
			if e == nil && original != nil && os.SameFile(original, current) {
				_ = root.Remove(temporary)
			}
		}
		if err = errors.Join(statErr, writeErr, syncErr, closeErr, ctx.Err()); err != nil {
			removeOwn()
			return "", err
		}
		current, err := root.Lstat(temporary)
		if err != nil || !current.Mode().IsRegular() || !os.SameFile(original, current) {
			return "", errors.New("diagnostic temporary file identity changed")
		}
		err = root.Link(temporary, name)
		removeOwn()
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, name), nil
	}
	return "", ErrSlotsFull
}
