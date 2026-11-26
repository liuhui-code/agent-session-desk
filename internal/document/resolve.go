package document

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

func referencePath(ref string) (string, error) {
	if len(ref) > 4096 || strings.IndexFunc(ref, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("invalid Markdown reference")
	}
	m := markdownReference.FindStringSubmatch(ref)
	if m == nil {
		return "", fmt.Errorf("only .md or .markdown files can be opened")
	}
	path := m[1]
	if strings.Contains(path, "://") {
		u, err := url.Parse(path)
		if err != nil || u.Scheme != "file" || u.Host != "" || u.User != nil || u.RawQuery != "" {
			return "", fmt.Errorf("only local Markdown files can be opened")
		}
		path = u.Path
		if runtime.GOOS == "windows" && len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
	}
	path = filepath.FromSlash(path)
	if strings.HasPrefix(path, `\\`) || strings.HasPrefix(path, "//") {
		return "", fmt.Errorf("network and device paths are not supported")
	}
	if runtime.GOOS == "windows" && strings.HasPrefix(path, `\`) && !filepath.IsAbs(path) {
		return "", fmt.Errorf("root-relative paths require an absolute path")
	}
	remainder := path
	if runtime.GOOS == "windows" && len(path) > 1 && path[1] == ':' {
		remainder = path[2:]
		if !filepath.IsAbs(path) {
			return "", fmt.Errorf("drive-relative paths require an absolute path")
		}
	}
	if strings.ContainsRune(remainder, ':') {
		return "", fmt.Errorf("invalid local Markdown path")
	}
	return path, nil
}
func NeedsDirectory(ref string) (bool, error) {
	path, err := referencePath(ref)
	return !filepath.IsAbs(path), err
}

// Resolve never expands environment values or evaluates a command. A relative
// reference needs an explicit directory from its original terminal, not os.Getwd.
func Resolve(ref, directory string) (string, error) {
	path, err := referencePath(ref)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(path) {
		if !filepath.IsAbs(directory) {
			return "", fmt.Errorf("pane directory unknown; use an absolute Markdown path")
		}
		path = filepath.Join(directory, path)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("Markdown file unavailable: %w", err)
	}
	path = filepath.Clean(path)
	if _, err = referencePath(path); err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".md" && ext != ".markdown" {
		return "", fmt.Errorf("resolved target is not a Markdown file")
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("Markdown target is not a regular file")
	}
	return path, nil
}
