//go:build windows

package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

// Explicit fixture entry only. It seeds bounded synthetic reminder metadata,
// not 255 genuine agent prompts; never exported as a product command.
func TestWindowsAttentionCapacityFixtureHelper(t *testing.T) {
	file := os.Getenv("ASD_ATTENTION_FIXTURE_INPUT")
	if file == "" {
		t.Skip("explicit private capacity fixture required")
	}
	var request struct {
		Root, Database, Scope, CredentialFile, Output string
		Count                                         int
	}
	data, err := os.ReadFile(file)
	if err != nil || len(data) > 16384 {
		t.Fatal("private fixture request unavailable")
	}
	if json.Unmarshal(data, &request) != nil || request.Count < 0 || request.Count > 255 {
		t.Fatal("invalid capacity fixture")
	}
	// Every input/output is confined to a caller-created marker directory.
	if _, err = os.Stat(filepath.Join(request.Root, "attention-fixture.marker")); err != nil {
		t.Fatal("private fixture marker required")
	}
	for _, path := range []string{request.Database, request.CredentialFile, request.Output} {
		rel, e := filepath.Rel(request.Root, path)
		if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Fatal("fixture path outside private root")
		}
	}
	var doc struct {
		SessionID string `json:"session_id"`
		RunID     string `json:"run_id"`
		Token     string `json:"token"`
	}
	data, err = os.ReadFile(request.CredentialFile)
	if err != nil || json.Unmarshal(data, &doc) != nil {
		t.Fatal("private credential unavailable")
	}
	credential := session.AgentCredential{SessionID: doc.SessionID, RunID: doc.RunID, Secret: doc.Token}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	db, err := store.Open(ctx, request.Database, request.Scope)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for i := 0; i < request.Count; i++ {
		item, e := db.ReportAttention(ctx, credential, session.AttentionEvent{ID: fmt.Sprintf("%032x", 4096+i), Version: 1, Kind: "question", State: "pending", Title: fmt.Sprintf("Synthetic reminder %03d", i+1), Detail: strings.Repeat("d", 8192)})
		if e != nil {
			t.Fatal("capacity fixture report failed")
		}
		if _, e = db.SnoozeAttention(ctx, item, 30*time.Minute); e != nil {
			t.Fatal("capacity fixture snooze failed")
		}
	}
	items, err := db.PendingAttention(ctx)
	if err != nil || len(items) != request.Count+1 {
		t.Fatal("capacity fixture count mismatch")
	}
	data, _ = json.Marshal(map[string]any{"seeded": request.Count, "total_pending": len(items), "detail_bytes": 8192, "synthetic_metadata_only": true})
	if err = os.WriteFile(request.Output, data, 0600); err != nil {
		t.Fatal(err)
	}
}
