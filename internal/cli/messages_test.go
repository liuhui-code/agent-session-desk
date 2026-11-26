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

func TestMessagingStatusIsOfflineAndScopeIsolated(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	opts := backend.Options{Binary: filepath.Join(root, "missing"), RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-a"}
	state := filepath.Join(root, "state")
	db, e := openMetadata(ctx, opts, state)
	if e != nil {
		t.Fatal(e)
	}
	target := session.Session{ID: strings.Repeat("1", 32), RunID: strings.Repeat("2", 32), BackendID: "original", Command: "shell"}
	if e = db.Save(ctx, target); e != nil {
		t.Fatal(e)
	}
	request := session.MessageRequest{RequestID: strings.Repeat("a", 32), TargetSessionID: target.ID, TargetRunID: target.RunID, Body: "task"}
	fp, _ := request.Fingerprint()
	if _, e = db.QueueMessage(ctx, session.Message{ID: request.RequestID, Target: target, Body: request.Body, Fingerprint: fp}); e != nil {
		t.Fatal(e)
	}
	db.Close()
	args := []string{"--backend", opts.Binary, "--runtime-dir", opts.RuntimeDir, "--namespace", opts.Namespace, "--state-dir", state, "status", request.RequestID}
	var out, errOut bytes.Buffer
	if code := Run(ctx, args, nil, &out, &errOut); code != 0 {
		t.Fatalf("offline status: %s", &errOut)
	}
	var message session.Message
	if e = json.Unmarshal(out.Bytes(), &message); e != nil || message.State != "queued" || message.ID != request.RequestID {
		t.Fatalf("receipt %s %v", &out, e)
	}
	if _, e = os.Stat(opts.RuntimeDir); !os.IsNotExist(e) {
		t.Fatal("offline status touched backend runtime")
	}
	args[5] = "asd-b"
	out.Reset()
	errOut.Reset()
	if code := Run(ctx, args, nil, &out, &errOut); code == 0 {
		t.Fatal("cross-scope status exposed content")
	}
}
func TestMessagingInvalidArgumentsHaveNoBackendOrStateEffects(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	runtime := filepath.Join(root, "runtime")
	tails := [][]string{{"read", "--lines", "0", strings.Repeat("1", 32)}, {"read", "name"}, {"status", "name"}, {"deliver", "name"}, {"send", "--to", "name", "--body-file", "-"}, {"send", "--request-id", strings.Repeat("a", 32), "--to", strings.Repeat("1", 32), "--run", strings.Repeat("2", 32), "--body-file", "-"}}
	for _, tail := range tails {
		var out, errOut bytes.Buffer
		args := append([]string{"--backend", filepath.Join(root, "missing"), "--runtime-dir", runtime, "--state-dir", state}, tail...)
		if code := Run(context.Background(), args, strings.NewReader("bad\x1bdata"), &out, &errOut); code == 0 {
			t.Fatalf("accepted invalid request: %v", tail)
		}
	}
	for _, p := range []string{state, runtime} {
		if _, e := os.Stat(p); !os.IsNotExist(e) {
			t.Fatal("invalid message request touched runtime/state")
		}
	}
}
