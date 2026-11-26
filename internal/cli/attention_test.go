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

func TestAttentionEventStrictBoundedJSON(t *testing.T) {
	e := session.AttentionEvent{ID: strings.Repeat("a", 32), Version: 1, Kind: "question", State: "pending", Title: "Choose"}
	data, _ := json.Marshal(e)
	if got, err := readAttentionEvent("-", bytes.NewReader(data)); err != nil || got != e {
		t.Fatal(err)
	}
	for _, data := range []string{string(data) + "{}", strings.Replace(string(data), `"Choose"`, `"Choose","decision":"allow"`, 1), strings.Repeat(" ", session.MaxAttentionEventBytes+1), strings.Replace(string(data), `"Choose"`, `"\u001b[2J"`, 1)} {
		if _, err := readAttentionEvent("-", strings.NewReader(data)); err == nil {
			t.Fatal("unsafe event accepted")
		}
	}
}

func TestAttentionInvalidCommandHasNoEffects(t *testing.T) {
	root := t.TempDir()
	state, runtime := filepath.Join(root, "state"), filepath.Join(root, "runtime")
	for _, tail := range [][]string{{"attention-event"}, {"attention-status", "name"}, {"attention-list", "unexpected"}, {"attention-event", "--token-file", "missing", "--event-file", "-"}} {
		var out, stderr bytes.Buffer
		args := append([]string{"--backend", filepath.Join(root, "missing"), "--state-dir", state, "--runtime-dir", runtime}, tail...)
		if code := Run(context.Background(), args, strings.NewReader("{}"), &out, &stderr); code == 0 {
			t.Fatal("invalid command accepted")
		}
	}
	for _, path := range []string{state, runtime} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("invalid command changed state")
		}
	}
}

func TestAttentionQueriesOfflineWithoutLeakingCredential(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	state := filepath.Join(root, "state")
	opts := backend.Options{Binary: filepath.Join(root, "missing"), RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-attention"}
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
	e := session.AttentionEvent{ID: strings.Repeat("a", 32), Version: 1, Kind: "approval", State: "pending", Title: "Allow?"}
	if _, err = db.ReportAttention(ctx, c, e); err != nil {
		t.Fatal(err)
	}
	for _, tail := range [][]string{{"attention-list"}, {"attention-status", e.ID}} {
		var out, stderr bytes.Buffer
		args := append([]string{"--backend", opts.Binary, "--runtime-dir", opts.RuntimeDir, "--namespace", opts.Namespace, "--state-dir", state}, tail...)
		if code := Run(ctx, args, nil, &out, &stderr); code != 0 {
			t.Fatalf("offline query %s", &stderr)
		}
		if !bytes.Contains(out.Bytes(), []byte(e.ID)) || bytes.Contains(out.Bytes(), []byte(c.Secret)) || bytes.Contains(out.Bytes(), []byte(`"terminal_available":true`)) {
			t.Fatal("offline query leaked authority or claimed liveness")
		}
	}
}
