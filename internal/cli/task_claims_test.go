package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestClaimFileRetainsPrivateKeyAndOriginalIntent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "claim.json")
	id, msg := strings.Repeat("a", 32), strings.Repeat("b", 32)
	k, err := ensureTaskClaimFile(path, id, msg, 300)
	if err != nil {
		t.Fatal(err)
	}
	again, err := ensureTaskClaimFile(path, id, msg, 300)
	if err != nil || again != k {
		t.Fatal("claim retry rotated private key")
	}
	if _, err = ensureTaskClaimFile(path, id, msg, 301); err == nil {
		t.Fatal("changed lease intent accepted")
	}
	if _, err = ensureTaskClaimFile(path, strings.Repeat("c", 32), msg, 300); err == nil {
		t.Fatal("another claimant overwrote private file")
	}
	data, err := json.Marshal(k)
	if err != nil || bytes.Contains(data, []byte(k.Secret)) {
		t.Fatal("public JSON leaked processing secret")
	}
	files, err := os.ReadDir(root)
	if err != nil || len(files) != 1 {
		t.Fatal("claim staging file remains")
	}
}

func TestClaimInvalidCommandsHaveNoEffects(t *testing.T) {
	root := t.TempDir()
	state, runtime := filepath.Join(root, "state"), filepath.Join(root, "runtime")
	for _, tail := range [][]string{{"claim"}, {"renew"}, {"claim-status", "name"}, {"claim", "--claim-id", "name", "--message-id", strings.Repeat("a", 32), "--token-file", "missing", "--claim-file", filepath.Join(root, "new"), "--lease-seconds", "0"}} {
		var out, stderr bytes.Buffer
		args := append([]string{"--backend", filepath.Join(root, "missing"), "--state-dir", state, "--runtime-dir", runtime}, tail...)
		if code := Run(context.Background(), args, nil, &out, &stderr); code == 0 {
			t.Fatal("invalid claim command accepted")
		}
	}
	for _, path := range []string{state, runtime, filepath.Join(root, "new")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("invalid claim changed state")
		}
	}
}

func TestClaimStatusIsOfflineAndDoesNotExposePrivateKey(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	state := filepath.Join(root, "state")
	opts := backend.Options{Binary: filepath.Join(root, "missing"), RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-claim"}
	db, err := openMetadata(ctx, opts, state)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := session.Session{ID: strings.Repeat("1", 32), RunID: strings.Repeat("2", 32), BackendID: "original", Command: "agent.exe"}
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	c, err := session.NewAgentCredential(v.ID, v.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.BindAgent(ctx, c, v); err != nil {
		t.Fatal(err)
	}
	msg := session.Message{ID: strings.Repeat("a", 32), Target: v, Body: "task", Fingerprint: strings.Repeat("b", 64)}
	if _, err = db.QueueMessage(ctx, msg); err != nil {
		t.Fatal(err)
	}
	args := []string{"--backend", opts.Binary, "--runtime-dir", opts.RuntimeDir, "--namespace", opts.Namespace, "--state-dir", state, "claim-status", msg.ID}
	call := func() []byte {
		var out, stderr bytes.Buffer
		if code := Run(ctx, args, nil, &out, &stderr); code != 0 {
			t.Fatalf("offline claim status: %s", &stderr)
		}
		return out.Bytes()
	}
	var doc struct {
		MessageID string             `json:"message_id"`
		Claim     *session.TaskClaim `json:"claim"`
	}
	if err = json.Unmarshal(call(), &doc); err != nil || doc.Claim != nil || doc.MessageID != msg.ID {
		t.Fatal("unclaimed status not explicit null")
	}
	k, err := ensureTaskClaimFile(filepath.Join(root, "claim.json"), strings.Repeat("c", 32), msg.ID, 300)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.ClaimTask(ctx, c, k, 300); err != nil {
		t.Fatal(err)
	}
	data := call()
	if err = json.Unmarshal(data, &doc); err != nil || doc.Claim == nil || doc.Claim.NewlyClaimed || bytes.Contains(data, []byte(k.Secret)) {
		t.Fatal("status exposed permission or secret")
	}
	if _, err = os.Stat(opts.RuntimeDir); !os.IsNotExist(err) {
		t.Fatal("offline status created backend runtime")
	}
}
