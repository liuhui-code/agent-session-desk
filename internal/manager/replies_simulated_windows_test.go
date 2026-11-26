//go:build windows

package manager_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type replyAgentConfig struct {
	CLI                      string
	Globals                  []string
	CredentialFile           string
	Target                   session.Session
	MessageID, ReplyID, Task string
}
type replyAgentReport struct {
	MessageID       string        `json:"message_id"`
	Reply           session.Reply `json:"reply"`
	DuplicateStable bool          `json:"duplicate_stable"`
}

func simulatedReply(body string) string {
	sum := sha256.Sum256([]byte(body))
	return "result " + hex.EncodeToString(sum[:]) + " 中文🙂"
}

func replyProcessCall(ctx context.Context, c replyAgentConfig, input string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, c.CLI, append(append([]string(nil), c.Globals...), args...)...)
	cmd.Stdin = bytes.NewBufferString(input)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return out.Bytes(), fmt.Errorf("CLI operation %s failed: %s", args[0], &stderr)
	}
	return out.Bytes(), nil
}

// This worker is an ordinary controlled CLI hosted in a real psmux Run.
// Its tools are separate asd.exe processes; no in-process cli.Run shortcut.
func TestReplyAgentProcess(t *testing.T) {
	role := os.Getenv("ASD_REPLY_SIM_ROLE")
	if role == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	var c replyAgentConfig
	configPath := os.Getenv("ASD_REPLY_SIM_CONFIG")
	for {
		data, err := os.ReadFile(configPath)
		if err == nil && json.Unmarshal(data, &c) == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("worker config deadline")
		case <-time.After(25 * time.Millisecond):
		}
	}
	var report replyAgentReport
	if role == "A" {
		data, err := replyProcessCall(ctx, c, c.Task, "send", "--request-id", c.MessageID, "--to", c.Target.ID, "--run", c.Target.RunID, "--body-file", "-")
		if err != nil {
			t.Fatal(err)
		}
		var msg session.Message
		if json.Unmarshal(data, &msg) != nil || msg.ID != c.MessageID || msg.State != "queued" {
			t.Fatal("sender did not receive original request receipt")
		}
		for {
			data, err = replyProcessCall(ctx, c, "", "replies", msg.ID)
			if err != nil {
				t.Fatal(err)
			}
			var replies []session.Reply
			if json.Unmarshal(data, &replies) != nil {
				t.Fatal("invalid reply array")
			}
			if len(replies) == 1 {
				if replies[0].ReplyTo != msg.ID || replies[0].Source.RunID != c.Target.RunID || replies[0].Body != simulatedReply(c.Task) {
					t.Fatal("sender received wrong result or source")
				}
				report = replyAgentReport{MessageID: msg.ID, Reply: replies[0]}
				break
			}
			select {
			case <-ctx.Done():
				t.Fatal("sender reply deadline")
			case <-time.After(50 * time.Millisecond):
			}
		}
	} else if role == "B" {
		var request session.Message
		for request.ID == "" {
			data, err := replyProcessCall(ctx, c, "", "inbox", "--token-file", c.CredentialFile)
			if err != nil {
				t.Fatal(err)
			}
			var inbox []session.Message
			if json.Unmarshal(data, &inbox) != nil {
				t.Fatal("invalid inbox array")
			}
			for _, msg := range inbox {
				if msg.ID == c.MessageID {
					request = msg
					break
				}
			}
			if request.ID != "" {
				break
			}
			select {
			case <-ctx.Done():
				t.Fatal("receiver inbox deadline")
			case <-time.After(50 * time.Millisecond):
			}
		}
		if request.Target.ID != c.Target.ID || request.Target.RunID != c.Target.RunID || request.Body != c.Task {
			t.Fatal("wrong request reached receiver")
		}
		claimID, e := session.NewID()
		if e != nil {
			t.Fatal(e)
		}
		claimFile := c.CredentialFile + ".claim"
		claimData, e := replyProcessCall(ctx, c, "", "claim", "--token-file", c.CredentialFile, "--claim-file", claimFile, "--claim-id", claimID, "--message-id", request.ID)
		if e != nil {
			t.Fatal(e)
		}
		var grant session.TaskClaim
		if json.Unmarshal(claimData, &grant) != nil || !grant.NewlyClaimed || grant.State != "active" {
			t.Fatal("receiver did not get a new processing grant")
		}
		body := simulatedReply(request.Body) // computed from the actual tool response
		args := []string{"reply", "--token-file", c.CredentialFile, "--claim-file", claimFile, "--reply-id", c.ReplyID, "--reply-to", request.ID, "--body-file", "-"}
		data, err := replyProcessCall(ctx, c, body, args...)
		if err != nil {
			t.Fatal(err)
		}
		var first, again session.Reply
		if json.Unmarshal(data, &first) != nil {
			t.Fatal("reply JSON invalid")
		}
		data, err = replyProcessCall(ctx, c, body, args...)
		if err != nil {
			t.Fatal(err)
		}
		if json.Unmarshal(data, &again) != nil || again.ID != first.ID || !again.RecordedAt.Equal(first.RecordedAt) {
			t.Fatal("separate CLI retry duplicated result")
		}
		report = replyAgentReport{MessageID: request.ID, Reply: again, DuplicateStable: true}
	} else {
		t.Fatal("unknown controlled worker role")
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(os.Getenv("ASD_REPLY_SIM_REPORT"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestWindowsNativeReplyTwoAgentProcesses(t *testing.T) {
	bundle, product := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_REPLY_CLI")
	if bundle == "" || product == "" {
		t.Skip("explicit pinned bundle and newly built product required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	root := t.TempDir()
	id, err := session.NewID()
	if err != nil {
		t.Fatal(err)
	}
	opts := backend.Options{TerminalBundle: bundle, RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-reply-" + id[:12]}
	b, err := backend.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	kind, location, err := backend.Location(opts)
	if err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(ctx, filepath.Join(root, "state", "metadata.sqlite3"), strings.ToLower(kind+":"+location))
	if err != nil {
		t.Fatal(err)
	}
	m := manager.New(b, db)
	var originals []session.Session
	defer func() {
		for _, v := range originals {
			c, stop := context.WithTimeout(context.Background(), 10*time.Second)
			if err := m.Stop(c, v); err != nil {
				t.Errorf("owned worker cleanup: %v", err)
			}
			stop()
		}
		db.Close()
	}()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	start := func(role string) session.Session {
		v, err := m.Start(ctx, session.Launch{Name: "Simulated agent " + role, Executable: self, Dir: root, Args: []string{"-test.run=^TestReplyAgentProcess$"}, Env: []string{"ASD_REPLY_SIM_ROLE=" + role, "ASD_REPLY_SIM_CONFIG=" + filepath.Join(root, role+".config.json"), "ASD_REPLY_SIM_REPORT=" + filepath.Join(root, role+".report.json")}})
		if err != nil {
			t.Fatal(err)
		}
		originals = append(originals, v)
		return v
	}
	a, target := start("A"), start("B")
	globals := []string{"--terminal-bundle", bundle, "--runtime-dir", opts.RuntimeDir, "--namespace", opts.Namespace, "--state-dir", filepath.Join(root, "state")}
	c := replyAgentConfig{CLI: product, Globals: globals, Target: target, Task: "controlled request 中文 é 🧪"}
	c.MessageID, err = session.NewID()
	if err != nil {
		t.Fatal(err)
	}
	c.ReplyID, err = session.NewID()
	if err != nil {
		t.Fatal(err)
	}
	credential := filepath.Join(root, "B.credential.json")
	public, err := replyProcessCall(ctx, c, "", "agent-bind", "--session", target.ID, "--run", target.RunID, "--token-file", credential)
	if err != nil {
		t.Fatal(err)
	}
	var binding struct {
		Result string `json:"result"`
	}
	if json.Unmarshal(public, &binding) != nil || binding.Result != "bound" {
		t.Fatal("binding did not publish a public receipt")
	}
	for _, role := range []string{"B", "A"} {
		cfg := c
		if role == "B" {
			cfg.CredentialFile = credential
		}
		data, _ := json.Marshal(cfg)
		if err = os.WriteFile(filepath.Join(root, role+".config.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	reports := map[string]replyAgentReport{}
	for _, role := range []string{"A", "B"} {
		for {
			data, e := os.ReadFile(filepath.Join(root, role+".report.json"))
			var report replyAgentReport
			if e == nil && json.Unmarshal(data, &report) == nil {
				reports[role] = report
				break
			}
			select {
			case <-ctx.Done():
				p, _ := b.Preview(context.Background(), originals[map[string]int{"A": 0, "B": 1}[role]], 30)
				t.Fatalf("worker %s deadline; controlled screen: %s", role, p.Text)
			case <-time.After(50 * time.Millisecond):
			}
		}
	}
	if reports["A"].Reply.ID != c.ReplyID || reports["B"].Reply.ID != c.ReplyID || !reports["B"].DuplicateStable {
		t.Fatal("agent reports do not share the original reply")
	}
	status, err := m.Message(ctx, c.MessageID)
	if err != nil || status.State != "queued" {
		t.Fatal("metadata reply fabricated terminal write or completion")
	}
	rows, err := m.Replies(ctx, c.MessageID)
	if err != nil || len(rows) != 1 || rows[0].Source.ID != target.ID || rows[0].Source.RunID != target.RunID {
		t.Fatal("parent query lost source or deduplication")
	}
	if a.ID == target.ID || a.RunID == target.RunID {
		t.Fatal("workers did not have separate original Runs")
	}
	t.Log("two independent simulated agent processes used separate asd.exe tool calls: A send -> B inbox -> computed reply -> A replies; separate-process retry retained one result; queued unchanged")
}
