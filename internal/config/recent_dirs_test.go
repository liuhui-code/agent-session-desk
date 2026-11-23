package config_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func TestRecentDirectoriesRememberLatestThenRestoreWithoutRewritingOnRead(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "recent-dirs.toml")
	a, b := filepath.Join(t.TempDir(), "a"), filepath.Join(t.TempDir(), "b")
	if err := config.RememberDirs(ctx, path, []string{a, b}); err != nil {
		t.Fatal(err)
	}
	if err := config.RememberDirs(ctx, path, []string{b}); err != nil {
		t.Fatal(err)
	}
	before, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	info, e := os.Stat(path)
	if e != nil {
		t.Fatal(e)
	}
	got, err := config.LoadRecentDirs(ctx, path)
	if err != nil || !reflect.DeepEqual(got, []string{b, a}) {
		t.Fatalf("recent directories not restored in latest-first order: %v %v", got, err)
	}
	after, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	changed, e := os.Stat(path)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(before, after) || !info.ModTime().Equal(changed.ModTime()) {
		t.Fatal("read rewrote cache")
	}
}

func TestRecentDirectoryCacheRejectsInvalidDataWithoutReplacingIt(t *testing.T) {
	for _, data := range []string{"schema_version=2\ndirectories=[]\n", "schema_version=1\nunknown=1\n", "schema_version=1\ndirectories=[\"bad\\npath\"]\n", strings.Repeat("#", 192*1024+1)} {
		t.Run(fmt.Sprint(len(data), data[:min(20, len(data))]), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "recent.toml")
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := config.LoadRecentDirs(context.Background(), path); err == nil {
				t.Fatal("invalid cache accepted")
			}
			if err := config.RememberDirs(context.Background(), path, []string{"/valid"}); err == nil {
				t.Fatal("bad cache overwritten")
			}
			b, err := os.ReadFile(path)
			if err != nil || string(b) != data {
				t.Fatal("cache changed")
			}
			tmp, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".recent-dirs-*.tmp"))
			if len(tmp) != 0 {
				t.Fatal("temporary file leaked")
			}
		})
	}
}
func TestRecentDirectoryCacheBoundsCancellationAndMissingRead(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	path := filepath.Join(root, "absent", "recent.toml")
	if got, err := config.LoadRecentDirs(ctx, path); err != nil || len(got) != 0 {
		t.Fatal(got, err)
	}
	if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatal("read created storage")
	}
	var dirs []string
	for i := 0; i < 20; i++ {
		dirs = append(dirs, fmt.Sprintf("/bounded/%d", i))
	}
	if err := config.RememberDirs(ctx, path, dirs); err != nil {
		t.Fatal(err)
	}
	if err := config.RememberDirs(ctx, path, []string{"/new"}); err != nil {
		t.Fatal(err)
	}
	got, err := config.LoadRecentDirs(ctx, path)
	if err != nil || len(got) != 20 || got[0] != "/new" || got[19] != "/bounded/18" {
		t.Fatal(got, err)
	}
	before, _ := os.ReadFile(path)
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := config.RememberDirs(cancelled, path, []string{"/cancelled"}); err == nil {
		t.Fatal("cancelled save accepted")
	}
	for _, bad := range [][]string{append(dirs, "/excess"), {""}, {strings.Repeat("x", 4097)}, {"line\nfeed"}, {string([]byte{0xff})}} {
		if err := config.RememberDirs(ctx, path, bad); err == nil {
			t.Fatal("invalid write accepted")
		}
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("rejected write changed cache")
	}
}
