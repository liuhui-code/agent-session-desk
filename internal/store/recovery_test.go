package store_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestSnapshotRetainsAllScopedEvidenceWithoutNeedingBackend(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	source := filepath.Join(dir, "metadata.sqlite3")
	db, err := store.Open(ctx, source, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Save(ctx, exited()); err != nil {
		t.Fatal(err)
	}
	if err = db.MarkStopped(ctx, exited()); err != nil {
		t.Fatal(err)
	}
	before, err := db.History(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	other, err := store.Open(ctx, source, "scope-b")
	if err != nil {
		t.Fatal(err)
	}
	if err = other.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	other.Close()
	backup := filepath.Join(dir, "备份 '#%.sqlite3")
	if err = store.Snapshot(ctx, source, backup); err != nil {
		t.Fatal(err)
	}
	restored, err := store.Open(ctx, backup, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	after, err := restored.History(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("snapshot changed evidence: %#v -> %#v", before, after)
	}
	second, err := store.Open(ctx, backup, "scope-b")
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	rows, err := second.History(ctx, 100)
	if err != nil || len(rows) != 1 {
		t.Fatalf("snapshot lost another scope: %#v %v", rows, err)
	}
}

func TestSnapshotRejectsInvalidDatabaseAndPreservesBothPaths(t *testing.T) {
	for _, kind := range []string{"foreign", "future", "corrupt", "missing-table"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			dir := t.TempDir()
			source := filepath.Join(dir, "source.sqlite3")
			target := filepath.Join(dir, "backup.sqlite3")
			if kind == "corrupt" {
				if err := os.WriteFile(source, []byte("synthetic damaged database"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				db, err := store.Open(ctx, source, "a")
				if err != nil {
					t.Fatal(err)
				}
				db.Close()
				raw, err := sql.Open("sqlite", source)
				if err != nil {
					t.Fatal(err)
				}
				query := "PRAGMA application_id=123"
				if kind == "future" {
					query = "PRAGMA user_version=99"
				}
				if kind == "missing-table" {
					query = "DROP TABLE runs"
				}
				if _, err = raw.Exec(query); err != nil {
					t.Fatal(err)
				}
				raw.Close()
			}
			before, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			if err = store.Snapshot(ctx, source, target); err == nil {
				t.Fatal("invalid database accepted as a product backup")
			}
			after, err := os.ReadFile(source)
			if err != nil || sha256.Sum256(after) != sha256.Sum256(before) {
				t.Fatal("invalid source changed")
			}
			if _, err = os.Stat(target); !os.IsNotExist(err) {
				t.Fatal("failed backup left a destination")
			}
		})
	}
}

func TestCheckIsReadOnlyAndDoesNotMigrateV1OrCreateMissingSource(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "old metadata #%.sqlite3")
	createV1(t, path)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	report, err := store.Check(ctx, path)
	if err != nil || report.SchemaVersion != 1 || report.Integrity != "ok" || report.ForeignKeys != "ok" {
		t.Fatalf("check: %#v %v", report, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("checking migrated or changed source")
	}
	missing := filepath.Join(dir, "missing.sqlite3")
	if _, err = store.Check(ctx, missing); err == nil {
		t.Fatal("missing database reported healthy")
	}
	if _, err = os.Stat(missing); !os.IsNotExist(err) {
		t.Fatal("check created a database")
	}
}

func TestSnapshotNeverOverwritesAndCancellationLeavesNoOutput(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	source := filepath.Join(dir, "metadata.sqlite3")
	db, err := store.Open(ctx, source, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	for _, content := range []string{"", "existing backup sentinel"} {
		target := filepath.Join(t.TempDir(), "existing.sqlite3")
		if err = os.WriteFile(target, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if err = store.Snapshot(ctx, source, target); err == nil {
			t.Fatal("existing destination accepted")
		}
		data, err := os.ReadFile(target)
		if err != nil || string(data) != content {
			t.Fatal("existing destination changed")
		}
	}
	target := filepath.Join(dir, "cancelled.sqlite3")
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err = store.Snapshot(cancelled, source, target); err == nil {
		t.Fatal("cancelled backup reported success")
	}
	if _, err = os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("cancelled backup left output")
	}
}

func TestCheckAndSnapshotRejectBrokenForeignKeyEvidence(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	source := filepath.Join(dir, "metadata.sqlite3")
	db, err := store.Open(ctx, source, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	db.Close()
	raw, err := sql.Open("sqlite", source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = raw.Exec("PRAGMA foreign_keys=OFF; DELETE FROM sessions;"); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	before, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Check(ctx, source); err == nil {
		t.Fatal("orphan Run declared healthy")
	}
	if err = store.Snapshot(ctx, source, filepath.Join(dir, "backup.sqlite3")); err == nil {
		t.Fatal("orphan Run accepted for backup")
	}
	after, err := os.ReadFile(source)
	if err != nil || sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("broken evidence modified")
	}
}

func TestV1SnapshotRestoresThroughNormalMigrationWithoutChangingBackup(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	backup := filepath.Join(dir, "v1.sqlite3")
	target := filepath.Join(dir, "restored.sqlite3")
	createV1(t, backup)
	before, err := os.ReadFile(backup)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Snapshot(ctx, backup, target); err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(ctx, target, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.History(ctx, 100)
	db.Close()
	if err != nil || len(rows) != 1 || rows[0].RecordedAt.UnixNano() != 1234 || rows[0].StopConfirmedAt == nil || rows[0].StopConfirmedAt.UnixNano() != 5678 {
		t.Fatalf("restored v1 changed: %#v %v", rows, err)
	}
	result, err := store.Check(ctx, target)
	if err != nil || result.SchemaVersion != 11 {
		t.Fatalf("migration: %#v %v", result, err)
	}
	after, err := os.ReadFile(backup)
	if err != nil || sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("restoring upgraded backup source")
	}
}

func TestSnapshotDestinationIsNeverPublishedAsAnIncompleteFile(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	source := filepath.Join(dir, "source.sqlite3")
	target := filepath.Join(dir, "published.sqlite3")
	db, err := store.Open(ctx, source, "a")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 400; i++ {
		v := sample()
		v.ID = fmt.Sprintf("%032x", i+1)
		v.RunID = fmt.Sprintf("%032x", i+1000)
		v.Title = strings.Repeat("x", 4000)
		if err = db.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	db.Close()
	result := make(chan error, 1)
	go func() { result <- store.Snapshot(ctx, source, target) }()
	partial := false
	for {
		select {
		case err = <-result:
			if err != nil {
				t.Fatal(err)
			}
			if partial {
				t.Fatal("backup destination was visible before it was a complete database")
			}
			if _, err = store.Check(ctx, target); err != nil {
				t.Fatal(err)
			}
			return
		default:
			info, e := os.Stat(target)
			if e == nil && info.Size() < 100 {
				partial = true
			}
			runtime.Gosched()
		}
	}
}
