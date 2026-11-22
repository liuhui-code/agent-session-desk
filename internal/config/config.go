// Package config loads explicitly selected launch templates. It never launches
// a process, interpolates shell expressions, or searches the current directory.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type document struct {
	Version   int                 `toml:"schema_version"`
	Templates map[string]template `toml:"templates"`
}

type template struct {
	Name       string            `toml:"name"`
	Executable string            `toml:"executable"`
	Args       []string          `toml:"args"`
	Dir        string            `toml:"cwd"`
	Env        map[string]string `toml:"env"`
}

// LoadLaunch returns a launch detached from its source configuration. Relative
// cwd is anchored to the configuration file, not the invoking shell directory.
func LoadLaunch(path, id string) (session.Launch, error) {
	f, err := os.Open(path)
	if err != nil {
		return session.Launch{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return session.Launch{}, err
	}
	if !info.Mode().IsRegular() {
		return session.Launch{}, fmt.Errorf("configuration must be a regular file")
	}
	const maxBytes = 1 << 20 // Defensive read bound, not a measured resource budget.
	data, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		return session.Launch{}, err
	}
	if len(data) > maxBytes {
		return session.Launch{}, fmt.Errorf("configuration exceeds 1 MiB read limit")
	}
	var doc document
	meta, err := toml.Decode(string(data), &doc)
	if err != nil {
		var parseErr toml.ParseError
		if errors.As(err, &parseErr) {
			return session.Launch{}, fmt.Errorf("invalid TOML at line %d, column %d", parseErr.Position.Line, parseErr.Position.Col)
		}
		return session.Launch{}, fmt.Errorf("invalid TOML or field type")
	}
	if len(meta.Undecoded()) != 0 {
		return session.Launch{}, fmt.Errorf("configuration contains unknown fields")
	}
	for _, key := range meta.Keys() {
		if !knownKey(key) {
			return session.Launch{}, fmt.Errorf("configuration contains unknown fields")
		}
	}
	if doc.Version != 1 {
		return session.Launch{}, fmt.Errorf("configuration schema_version must be 1")
	}
	t, ok := doc.Templates[id]
	if id == "" || !ok {
		return session.Launch{}, fmt.Errorf("launch template not found")
	}
	base, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return session.Launch{}, err
	}
	if !filepath.IsAbs(t.Dir) {
		t.Dir = filepath.Join(base, t.Dir)
	}
	keys := make([]string, 0, len(t.Env))
	for key := range t.Env {
		if key == "" || strings.ContainsAny(key, "=\x00\r\n") {
			return session.Launch{}, fmt.Errorf("invalid environment key in launch template")
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, key+"="+t.Env[key])
	}
	launch := session.Launch{Name: t.Name, Executable: t.Executable,
		Args: append([]string(nil), t.Args...), Dir: t.Dir, Env: env}
	return launch, launch.Validate()
}

// The decoder permits case-insensitive Go field matching; the file schema does
// not. Template IDs and environment names are data, not schema field names.
func knownKey(key toml.Key) bool {
	if len(key) == 1 {
		return key[0] == "schema_version" || key[0] == "templates"
	}
	if len(key) < 2 || key[0] != "templates" {
		return false
	}
	if len(key) == 2 {
		return true
	}
	if len(key) == 4 {
		return key[2] == "env"
	}
	if len(key) == 3 {
		switch key[2] {
		case "name", "executable", "args", "cwd", "env":
			return true
		}
	}
	return false
}
