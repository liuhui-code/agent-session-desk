package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func readPrivateDocument(path string, value any) error {
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() {
		return fmt.Errorf("credential file must be an existing regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("credential file unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !os.SameFile(before, info) {
		return fmt.Errorf("credential file identity changed")
	}
	data, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil || len(data) > 4096 {
		return fmt.Errorf("invalid credential file")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil || d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("invalid credential file")
	}
	return nil
}

// The bool distinguishes publication from a concurrent existing destination.
func publishPrivateDocument(path string, data []byte) (bool, error) {
	f, err := os.CreateTemp(filepath.Dir(path), ".asd-private-*")
	if err != nil {
		return false, fmt.Errorf("credential staging failed; destination directory must exist")
	}
	stage := f.Name()
	owned, _ := f.Stat()
	defer func() {
		if now, e := os.Lstat(stage); e == nil && owned != nil && os.SameFile(owned, now) {
			_ = os.Remove(stage)
		}
	}()
	n, err := f.Write(append(data, '\n'))
	if err == nil && n != len(data)+1 {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return false, fmt.Errorf("credential write failed")
	}
	if err = os.Link(stage, path); os.IsExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("credential publication requires hard-link support and absent destination")
	}
	if err = os.Remove(stage); err != nil {
		return false, fmt.Errorf("credential published; staging cleanup failed; preserve destination")
	}
	return true, nil
}
