//go:build windows

package document

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMarkdownSystemEditorKeepsOneLiteralDocumentArgument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review & literal.md")
	if err := os.WriteFile(path, []byte("# fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	expected, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	launch := func(got string) error {
		calls++
		if got != expected {
			t.Fatal("file path changed")
		}
		return nil
	}
	if err := openSystemEditor(context.Background(), path, launch); err != nil || calls != 1 {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := openSystemEditor(ctx, path, launch); !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatal("cancelled request launched editor")
	}
	if err := openSystemEditor(context.Background(), path+".exe", launch); err == nil || calls != 1 {
		t.Fatal("executable or missing target launched editor")
	}
}
