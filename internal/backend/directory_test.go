package backend

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMarkdownDirectoryUsesOriginalPaneAndRejectsUnknown(t *testing.T) {
	b := newFake(t)
	target := startFake(t, b)
	directory := filepath.Join(t.TempDir(), "中文 folder")
	t.Setenv("ASD_TEST_PANE_DIRECTORY", directory)
	got, err := b.CurrentDirectory(context.Background(), target)
	if err != nil || got != directory {
		t.Fatalf("%q %v", got, err)
	}
	for _, bad := range []string{"", "relative/path", "first\nsecond"} {
		t.Setenv("ASD_TEST_PANE_DIRECTORY", bad)
		if _, err := b.CurrentDirectory(context.Background(), target); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
	target.RunID = "changed"
	if _, err := b.CurrentDirectory(context.Background(), target); err == nil {
		t.Fatal("changed Run accepted")
	}
}
