// Package snippets loads literal, bounded, read-only instruction fragments.
package snippets

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
)

const (
	MaxFileBytes = 256 * 1024
	MaxItems     = 64 // per file; one global and one project file
	MaxBodyBytes = 4096
)

type Item struct{ ID, Title, Body, Source string }
type Library struct {
	Items      []Item
	ProjectDir string
}
type entry struct {
	Title string `toml:"title"`
	Body  string `toml:"body"`
}

// Load reads exactly two optional files, never walks ancestors or expands text.
// Same IDs in the two files remain distinct, with an explicit source label.
func Load(ctx context.Context, globalPath, projectDir string) (Library, error) {
	out := Library{ProjectDir: projectDir}
	for _, src := range []struct{ path, source string }{{projectPath(projectDir), "Project"}, {globalPath, "Global"}} {
		if err := ctx.Err(); err != nil {
			return Library{}, err
		}
		rows, err := loadFile(src.path, src.source)
		if err != nil {
			return Library{}, fmt.Errorf("%s snippets: %w", src.source, err)
		}
		out.Items = append(out.Items, rows...)
	}
	if err := ctx.Err(); err != nil {
		return Library{}, err
	}
	return out, nil
}
func projectPath(dir string) string {
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, ".asd", "snippets.toml")
}
func loadFile(path, source string) ([]Item, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot open configuration")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > MaxFileBytes {
		return nil, fmt.Errorf("expected regular file of at most 256 KiB")
	}
	data, err := io.ReadAll(io.LimitReader(f, MaxFileBytes+1))
	if err != nil || len(data) > MaxFileBytes {
		return nil, fmt.Errorf("configuration read limit")
	}
	var d struct {
		Version  int              `toml:"schema_version"`
		Snippets map[string]entry `toml:"snippets"`
	}
	meta, err := toml.Decode(string(data), &d)
	if err != nil || len(meta.Undecoded()) != 0 || d.Version != 1 {
		return nil, fmt.Errorf("invalid snippets TOML v1")
	}
	for _, k := range meta.Keys() {
		if !knownKey(k) {
			return nil, fmt.Errorf("unknown configuration field")
		}
	}
	if len(d.Snippets) > MaxItems {
		return nil, fmt.Errorf("at most 64 snippets per file")
	}
	ids := make([]string, 0, len(d.Snippets))
	for id := range d.Snippets {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	rows := make([]Item, 0, len(ids))
	for _, id := range ids {
		e := d.Snippets[id]
		if !validID(id) || !validText(e.Title, 120, false) || !validText(e.Body, MaxBodyBytes, true) {
			return nil, fmt.Errorf("invalid snippet ID, title or body")
		}
		rows = append(rows, Item{id, e.Title, e.Body, source})
	}
	return rows, nil
}
func knownKey(k toml.Key) bool {
	if len(k) == 1 {
		return k[0] == "schema_version" || k[0] == "snippets"
	}
	return k[0] == "snippets" && (len(k) == 2 || (len(k) == 3 && (k[2] == "title" || k[2] == "body")))
}
func validID(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
func validText(s string, limit int, multiline bool) bool {
	if strings.TrimSpace(s) == "" || len(s) > limit || !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) && !(multiline && (r == '\n' || r == '\r' || r == '\t')) {
			return false
		}
	}
	return true
}
