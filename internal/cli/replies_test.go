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

func TestReplyCredentialFilePublicationIsPrivateAndRepeatable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credential.json")
	sid, rid := strings.Repeat("1", 32), strings.Repeat("2", 32)
	c, err := ensureAgentCredential(path, sid, rid)
	if err != nil {
		t.Fatal(err)
	}
	again, err := ensureAgentCredential(path, sid, rid)
	if err != nil || c != again {
		t.Fatal("retry rotated a credential")
	}
	data, err := json.Marshal(c)
	if err != nil || bytes.Contains(data, []byte(c.Secret)) {
		t.Fatal("public credential JSON leaked token")
	}
	if _, err = ensureAgentCredential(path, sid, strings.Repeat("3", 32)); err == nil {
		t.Fatal("file was rebound")
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 1 {
		t.Fatal("credential staging files remain")
	}
	sentinel := filepath.Join(dir, "sentinel")
	if err = os.WriteFile(sentinel, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ensureAgentCredential(sentinel, sid, rid); err == nil {
		t.Fatal("unrelated file accepted")
	}
	data, _ = os.ReadFile(sentinel)
	if string(data) != "keep" {
		t.Fatal("credential publication overwrote existing file")
	}
}

func TestReplyInvalidArgumentsHaveNoStateOrBackendEffects(t *testing.T) {
	root := t.TempDir()
	runtime, state := filepath.Join(root, "runtime"), filepath.Join(root, "state")
	credential := filepath.Join(root, "credential.json")
	if _, err := ensureAgentCredential(credential, strings.Repeat("1", 32), strings.Repeat("2", 32)); err != nil {
		t.Fatal(err)
	}
	for _, tail := range [][]string{
		{"replies", "name"}, {"inbox"}, {"agent-bind", "--session", "name", "--run", strings.Repeat("2", 32), "--token-file", filepath.Join(root, "unused")},
		{"reply", "--token-file", credential, "--reply-id", strings.Repeat("a", 32), "--reply-to", strings.Repeat("b", 32), "--body-file", "-"},
		{"reply", "--from", strings.Repeat("1", 32)},
	} {
		var out, stderr bytes.Buffer
		args := append([]string{"--backend", filepath.Join(root, "missing"), "--runtime-dir", runtime, "--state-dir", state}, tail...)
		if code := Run(context.Background(), args, strings.NewReader("bad\x1bbody"), &out, &stderr); code == 0 {
			t.Fatal("invalid reply command accepted")
		}
	}
	for _, path := range []string{runtime, state, filepath.Join(root, "unused")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("invalid arguments changed state")
		}
	}
}

func TestRepliesQueryIsOfflineAndScopeIsolated(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	state := filepath.Join(root, "state")
	opts := backend.Options{Binary: filepath.Join(root, "missing"), RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-reply"}
	db, err := openMetadata(ctx, opts, state)
	if err != nil {
		t.Fatal(err)
	}
	v := session.Session{ID: strings.Repeat("1", 32), RunID: strings.Repeat("2", 32), BackendID: "target", Command: "agent.exe"}
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
	msg := session.Message{ID: strings.Repeat("a", 32), Target: v, Body: "request", Fingerprint: strings.Repeat("b", 64)}
	if _, err = db.QueueMessage(ctx, msg); err != nil {
		t.Fatal(err)
	}
	claimSecret, e := session.NewAgentCredential(strings.Repeat("d", 32), msg.ID)
	if e != nil {
		t.Fatal(e)
	}
	key := session.TaskClaimCredential{ID: strings.Repeat("d", 32), MessageID: msg.ID, Secret: claimSecret.Secret}
	if _, e = db.ClaimTask(ctx, c, key, 300); e != nil {
		t.Fatal(e)
	}
	if _, err = db.RecordReply(ctx, c, session.ReplyRequest{ID: strings.Repeat("c", 32), ReplyTo: msg.ID, Body: "result", Claim: key}); err != nil {
		t.Fatal(err)
	}
	db.Close()
	args := []string{"--backend", opts.Binary, "--runtime-dir", opts.RuntimeDir, "--namespace", opts.Namespace, "--state-dir", state, "replies", msg.ID}
	var out, stderr bytes.Buffer
	if code := Run(ctx, args, nil, &out, &stderr); code != 0 {
		t.Fatalf("offline reply query: %s", &stderr)
	}
	var rows []session.Reply
	if err = json.Unmarshal(out.Bytes(), &rows); err != nil || len(rows) != 1 || rows[0].ReplyTo != msg.ID || bytes.Contains(out.Bytes(), []byte(c.Secret)) {
		t.Fatal("invalid public result JSON")
	}
	if _, err = os.Stat(opts.RuntimeDir); !os.IsNotExist(err) {
		t.Fatal("offline query created backend runtime")
	}
	args[5] = "asd-other"
	out.Reset()
	stderr.Reset()
	if code := Run(ctx, args, nil, &out, &stderr); code == 0 {
		t.Fatal("reply exposed across namespace")
	}
}
