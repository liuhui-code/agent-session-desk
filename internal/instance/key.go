// Package instance coordinates one interactive workbench per local data scope.
// Its only cross-process action is restoring an existing workbench window.
package instance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

func ScopeKey(kind, runtimePath, metadataFile string) (string, error) {
	runtimePath, err := canonical(runtimePath)
	if err != nil {
		return "", err
	}
	metadataFile, err = canonical(metadataFile)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal([]string{kind, runtimePath, metadataFile})
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

// Resolve the nearest existing ancestor without creating state directories.
func canonical(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var missing []string
	for {
		if _, err = os.Lstat(path); err == nil {
			break
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", err
		}
		missing = append(missing, filepath.Base(path))
		path = parent
	}
	path, err = finalPath(path)
	if err != nil {
		return "", err
	}
	for i := len(missing) - 1; i >= 0; i-- {
		path = filepath.Join(path, missing[i])
	}
	return normalizePath(path), nil
}
