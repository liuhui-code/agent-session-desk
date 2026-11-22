package store_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestConcurrentConnectionsKeepAllCommittedIdentities(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	const n = 4
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			db, err := store.Open(ctx, path, "scope")
			if err != nil {
				errs <- err
				return
			}
			defer db.Close()
			v := sample()
			v.ID = fmt.Sprintf("%032x", i+1)
			v.RunID = fmt.Sprintf("%032x", i+100)
			v.BackendID = fmt.Sprint("asd_", i)
			errs <- db.Save(ctx, v)
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	db, err := store.Open(ctx, path, "scope")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.History(ctx, 100)
	if err != nil || len(rows) != n {
		t.Fatalf("lost commits: %d %v", len(rows), err)
	}
	rows, err = db.History(ctx, 2)
	if err != nil || len(rows) != 2 {
		t.Fatalf("limit ignored: %d %v", len(rows), err)
	}
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	for query, want := range map[string]string{"PRAGMA integrity_check": "ok", "SELECT sqlite_version()": "3.54.1", "PRAGMA journal_mode": "delete"} {
		var value string
		if err = raw.QueryRow(query).Scan(&value); err != nil || value != want {
			t.Fatalf("%s = %s %v", query, value, err)
		}
	}
}
