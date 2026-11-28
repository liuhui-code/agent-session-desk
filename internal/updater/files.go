package updater

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var ErrBusy = errors.New("update operation already running")

type files struct {
	root *os.Root
	path string
}

func (e Engine) locked(ctx context.Context, fn func(*files) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	abs, err := filepath.Abs(e.Root)
	if err != nil {
		return err
	}
	st, err := os.Lstat(abs)
	if err != nil || !safeInfo(st, true) {
		return fmt.Errorf("installation must be an existing regular directory")
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return err
	}
	defer root.Close()
	f := &files{root, abs}
	if err = f.directory(work); err != nil {
		return err
	}
	lock, err := root.OpenFile(work+"/lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	st, err = lock.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Size() != 0 {
		return fmt.Errorf("update lock invalid")
	}
	if err = lockFile(lock); err != nil {
		return ErrBusy
	}
	defer unlockFile(lock)
	return fn(f)
}
func (f *files) directory(name string) error {
	st, err := f.root.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		if err = f.root.Mkdir(name, 0700); err != nil && !os.IsExist(err) {
			return err
		}
		st, err = f.root.Lstat(name)
	}
	if err != nil || !safeInfo(st, true) {
		return fmt.Errorf("update directory unavailable or linked")
	}
	return nil
}
func (f *files) read(name string, limit int64) ([]byte, error) {
	st, err := f.root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !safeInfo(st, false) || st.Size() > limit {
		return nil, fmt.Errorf("update file not a bounded regular file")
	}
	file, err := f.root.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(st, opened) {
		return nil, fmt.Errorf("update file identity changed")
	}
	b, err := io.ReadAll(io.LimitReader(file, limit+1))
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("update file exceeds bound")
	}
	return b, err
}
func (f *files) hash(name string) (string, error) {
	st, err := f.root.Lstat(name)
	if err != nil {
		return "", err
	}
	if !safeInfo(st, false) || st.Size() < 1 || st.Size() > MaxProductBytes {
		return "", fmt.Errorf("program/runtime must be bounded regular file")
	}
	file, err := f.root.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(st, opened) {
		return "", fmt.Errorf("program identity changed")
	}
	h := sha256.New()
	n, err := io.CopyBuffer(h, io.LimitReader(file, MaxProductBytes+1), make([]byte, 32*1024))
	if err != nil || n != st.Size() {
		return "", fmt.Errorf("program changed while hashing")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func hashBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func (f *files) fresh(name string, data io.Reader, limit int64) (err error) {
	file, err := f.root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	original, _ := file.Stat()
	defer func() {
		if err != nil {
			if st, e := f.root.Lstat(name); e == nil && original != nil && os.SameFile(st, original) {
				f.root.Remove(name)
			}
		}
	}()
	n, writeErr := io.CopyBuffer(file, io.LimitReader(data, limit+1), make([]byte, 32*1024))
	syncErr := file.Sync()
	closeErr := file.Close()
	if n > limit {
		return fmt.Errorf("download exceeds bound")
	}
	return errors.Join(writeErr, syncErr, closeErr)
}
func (f *files) publish(name string, data []byte) error {
	stage := name + ".tmp"
	if err := f.fresh(stage, bytes.NewReader(data), maxAlphaBytes); err != nil {
		return err
	}
	if err := f.root.Link(stage, name); err != nil {
		f.root.Remove(stage)
		return err
	}
	return f.root.Remove(stage)
}
func (f *files) replaceAlpha(ctx context.Context, old, new []byte) error {
	if err := f.fresh("alpha.update.tmp", bytes.NewReader(new), maxAlphaBytes); err != nil {
		return err
	}
	before, err := f.read("alpha.json", maxAlphaBytes)
	if err == nil && hashBytes(before) != hashBytes(old) {
		err = fmt.Errorf("installation changed; inspect update status")
	}
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		f.root.Remove("alpha.update.tmp")
		return err
	}
	return f.root.Rename("alpha.update.tmp", "alpha.json")
}
