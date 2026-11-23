package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestOutputPublicationNeverExposesIncompleteOrExcessFiles(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := store.Open(ctx, filepath.Join(dir, "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for i := 0; i < 32; i++ {
		v := sample()
		v.RunID = fmt.Sprintf("%032x", i+100)
		if err = db.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
		if _, err = db.RecordOutput(ctx, v, session.Preview{Text: "seed"}); err != nil {
			t.Fatal(err)
		}
	}
	result := make(chan error, 1)
	go func() {
		for i := 0; i < 20; i++ {
			v := sample()
			v.RunID = fmt.Sprintf("%032x", i+1000)
			if err := db.Save(ctx, v); err != nil {
				result <- err
				return
			}
			if _, err := db.RecordOutput(ctx, v, session.Preview{Text: strings.Repeat("\x00", 64*1024)}); err != nil {
				result <- err
				return
			}
		}
		result <- nil
	}()
	root := filepath.Join(dir, "output-records")
	partial, excess := false, false
	for {
		select {
		case err = <-result:
			if err != nil {
				t.Fatal(err)
			}
			if partial || excess {
				t.Fatalf("published partial=%v or excess=%v output files", partial, excess)
			}
			return
		default:
			entries, e := os.ReadDir(root)
			if e != nil {
				t.Fatal(e)
			}
			count := 0
			for _, entry := range entries {
				if !strings.HasSuffix(entry.Name(), ".json") {
					continue
				}
				count++
				observer, e := openOutputObserver(filepath.Join(root, entry.Name()))
				if e != nil {
					continue
				} // Rotation may remove an entry after ReadDir.
				data, e := io.ReadAll(observer)
				closeErr := observer.Close()
				if e == nil {
					e = closeErr
				}
				if e == nil && !json.Valid(data) {
					partial = true
				}
			}
			if count > 32 {
				excess = true
			}
			runtime.Gosched()
		}
	}
}

func TestOutputObserverDoesNotPreventPublishedRecordRotation(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := store.Open(ctx, filepath.Join(dir, "metadata.sqlite3"), "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Save(ctx, sample()); err != nil {
		t.Fatal(err)
	}
	first, err := db.RecordOutput(ctx, sample(), session.Preview{Text: "held observer"})
	if err != nil {
		t.Fatal(err)
	}
	observer, err := openOutputObserver(filepath.Join(dir, "output-records", first.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	defer observer.Close()
	for i := 0; i < 2; i++ {
		if _, err = db.RecordOutput(ctx, sample(), session.Preview{Text: "next"}); err != nil {
			t.Fatalf("independent observer prevented pruning: %v", err)
		}
	}
	data, err := io.ReadAll(observer)
	if err != nil || !json.Valid(data) {
		t.Fatalf("held observer lost complete published bytes: %v", err)
	}
	records, err := db.Output(ctx, sample().RunID)
	if err != nil || len(records) != 2 {
		t.Fatalf("retention changed: %+v %v", records, err)
	}
}
