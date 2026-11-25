package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
)

// Explicit flags always win. A portable bundle beside the executable uses the
// same private directories as its launcher, including the manifest namespace.
func portableOptions(executable string, opts backend.Options, stateDir string) (backend.Options, string, error) {
	root := filepath.Dir(executable)
	if opts.Binary != "" || opts.TerminalBundle != "" {
		return opts, stateDir, nil
	}
	bundle := filepath.Join(root, "terminal-runtime")
	if _, err := os.Stat(filepath.Join(bundle, "psmux.exe")); os.IsNotExist(err) {
		return opts, stateDir, nil
	} else if err != nil {
		return opts, stateDir, err
	}
	manifestPath := filepath.Join(root, "alpha.json")
	data, err := readHostManifest(manifestPath, 1<<20)
	if err != nil {
		return opts, stateDir, err
	}
	var manifest struct {
		Namespace string `json:"namespace"`
	}
	// The local launcher writes UTF-8 BOM for PowerShell 5.1 compatibility.
	if len(data) >= 3 && string(data[:3]) == "\xef\xbb\xbf" {
		data = data[3:]
	}
	if json.Unmarshal(data, &manifest) != nil {
		return opts, stateDir, fmt.Errorf("invalid portable manifest")
	}
	opts.TerminalBundle = bundle
	if opts.RuntimeDir == "" {
		opts.RuntimeDir = filepath.Join(root, "data", "runtime")
	}
	if stateDir == "" {
		stateDir = filepath.Join(root, "data", "state")
	}
	if opts.Namespace == "" {
		opts.Namespace = manifest.Namespace
		if opts.Namespace == "" {
			opts.Namespace = "asd-release-alpha1"
		}
	}
	return opts, stateDir, nil
}

func readHostManifest(path string, limit int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("host manifest must be a bounded regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return nil, fmt.Errorf("host manifest changed while opening")
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("host manifest exceeds read limit")
	}
	return data, nil
}
