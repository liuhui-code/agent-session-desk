// Adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/session/utils.go and internal/ui/pathfuzzy.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: context, result bounds,
// native Windows paths, control rejection and explicit exact-directory state.
package pathsuggest

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sahilm/fuzzy"
)

const MaxResults = 20
const maxScan = 10000

type Completion struct {
	Paths []string
	Exact bool
}

func Safe(path string) bool {
	if path == "" || len(path) > 4096 || !utf8.ValidString(path) {
		return false
	}
	for _, r := range path {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return false
		}
	}
	return true
}

// ExpandHome expands only the current user's alias, never another user's name.
func ExpandHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") && !(runtime.GOOS == "windows" && strings.HasPrefix(path, `~\`)) {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	return home + string(os.PathSeparator) + path[2:], nil
}

func PathQuery(query string) bool {
	return strings.ContainsAny(query, `/\`) || strings.HasPrefix(query, "~") || (len(query) > 1 && query[1] == ':')
}

func NativeQuery(query string) bool {
	return !(runtime.GOOS != "windows" && (strings.ContainsRune(query, '\\') || (len(query) > 1 && query[1] == ':')))
}

func Complete(ctx context.Context, query string) Completion {
	if ctx.Err() != nil || !Safe(query) || !NativeQuery(query) {
		return Completion{}
	}
	// Foreign Windows syntax is a hint on Unix, never a relative filesystem query.
	if runtime.GOOS != "windows" && (strings.ContainsRune(query, '\\') || (len(query) > 1 && query[1] == ':')) {
		return Completion{}
	}
	input, err := ExpandHome(query)
	if err != nil {
		return Completion{}
	}
	if strings.HasPrefix(input, "~") {
		return Completion{}
	}
	original := input
	if runtime.GOOS == "windows" {
		input = strings.ReplaceAll(input, "/", `\`)
	}
	dir, prefix := filepath.Split(input)
	if dir == "" {
		dir = "."
	}
	resolvedParent := false
	for _, segment := range strings.Split(dir, string(os.PathSeparator)) {
		if segment == ".." {
			dir, err = filepath.EvalSymlinks(dir)
			if err != nil {
				return Completion{}
			}
			input = filepath.Join(dir, prefix)
			resolvedParent = true
			break
		}
	}
	if prefix == ".." {
		input, err = filepath.EvalSymlinks(input)
		if err != nil {
			return Completion{}
		}
		resolvedParent = true
	}
	if strings.HasSuffix(input, string(os.PathSeparator)) {
		dir = input
		prefix = ""
	} else if info, err := os.Stat(input); err == nil && info.IsDir() && prefix != "." {
		if ctx.Err() != nil {
			return Completion{}
		}
		if !resolvedParent {
			input = query
		}
		return Completion{Paths: []string{input}, Exact: true}
	}
	entries := readBounded(ctx, dir)
	result := Completion{}
	for _, entry := range entries {
		if ctx.Err() != nil {
			return Completion{}
		}
		name := entry.Name()
		if !entry.IsDir() || !Safe(name) || (strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".")) || !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			continue
		}
		match := filepath.Join(dir, name)
		if strings.HasPrefix(query, "~") {
			home, _ := os.UserHomeDir()
			rel, e := filepath.Rel(home, match)
			if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
				match = "~" + string(os.PathSeparator) + rel
			}
		} else if runtime.GOOS == "windows" && strings.Contains(original, "/") && !strings.Contains(original, `\`) {
			match = filepath.ToSlash(match)
		}
		if Safe(match) {
			result.Paths = append(result.Paths, match)
		}
		if len(result.Paths) == MaxResults {
			break
		}
	}
	return result
}

func readBounded(ctx context.Context, dir string) []os.DirEntry {
	if ctx.Err() != nil {
		return nil
	}
	f, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer f.Close()
	var entries []os.DirEntry
	for len(entries) < maxScan && ctx.Err() == nil {
		batch, err := f.ReadDir(min(512, maxScan-len(entries)))
		entries = append(entries, batch...)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil
			}
			break
		}
	}
	if ctx.Err() != nil {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	return entries
}

// Corpus order wins equal fuzzy scores and duplicates; filesystem is appended.
func Filter(corpus []string, query string, disk []string) []string {
	var candidates, lower []string
	for _, s := range corpus {
		if Safe(s) {
			candidates = append(candidates, s)
			lower = append(lower, strings.ToLower(s))
		}
	}
	query = strings.TrimSpace(query)
	var ordered []string
	if query == "" {
		ordered = candidates
	} else {
		for _, m := range fuzzy.Find(strings.ToLower(query), lower) {
			ordered = append(ordered, candidates[m.Index])
		}
	}
	seen := map[string]bool{}
	var result []string
	for _, s := range append(ordered, disk...) {
		key := s
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if !Safe(s) || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, s)
		if len(result) == MaxResults {
			break
		}
	}
	return result
}
