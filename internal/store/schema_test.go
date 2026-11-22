package store_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestUnrecognizedOrCorruptDatabaseIsPreserved(t *testing.T) {
	for _, kind := range []string{"corrupt", "foreign", "future", "missing-table"} {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "metadata.sqlite3")
			ctx := context.Background()
			if kind == "corrupt" {
				if err := os.WriteFile(path, []byte("synthetic broken database"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				db, err := store.Open(ctx, path, "scope-a")
				if err != nil {
					t.Fatal(err)
				}
				db.Close()
				raw, err := sql.Open("sqlite", path)
				if err != nil {
					t.Fatal(err)
				}
				query := "PRAGMA application_id=123"
				if kind == "future" {
					query = "PRAGMA user_version=99"
				} else if kind == "missing-table" {
					query = "DROP TABLE runs"
				}
				if _, err = raw.Exec(query); err != nil {
					t.Fatal(err)
				}
				raw.Close()
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			db, err := store.Open(ctx, path, "scope-a")
			if err == nil {
				db.Close()
				t.Fatal("unrecognized database accepted")
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if sha256.Sum256(before) != sha256.Sum256(after) {
				t.Fatal("rejected database changed")
			}
		})
	}
}

func TestBusyPreflightFailsWithinBoundWithoutChangingHistory(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if _, err = raw.Exec("BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	defer raw.Exec("ROLLBACK")
	at := time.Now()
	if err = db.CheckWritable(ctx); err == nil {
		t.Fatal("locked write preflight succeeded")
	}
	if time.Since(at) > 4*time.Second {
		t.Fatal("lock wait was not bounded")
	}
	rows, err := db.History(ctx, 100)
	if err != nil || len(rows) != 1 {
		t.Fatalf("committed metadata changed: %#v %v", rows, err)
	}
}

func TestRepeatedDiscoveryDoesNotNeedWriterLock(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if _, err = raw.Exec("BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	defer raw.Exec("ROLLBACK")
	if err = db.Save(ctx, sample()); err != nil {
		t.Fatalf("unchanged discovery took a writer lock: %v", err)
	}
}
