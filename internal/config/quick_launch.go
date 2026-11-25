package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// The authored default stays read-only. The last successful launch has its own
// bounded private file; a broken snapshot never silently selects another command.
func LoadQuickLaunch(defaultPath, lastPath string) (session.Launch, error) {
	if info, err := os.Lstat(lastPath); err == nil {
		if !info.Mode().IsRegular() {
			return session.Launch{}, fmt.Errorf("last launch must be a regular file")
		}
		return LoadLaunch(lastPath, "last")
	} else if !errors.Is(err, os.ErrNotExist) {
		return session.Launch{}, err
	}
	return LoadLaunch(defaultPath, "default")
}

func EnsureDefaultLaunch(path string, l session.Launch) error {
	if _, err := os.Lstat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return writeLaunchSnapshot(path, "default", l, false)
}

func SaveLastLaunch(path string, l session.Launch) error {
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("last launch must be a regular file")
		}
		if _, err := LoadLaunch(path, "last"); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return writeLaunchSnapshot(path, "last", l, true)
}

func writeLaunchSnapshot(path, id string, l session.Launch, replace bool) error {
	if err := l.Validate(); err != nil {
		return fmt.Errorf("invalid launch configuration")
	}
	t := template{Name: l.Name, Executable: l.Executable, Args: append([]string(nil), l.Args...), Dir: l.Dir, Env: map[string]string{}}
	for _, entry := range l.Env {
		k, v, ok := strings.Cut(entry, "=")
		if !ok {
			return fmt.Errorf("invalid launch environment")
		}
		if runtime.GOOS == "windows" {
			k = strings.ToUpper(k)
		}
		t.Env[k] = v
	}
	var data bytes.Buffer
	if err := toml.NewEncoder(&data).Encode(document{Version: 1, Templates: map[string]template{id: t}}); err != nil {
		return fmt.Errorf("could not encode launch configuration")
	}
	if data.Len() > 1<<20 {
		return fmt.Errorf("launch configuration exceeds 1 MiB")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, ".launch-*.tmp")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if _, err = temp.Write(data.Bytes()); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if _, err = LoadLaunch(name, id); err != nil {
		return err
	}
	if !replace {
		err = os.Link(name, path)
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("last launch destination changed")
	}
	return os.Rename(name, path)
}
