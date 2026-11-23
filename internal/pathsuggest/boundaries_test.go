package pathsuggest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestHiddenControlUnicodeAndResultBounds(t *testing.T) {
	root := t.TempDir()
	for _, n := range []string{".hidden", "中文 空格", "normal"} {
		if err := os.Mkdir(filepath.Join(root, n), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 25; i++ {
		if err := os.Mkdir(filepath.Join(root, fmt.Sprintf("project-%02d", i)), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if runtime.GOOS != "windows" {
		_ = os.Mkdir(filepath.Join(root, "bad\nname"), 0700)
	}
	all := Complete(context.Background(), root+string(os.PathSeparator))
	if len(all.Paths) != MaxResults {
		t.Fatal("filesystem cap not applied")
	}
	for _, s := range all.Paths {
		if strings.Contains(s, ".hidden") || !Safe(s) {
			t.Fatal("hidden/control entry leaked")
		}
	}
	hidden := Complete(context.Background(), root+string(os.PathSeparator)+".")
	if len(hidden.Paths) != 1 || !strings.HasSuffix(hidden.Paths[0], ".hidden") {
		t.Fatalf("dot prefix: %v", hidden)
	}
	unicode := Complete(context.Background(), root+string(os.PathSeparator)+"中")
	if len(unicode.Paths) != 1 {
		t.Fatalf("unicode completion: %v", unicode)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if len(Complete(ctx, root+string(os.PathSeparator)).Paths) != 0 {
		t.Fatal("cancelled scan")
	}
	var corpus []string
	for i := 0; i < 50; i++ {
		corpus = append(corpus, fmt.Sprintf("/project/%d", i))
	}
	if len(Filter(corpus, "", corpus)) != MaxResults {
		t.Fatal("corpus cap")
	}
	if got := Filter([]string{"/same", "/same"}, "", []string{"/same", "/other"}); len(got) != 2 {
		t.Fatal("merge duplicates")
	}
	if got := Complete(context.Background(), "not-a-path"); len(got.Paths) != 0 {
		t.Fatal("bare query scanned cwd")
	}
	if runtime.GOOS != "windows" && len(Complete(context.Background(), `C:\other\dir`).Paths) != 0 {
		t.Fatal("foreign path scanned cwd")
	}
}
func TestSymlinkParentTraversalUsesFilesystemMeaning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink privilege-dependent; Windows junction gate remains separate")
	}
	root := t.TempDir()
	target := filepath.Join(root, "other", "child")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "other", "wanted"), 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	got := Complete(context.Background(), link+"/../wa")
	want, err := filepath.EvalSymlinks(filepath.Join(root, "other", "wanted"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Paths) != 1 || got.Paths[0] != want {
		t.Fatalf("lexical clean changed traversal: %v", got)
	}
}
