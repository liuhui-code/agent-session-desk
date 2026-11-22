package store_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestOutputRecordSurvivesReopenWithoutBecomingConversationOrLiveness(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	v := sample()
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	capture := session.Preview{Text: "合成输出\nline 2", Truncated: true}
	saved, err := db.RecordOutput(ctx, v, capture)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == "" || saved.RunID != v.RunID || saved.SessionID != v.ID || saved.BackendID != v.BackendID || saved.Source != "screen_snapshot" || saved.Complete || !saved.Truncated || saved.RecordedAt.IsZero() || saved.Sequence == 0 {
		t.Fatalf("source or completeness changed: %#v", saved)
	}
	db.Close()
	db, err = store.Open(ctx, path, "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Output(ctx, v.RunID)
	if err != nil || !reflect.DeepEqual(rows, []session.OutputRecord{saved}) || rows[0].Text != capture.Text {
		t.Fatalf("saved output changed: %#v, %v", rows, err)
	}
	history, err := db.History(ctx, 100)
	if err != nil || len(history) != 1 || history[0].Session.TerminalState != session.TerminalUnknown || history[0].StopConfirmedAt != nil {
		t.Fatalf("output invented lifecycle state: %#v %v", history, err)
	}
}

func TestOutputRecordBoundsTextAndKeepsUTF8WithExplicitTruncation(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := sample()
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	text := strings.Repeat("中🙂\xff", 20000) + "TAIL_MARKER"
	got, err := db.RecordOutput(ctx, v, session.Preview{Text: text})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Text) > 64*1024 || !utf8.ValidString(got.Text) || !got.Truncated || !strings.HasSuffix(got.Text, "TAIL_MARKER") || got.Complete {
		t.Fatalf("capture unbounded or stronger than its source: bytes=%d truncated=%v valid=%v", len(got.Text), got.Truncated, utf8.ValidString(got.Text))
	}
	rows, err := db.Output(ctx, v.RunID)
	if err != nil || len(rows) != 1 || rows[0].Text != got.Text {
		t.Fatal("bounded text changed on disk")
	}
}

func TestOutputRetentionKeepsTwoPerScopedRunWithoutMixingIdentities(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	a, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := store.Open(ctx, path, "b")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	v := sample()
	for _, db := range []*store.Store{a, b} {
		if err = db.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	var latest session.OutputRecord
	for i := 0; i < 6; i++ {
		latest, err = a.RecordOutput(ctx, v, session.Preview{Text: fmt.Sprint("a-", i)})
		if err != nil {
			t.Fatal(err)
		}
	}
	foreign, err := b.RecordOutput(ctx, v, session.Preview{Text: "scope-b"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := a.Output(ctx, v.RunID)
	if err != nil || len(got) != 2 || got[0].ID != latest.ID || got[0].Text != "a-5" || got[1].Text != "a-4" {
		t.Fatalf("per-Run retention changed or unbounded: %#v %v", got, err)
	}
	other, err := b.Output(ctx, v.RunID)
	if err != nil || len(other) != 1 || other[0].ID != foreign.ID || other[0].Text != "scope-b" {
		t.Fatalf("scope mixed: %#v %v", other, err)
	}
	if got[0].Sequence <= got[1].Sequence || foreign.Sequence <= latest.Sequence {
		t.Fatal("retained ordering is not causal")
	}
}

func TestOutputDoesNotAdoptAnUnrecognizedDirectoryOrDeleteItsFiles(t *testing.T) {
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
	root := filepath.Join(dir, "output-records")
	if err = os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(root, "unrelated.txt")
	if err = os.WriteFile(sentinel, []byte("private sentinel"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = db.RecordOutput(ctx, v, session.Preview{Text: "capture"}); err == nil {
		t.Fatal("silently adopted a directory without product ownership marker")
	}
	data, err := os.ReadFile(sentinel)
	if err != nil || string(data) != "private sentinel" {
		t.Fatal("unrelated file changed")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 {
		t.Fatal("unrecognized directory mutated")
	}
	if _, err = db.Output(ctx, v.RunID); err == nil {
		t.Fatal("reader accepted unrecognized output directory")
	}
}

func TestOutputGlobalRetentionBoundsPublishedFilesAcrossRuns(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := store.Open(ctx, filepath.Join(dir, "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var first, last string
	for i := 0; i < 40; i++ {
		v := sample()
		v.RunID = fmt.Sprintf("%032x", i+100)
		v.BackendID = fmt.Sprint("asd_", i)
		if err = db.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
		if _, err = db.RecordOutput(ctx, v, session.Preview{Text: strings.Repeat("<&\x00", 21000)}); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = v.RunID
		}
		last = v.RunID
	}
	entries, err := os.ReadDir(filepath.Join(dir, "output-records"))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			count++
			info, err := entry.Info()
			if err != nil || info.Size() > 512*1024 {
				t.Fatal("output file exceeds byte cap")
			}
		}
	}
	if count != 32 {
		t.Fatalf("published output files=%d, want 32", count)
	}
	older, err := db.Output(ctx, first)
	if err != nil || len(older) != 0 {
		t.Fatal("oldest retained record not evicted")
	}
	latest, err := db.Output(ctx, last)
	if err != nil || len(latest) != 1 || latest[0].RunID != last {
		t.Fatal("latest capture lost during global retention")
	}
}
