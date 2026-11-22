package store_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestOutputRejectsUnexpectedOrUnboundedFilesWithoutDeletingThem(t *testing.T) {
	for _, kind := range []string{"unrelated", "file-budget", "text-budget", "complete-claim", "backend-conflict"} {
		t.Run(kind, func(t *testing.T) {
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
			record, err := db.RecordOutput(ctx, v, session.Preview{Text: "saved"})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "output-records", record.ID+".json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "unrelated":
				path = filepath.Join(dir, "output-records", "unrelated.txt")
				data = []byte("private sentinel")
			case "file-budget":
				data = append(data, []byte(strings.Repeat(" ", 512*1024))...)
			default:
				var file map[string]json.RawMessage
				if err = json.Unmarshal(data, &file); err != nil {
					t.Fatal(err)
				}
				if kind == "text-budget" {
					record.Text = strings.Repeat("x", 64*1024+1)
				}
				if kind == "complete-claim" {
					record.Complete = true
				}
				if kind == "backend-conflict" {
					record.BackendID = "other-backend"
				}
				file["record"], err = json.Marshal(record)
				if err != nil {
					t.Fatal(err)
				}
				data, err = json.Marshal(file)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err = os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			before := sha256.Sum256(data)
			if _, err = db.Output(ctx, v.RunID); err == nil {
				t.Fatal("reader accepted unexpected or stronger output")
			}
			if _, err = db.RecordOutput(ctx, v, session.Preview{Text: "next"}); err == nil {
				t.Fatal("writer accepted invalid output directory")
			}
			after, err := os.ReadFile(path)
			if err != nil || sha256.Sum256(after) != before {
				t.Fatal("rejected file was deleted or changed")
			}
		})
	}
}
