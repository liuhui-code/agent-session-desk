package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
)

const MaxRecentDirs = 20
const maxRecentDirsBytes = 192 * 1024

type recentDirsDocument struct {
	Version     int      `toml:"schema_version"`
	Directories []string `toml:"directories"`
}

func LoadRecentDirs(ctx context.Context, path string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("最近目录必须是普通文件")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return nil, fmt.Errorf("最近目录文件在读取前变化")
	}
	data, err := io.ReadAll(io.LimitReader(f, maxRecentDirsBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxRecentDirsBytes {
		return nil, fmt.Errorf("最近目录文件超过读取上限")
	}
	var doc recentDirsDocument
	meta, err := toml.Decode(string(data), &doc)
	if err != nil || doc.Version != 1 || len(meta.Undecoded()) != 0 {
		return nil, fmt.Errorf("最近目录格式或版本不受支持")
	}
	dirs, err := validRecentDirs(doc.Directories)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return dirs, nil
}

func RememberDirs(ctx context.Context, path string, dirs []string) error {
	fresh, err := validRecentDirs(dirs)
	if err != nil {
		return err
	}
	if len(fresh) == 0 {
		return ctx.Err()
	}
	old, err := LoadRecentDirs(ctx, path)
	if err != nil {
		return err
	}
	merged := make([]string, 0, MaxRecentDirs)
	seen := map[string]bool{}
	for _, value := range append(fresh, old...) {
		key := value
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if !seen[key] {
			seen[key] = true
			merged = append(merged, value)
			if len(merged) == MaxRecentDirs {
				break
			}
		}
	}
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".recent-dirs-*.tmp")
	if err != nil {
		return err
	}
	temp := f.Name()
	defer os.Remove(temp)
	closed := false
	defer func() {
		if !closed {
			_ = f.Close()
		}
	}()
	if err = f.Chmod(0600); err != nil {
		return err
	}
	if err = toml.NewEncoder(f).Encode(recentDirsDocument{1, merged}); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	closed = true
	if _, err = LoadRecentDirs(ctx, temp); err != nil {
		return err
	}
	if _, err = LoadRecentDirs(ctx, path); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

func validRecentDirs(dirs []string) ([]string, error) {
	if len(dirs) > MaxRecentDirs {
		return nil, fmt.Errorf("最近目录数量超过上限")
	}
	result := make([]string, 0, len(dirs))
	seen := map[string]bool{}
	for _, value := range dirs {
		if strings.TrimSpace(value) == "" || len(value) > 4096 || !utf8.ValidString(value) {
			return nil, fmt.Errorf("最近目录条目无效")
		}
		for _, r := range value {
			if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
				return nil, fmt.Errorf("最近目录需要单行文本")
			}
		}
		key := value
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, value)
		}
	}
	return result, nil
}
