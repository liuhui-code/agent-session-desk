package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type terminalHostConfig struct{ Path, Profile string }

func localTerminalHost(root string) (terminalHostConfig, error) {
	path := filepath.Join(root, "local-host.json")
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return terminalHostConfig{}, nil
	}
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64*1024 {
		return terminalHostConfig{}, fmt.Errorf("local Windows Terminal configuration is invalid")
	}
	data, err := readHostManifest(path, 64*1024)
	if err != nil {
		return terminalHostConfig{}, err
	}
	data = []byte(strings.TrimPrefix(string(data), "\ufeff"))
	var doc struct {
		Terminal    struct{ Path, SHA256 string }
		ProfileGuid string
	}
	if json.Unmarshal(data, &doc) != nil || doc.Terminal.Path == "" || len(doc.Terminal.SHA256) != 64 {
		return terminalHostConfig{}, fmt.Errorf("local Windows Terminal configuration is invalid")
	}
	var file *os.File
	if !filepath.IsAbs(doc.Terminal.Path) {
		local := filepath.FromSlash(strings.ReplaceAll(doc.Terminal.Path, `\`, "/"))
		if !filepath.IsLocal(local) {
			return terminalHostConfig{}, fmt.Errorf("portable Windows Terminal path must stay inside the installation")
		}
		dir, openErr := os.OpenRoot(root)
		if openErr != nil {
			return terminalHostConfig{}, fmt.Errorf("portable Windows Terminal directory is unavailable")
		}
		defer dir.Close()
		file, err = dir.Open(local)
		doc.Terminal.Path = filepath.Join(root, local)
	} else {
		file, err = os.Open(doc.Terminal.Path)
	}
	if err != nil {
		return terminalHostConfig{}, fmt.Errorf("configured Windows Terminal is unavailable")
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil || !fileInfo.Mode().IsRegular() || fileInfo.Size() > 64<<20 {
		return terminalHostConfig{}, fmt.Errorf("configured Windows Terminal is not a bounded regular executable")
	}
	hash := sha256.New()
	if _, err = io.Copy(hash, io.LimitReader(file, 64<<20)); err != nil {
		return terminalHostConfig{}, err
	}
	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), doc.Terminal.SHA256) {
		return terminalHostConfig{}, fmt.Errorf("configured Windows Terminal has changed")
	}
	if strings.ContainsAny(doc.ProfileGuid, ";\r\n") {
		return terminalHostConfig{}, fmt.Errorf("invalid local Windows Terminal profile")
	}
	return terminalHostConfig{Path: doc.Terminal.Path, Profile: doc.ProfileGuid}, nil
}
