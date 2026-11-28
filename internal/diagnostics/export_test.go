package diagnostics_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/diagnostics"
)

func TestConcurrentExportsPublishCompleteReportsWithinFourSlots(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "reports")
	var journal diagnostics.Journal
	report := journal.Snapshot(diagnostics.Runtime{}, "", "unknown", "not_open")
	var wg sync.WaitGroup
	paths := make(chan string, 16)
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			path, err := diagnostics.Export(context.Background(), dir, report)
			if err == nil {
				paths <- path
			}
		}()
	}
	wg.Wait()
	close(paths)
	count := 0
	for path := range paths {
		count++
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			t.Fatal("published report incomplete", err)
		}
	}
	if count != 4 {
		t.Fatalf("wanted four nonoverwriting exports, got %d", count)
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 4 {
		t.Fatal("temporary files leaked", len(files), err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = diagnostics.Export(ctx, filepath.Join(t.TempDir(), "cancelled"), report); err == nil {
		t.Fatal("cancelled export published")
	}
}
