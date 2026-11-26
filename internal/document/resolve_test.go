package document

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMarkdownResolveLocalExistingFileOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "计划 & notes.md")
	if err := os.WriteFile(path, []byte("# local fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, ref := range []string{path, filepath.Base(path), filepath.Base(path) + ":12:3", filepath.Base(path) + "#intro"} {
		got, err := Resolve(ref, dir)
		if err != nil || got != canonical {
			t.Fatalf("%q: %q %v", ref, got, err)
		}
	}
	for _, ref := range []string{"missing.md", "script.md.exe", "https://example.com/a.md", "file://server/share/a.md", "//server/share/a.md", "/bad\x00.md"} {
		if _, err := Resolve(ref, dir); err == nil {
			t.Fatalf("accepted %q", ref)
		}
	}
	if _, err := Resolve(filepath.Base(path), ""); err == nil {
		t.Fatal("relative path used workbench cwd")
	}
	if err := os.Mkdir(filepath.Join(dir, "folder.md"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve("folder.md", dir); err == nil {
		t.Fatal("directory opened as document")
	}
}
