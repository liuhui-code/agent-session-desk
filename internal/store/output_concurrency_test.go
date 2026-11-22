package store_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestOutputWritersShareTheDatabaseLockAndRetainScopedOrder(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.sqlite3")
	seed, err := store.Open(ctx, path, "seed")
	if err != nil {
		t.Fatal(err)
	}
	seed.Close()
	var wg sync.WaitGroup
	result := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			db, err := store.Open(ctx, path, fmt.Sprint("scope-", i))
			if err != nil {
				result <- err
				return
			}
			defer db.Close()
			v := sample()
			if err = db.Save(ctx, v); err != nil {
				result <- err
				return
			}
			for j := 0; j < 6; j++ {
				if _, err = db.RecordOutput(ctx, v, session.Preview{Text: fmt.Sprintf("%d:%d", i, j)}); err != nil {
					result <- err
					return
				}
			}
			result <- nil
		}(i)
	}
	wg.Wait()
	close(result)
	for err := range result {
		if err != nil {
			t.Fatal(err)
		}
	}
	sequences := map[uint64]bool{}
	for i := 0; i < 4; i++ {
		db, err := store.Open(ctx, path, fmt.Sprint("scope-", i))
		if err != nil {
			t.Fatal(err)
		}
		rows, err := db.Output(ctx, sample().RunID)
		db.Close()
		if err != nil || len(rows) != 2 || rows[0].Text != fmt.Sprintf("%d:5", i) || rows[1].Text != fmt.Sprintf("%d:4", i) {
			t.Fatalf("concurrent scope/order: %#v %v", rows, err)
		}
		for _, r := range rows {
			if sequences[r.Sequence] {
				t.Fatal("concurrent captures reused ordering identity")
			}
			sequences[r.Sequence] = true
		}
	}
	entries, err := os.ReadDir(filepath.Join(dir, "output-records"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 9 {
		t.Fatalf("retained files or temporary cleanup changed: %d", len(entries))
	}
}

func TestOutputClearsOnlyRecognizedOrphanStageUnderDatabaseLock(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := store.Open(ctx, filepath.Join(dir, "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := sample()
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	if _, err = db.RecordOutput(ctx, v, session.Preview{Text: "first"}); err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(dir, "output-records", ".stage-33333333333333333333333333333333.tmp")
	if err = os.WriteFile(stage, []byte("partial synthetic capture"), 0600); err != nil {
		t.Fatal(err)
	}
	// Reading skips unfinished work without promoting it into a saved capture.
	rows, err := db.Output(ctx, v.RunID)
	if err != nil || len(rows) != 1 {
		t.Fatalf("orphan became output: %#v %v", rows, err)
	}
	if _, err = db.RecordOutput(ctx, v, session.Preview{Text: "next"}); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(stage); !os.IsNotExist(err) {
		t.Fatal("old stage survived the next serialized write")
	}
}
