package store

import (
	"fmt"
	"os"
	"path/filepath"
)

const outputMarker = "Agent Session Desk screen records v1\n"

func prepareOutputRoot(root string, create bool) (bool, error) {
	info, err := os.Lstat(root)
	if os.IsNotExist(err) {
		if !create {
			return false, nil
		}
		if err = os.Mkdir(root, 0700); err != nil {
			return false, err
		}
		marker, err := os.OpenFile(filepath.Join(root, ".format"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return false, err
		}
		_, err = marker.WriteString(outputMarker)
		if err == nil {
			err = marker.Sync()
		}
		closeErr := marker.Close()
		if err == nil {
			err = closeErr
		}
		return err == nil, err
	}
	if err != nil {
		return false, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("output directory must not be a link or non-directory")
	}
	path := filepath.Join(root, ".format")
	info, err = os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() != int64(len(outputMarker)) {
		return false, fmt.Errorf("unrecognized output directory; files preserved")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if string(data) != outputMarker {
		return false, fmt.Errorf("unrecognized output directory; files preserved")
	}
	return true, nil
}
