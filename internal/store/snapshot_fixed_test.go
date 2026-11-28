package store_test

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/store"
	"os"
	"path/filepath"
	"testing"
)

func TestFixedSnapshotNeverReusesOccupiedCapturePath(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	source := filepath.Join(dir, "source.sqlite3")
	db, err := store.Open(ctx, source, "fixed")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	target, stage := filepath.Join(dir, "backup.sqlite3"), filepath.Join(dir, "capture.tmp")
	os.WriteFile(stage, []byte("keep incomplete capture"), 0600)
	if err = store.SnapshotFixed(ctx, source, target, stage); err == nil {
		t.Fatal("occupied capture reused")
	}
	b, _ := os.ReadFile(stage)
	if string(b) != "keep incomplete capture" {
		t.Fatal("occupied capture overwritten")
	}
	os.Remove(stage)
	if err = store.SnapshotFixed(ctx, source, target, stage); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Check(ctx, target); err != nil {
		t.Fatal(err)
	}
}
