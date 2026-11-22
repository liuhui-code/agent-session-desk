package store_test

import (
	"context"
	"encoding/json"
	"fmt"
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
				data, e := os.ReadFile(filepath.Join(root, entry.Name()))
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
