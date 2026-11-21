package session

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ResolveLaunch resolves cwd and executable before a terminal can be created.
// PATH/PATHEXT use the requested environment; resolving never mutates os.Environ.
// Relative executable paths and relative PATH entries are relative to the
// requested cwd. Windows does not implicitly search the current directory.
func ResolveLaunch(l Launch) (Launch, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}
	if l.Dir == "" {
		var err error
		l.Dir, err = os.Getwd()
		if err != nil {
			return l, fmt.Errorf("working directory: %w", err)
		}
	}
	var err error
	l.Dir, err = filepath.Abs(l.Dir)
	if err != nil {
		return l, fmt.Errorf("working directory: %w", err)
	}
	st, err := os.Stat(l.Dir)
	if err != nil {
		return l, fmt.Errorf("working directory: %w", err)
	}
	if !st.IsDir() {
		return l, fmt.Errorf("working directory is not a directory: %s", l.Dir)
	}
	env := effectiveEnvironment(os.Environ(), l.Env)
	l.Executable, err = resolveExecutable(l.Executable, l.Dir, env)
	if err != nil {
		return l, err
	}
	if runtime.GOOS == "windows" {
		switch strings.ToLower(filepath.Ext(l.Executable)) {
		case ".cmd", ".bat", ".ps1":
			return l, fmt.Errorf("script entry requires an explicit shell: use cmd.exe /c or pwsh.exe -File")
		}
	}
	if strings.TrimSpace(l.Name) == "" {
		l.Name = filepath.Base(l.Executable)
	}
	// Keep the caller's slices independent of later UI/backend work.
	l.Args = append([]string(nil), l.Args...)
	l.Env = normalizeOverrides(l.Env)
	return l, nil
}

// Match the environment used by executable lookup: the last override wins.
// Windows variable names are case-insensitive even if a backend uses a map.
func normalizeOverrides(entries []string) []string {
	seen := make(map[string]bool, len(entries))
	result := make([]string, 0, len(entries))
	for i := len(entries) - 1; i >= 0; i-- {
		key, _, _ := strings.Cut(entries[i], "=")
		if runtime.GOOS == "windows" {
			key = strings.ToUpper(key)
		}
		if !seen[key] {
			result = append(result, entries[i])
			seen[key] = true
		}
	}
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

func effectiveEnvironment(base, overrides []string) map[string]string {
	env := make(map[string]string, len(base)+len(overrides))
	for _, entries := range [][]string{base, overrides} {
		for _, entry := range entries {
			key, value, found := strings.Cut(entry, "=")
			if found {
				if runtime.GOOS == "windows" {
					key = strings.ToUpper(key)
				}
				env[key] = value
			}
		}
	}
	return env
}

func resolveExecutable(command, dir string, env map[string]string) (string, error) {
	suffixes := []string{""}
	if runtime.GOOS == "windows" && filepath.Ext(command) == "" {
		pathext := env["PATHEXT"]
		if pathext == "" {
			pathext = ".COM;.EXE;.BAT;.CMD"
		}
		suffixes = nil
		for _, extension := range strings.Split(pathext, ";") {
			if strings.HasPrefix(extension, ".") && !strings.ContainsAny(extension, `/\\`) {
				suffixes = append(suffixes, extension)
			}
		}
	}
	search := []string{command}
	if !filepath.IsAbs(command) && !strings.ContainsRune(command, os.PathSeparator) && !(runtime.GOOS == "windows" && strings.ContainsAny(command, `/\\`)) {
		search = nil
		for _, entry := range filepath.SplitList(env["PATH"]) {
			if entry == "" { // No implicit cwd search; use an explicit '.' entry.
				continue
			}
			search = append(search, filepath.Join(entry, command))
		}
	}
	for _, candidate := range search {
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(dir, candidate)
		}
		for _, suffix := range suffixes {
			path := filepath.Clean(candidate + suffix)
			st, err := os.Stat(path)
			if err != nil || st.IsDir() {
				continue
			}
			if runtime.GOOS != "windows" && st.Mode()&0111 == 0 {
				continue
			}
			return path, nil
		}
	}
	return "", fmt.Errorf("executable not found or not executable: %s", command)
}
