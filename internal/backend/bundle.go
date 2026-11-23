package backend

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// This explicitly selected development bundle is not the official psmux release.
// Source, patch, build and SDK provenance are recorded in ADR-0019.
var terminalBundleHashes = map[string]string{
	"psmux.exe":       "ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937",
	"conpty.dll":      "e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492",
	"OpenConsole.exe": "b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775",
}

func prepareTerminalBundle(opts Options, kind string) (string, error) {
	if opts.TerminalBundle == "" {
		return "", nil
	}
	if kind != "psmux" || runtime.GOARCH != "amd64" {
		return "", errors.New("terminal bundle is currently supported only on Windows x64")
	}
	if opts.Binary != "" {
		return "", errors.New("terminal bundle selects its own backend; --backend cannot be combined with it")
	}
	if !filepath.IsAbs(opts.TerminalBundle) {
		return "", errors.New("terminal bundle directory must be absolute")
	}
	dir := filepath.Clean(opts.TerminalBundle)
	st, err := os.Lstat(dir)
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("terminal bundle directory is unavailable or is a link")
	}
	if err := validateTerminalBundle(dir); err != nil {
		return "", err
	}
	return dir, nil
}

func validateTerminalBundle(dir string) error {
	for name, expected := range terminalBundleHashes {
		if err := verifyBundleFile(filepath.Join(dir, name), expected); err != nil {
			return fmt.Errorf("terminal bundle %s: %w", name, err)
		}
	}
	return nil
}

func verifyBundleFile(path, expected string) error {
	st, err := os.Lstat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() < 1 || st.Size() > 64*1024*1024 {
		return errors.New("expected a bounded regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return errors.New("could not open file")
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(st, opened) {
		return errors.New("file changed while opening")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, 64*1024*1024+1))
	if err != nil || n != st.Size() || hex.EncodeToString(h.Sum(nil)) != expected {
		return errors.New("SHA-256 does not match the pinned development artifact")
	}
	return nil
}

func (b *Backend) verifyTerminalBundle(ctx context.Context, name string) error {
	if b.bundle == "" {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateTerminalBundle(b.bundle); err != nil {
		return err
	}
	text, err := readTerminalEndpoint(filepath.Join(b.dataDir, b.namespace+"__"+name+".pid"))
	if err != nil {
		return fmt.Errorf("terminal bundle server identity: %w", err)
	}
	pidText, timeText, ok := strings.Cut(text, ":")
	pid, pidErr := strconv.ParseUint(pidText, 10, 32)
	created, timeErr := strconv.ParseUint(timeText, 10, 64)
	if !ok || pidErr != nil || timeErr != nil || pid == 0 || created == 0 || text != fmt.Sprintf("%d:%d", pid, created) {
		return errors.New("terminal bundle server identity is invalid")
	}
	if err := verifyBundleProcess(uint32(pid), created, b.bin, b.bundle); err != nil {
		return fmt.Errorf("terminal bundle actual host could not be verified; no terminal input sent: %w", err)
	}
	return ctx.Err()
}
