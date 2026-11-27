package store_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/store"
)

// The child is a real SQLite writer killed while its transaction is open.
// This covers process interruption on the test filesystem, not power loss.
func TestCommittedMetadataSurvivesKilledWriter(t *testing.T) {
	if path := os.Getenv("ASD_TEST_CRASH_DB"); path != "" {
		raw, err := sql.Open("sqlite", path)
		if err != nil {
			t.Fatal(err)
		}
		raw.SetMaxOpenConns(1)
		if _, err = raw.Exec("PRAGMA cache_size=1; PRAGMA synchronous=FULL; BEGIN IMMEDIATE;"); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 300; i++ {
			id := fmt.Sprintf("%032x", i+1000)
			if _, err = raw.Exec("INSERT INTO sessions VALUES(?,?,?,?)", "scope-a", id, strings.Repeat("x", 4000), "synthetic.exe"); err != nil {
				t.Fatal(err)
			}
			if _, err = raw.Exec("INSERT INTO runs(scope,id,session_id,backend_id,recorded_ns,stopped_ns) VALUES(?,?,?,?,?,NULL)", "scope-a", fmt.Sprintf("%032x", i+2000), id, "asd_partial", 1234); err != nil {
				t.Fatal(err)
			}
		}
		if err = os.WriteFile(path+".ready", []byte("uncommitted"), 0600); err != nil {
			t.Fatal(err)
		}
		for {
			time.Sleep(time.Second)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Save(ctx, exited()); err != nil {
		t.Fatal(err)
	}
	if err = db.MarkStopped(ctx, exited()); err != nil {
		t.Fatal(err)
	}
	before, err := db.History(ctx, 1000)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCommittedMetadataSurvivesKilledWriter$")
	cmd.Env = append(os.Environ(), "ASD_TEST_CRASH_DB="+path)
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cmd.Process.Kill(); cmd.Wait() }()
	ready := false
	for ctx.Err() == nil {
		if _, err = os.Stat(path + ".ready"); err == nil {
			ready = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ready {
		t.Fatal("writer did not reach its uncommitted checkpoint")
	}
	if err = cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err = cmd.Wait(); err == nil {
		t.Fatal("writer was not interrupted")
	}
	db, err = store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	after, err := db.History(ctx, 1000)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("committed evidence changed or partial rows survived: %d rows, %v", len(after), err)
	}
}
