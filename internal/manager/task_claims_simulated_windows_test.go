//go:build windows

package manager_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type claimWorkerConfig struct {
	Base                                      replyAgentConfig
	ClaimID, ClaimFile, Start, Report, Effect string
	Seconds                                   int
	Hold                                      bool
}
type claimWorkerReport struct {
	Granted bool
	ClaimID string
}

func TestTaskClaimWorkerProcess(t *testing.T) {
	config := os.Getenv("ASD_TASK_CLAIM_WORKER")
	if config == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	data, err := os.ReadFile(config)
	var c claimWorkerConfig
	if err != nil || json.Unmarshal(data, &c) != nil {
		t.Fatal("worker config invalid")
	}
	for {
		if _, err = os.Stat(c.Start); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("start barrier expired")
		case <-time.After(20 * time.Millisecond):
		}
	}
	data, err = replyProcessCall(ctx, c.Base, "", "claim", "--token-file", c.Base.CredentialFile, "--claim-file", c.ClaimFile, "--claim-id", c.ClaimID, "--message-id", c.Base.MessageID, "--lease-seconds", strconv.Itoa(c.Seconds))
	var grant session.TaskClaim
	fresh := err == nil && json.Unmarshal(data, &grant) == nil && grant.State == "active" && grant.NewlyClaimed && grant.ExpiresAt.After(time.Now().UTC())
	if fresh {
		f, e := os.OpenFile(c.Effect, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal("second execution attempted the existing synthetic effect")
		}
		if _, e = f.WriteString(c.ClaimID); e == nil {
			e = f.Sync()
		}
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
		if !c.Hold {
			_, e = replyProcessCall(ctx, c.Base, "processed 中文", "reply", "--token-file", c.Base.CredentialFile, "--claim-file", c.ClaimFile, "--reply-id", c.Base.ReplyID, "--reply-to", c.Base.MessageID, "--body-file", "-")
			if e != nil {
				t.Fatal(e)
			}
		}
	}
	report, _ := json.Marshal(claimWorkerReport{Granted: fresh, ClaimID: c.ClaimID})
	if err = os.WriteFile(c.Report, report, 0600); err != nil {
		t.Fatal(err)
	}
	if fresh && c.Hold {
		var b [1]byte
		_, _ = os.Stdin.Read(b[:])
		t.Fatal("fault worker resumed instead of being terminated")
	}
}

func TestWindowsNativeTaskClaimCompetitionAndCrash(t *testing.T) {
	bundle, product, fixture := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_REPLY_CLI"), os.Getenv("ASD_NATIVE_MESSAGING_FIXTURE")
	if bundle == "" || product == "" || fixture == "" {
		t.Skip("explicit pinned bundle, product and controlled CLI required")
	}
	for _, crash := range []bool{false, true} {
		t.Run(map[bool]string{false: "competition", true: "killed_after_effect"}[crash], func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			root := t.TempDir()
			id, err := session.NewID()
			if err != nil {
				t.Fatal(err)
			}
			opts := backend.Options{TerminalBundle: bundle, RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-task-" + id[:12]}
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
			defer db.Close()
			target, err := m.Start(ctx, session.Launch{Name: "Claim delegation target", Executable: fixture, Dir: root})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				c, stop := context.WithTimeout(context.Background(), 10*time.Second)
				defer stop()
				if err := m.Stop(c, target); err != nil {
					t.Errorf("owned target cleanup: %v", err)
				}
			}()
			base := replyAgentConfig{CLI: product, Globals: []string{"--terminal-bundle", bundle, "--runtime-dir", opts.RuntimeDir, "--namespace", opts.Namespace, "--state-dir", filepath.Join(root, "state")}, Target: target, CredentialFile: filepath.Join(root, "agent.json")}
			base.MessageID, _ = session.NewID()
			base.ReplyID, _ = session.NewID()
			if _, err = replyProcessCall(ctx, base, "", "agent-bind", "--session", target.ID, "--run", target.RunID, "--token-file", base.CredentialFile); err != nil {
				t.Fatal(err)
			}
			if _, err = replyProcessCall(ctx, base, "task", "send", "--request-id", base.MessageID, "--to", target.ID, "--run", target.RunID, "--body-file", "-"); err != nil {
				t.Fatal(err)
			}
			self, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			type worker struct {
				cmd    *exec.Cmd
				config claimWorkerConfig
				waited bool
				log    bytes.Buffer
				stdin  interface{ Close() error }
			}
			var workers []*worker
			defer func() {
				for _, w := range workers {
					if !w.waited {
						_ = w.cmd.Process.Kill()
						_ = w.cmd.Wait()
					}
					_ = w.stdin.Close()
				}
			}()
			start := func(n int) *worker {
				key, _ := session.NewID()
				seconds := 30
				if crash {
					seconds = 2
				}
				c := claimWorkerConfig{Base: base, ClaimID: key, ClaimFile: filepath.Join(root, "claim-"+strconv.Itoa(n)+".json"), Start: filepath.Join(root, "go"), Report: filepath.Join(root, "report-"+strconv.Itoa(n)+".json"), Effect: filepath.Join(root, "effect"), Seconds: seconds, Hold: crash}
				data, _ := json.Marshal(c)
				path := filepath.Join(root, "worker-"+strconv.Itoa(n)+".json")
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
				w := &worker{config: c}
				w.cmd = exec.CommandContext(ctx, self, "-test.run=^TestTaskClaimWorkerProcess$", "-test.timeout=75s")
				w.cmd.Env = append(os.Environ(), "ASD_TASK_CLAIM_WORKER="+path)
				w.cmd.Stdout = &w.log
				w.cmd.Stderr = &w.log
				pipe, e := w.cmd.StdinPipe()
				if e != nil {
					t.Fatal(e)
				}
				w.stdin = pipe
				if e = w.cmd.Start(); e != nil {
					t.Fatal(e)
				}
				workers = append(workers, w)
				return w
			}
			start(0)
			start(1)
			if err = os.WriteFile(filepath.Join(root, "go"), []byte("go"), 0600); err != nil {
				t.Fatal(err)
			}
			await := func(w *worker) claimWorkerReport {
				for {
					data, e := os.ReadFile(w.config.Report)
					var report claimWorkerReport
					if e == nil && json.Unmarshal(data, &report) == nil {
						return report
					}
					select {
					case <-ctx.Done():
						t.Fatal("worker report deadline")
					case <-time.After(25 * time.Millisecond):
					}
				}
			}
			var winner *worker
			wins := 0
			for _, w := range workers {
				if await(w).Granted {
					wins++
					winner = w
				}
			}
			if wins != 1 {
				t.Fatalf("processing grants=%d", wins)
			}
			for _, w := range workers {
				if crash && w == winner {
					if e := w.cmd.Process.Kill(); e != nil {
						t.Fatal(e)
					}
					if e := w.cmd.Wait(); e == nil {
						t.Fatal("fault worker was not terminated")
					}
				} else {
					if e := w.cmd.Wait(); e != nil {
						t.Fatalf("controlled worker failed: %s", &w.log)
					}
				}
				w.waited = true
			}
			effect, e := os.ReadFile(filepath.Join(root, "effect"))
			if e != nil || string(effect) != winner.config.ClaimID {
				t.Fatal("independent synthetic effect is not unique")
			}
			var status struct {
				Claim *session.TaskClaim `json:"claim"`
			}
			for {
				data, e := replyProcessCall(ctx, base, "", "claim-status", base.MessageID)
				if e != nil || json.Unmarshal(data, &status) != nil || status.Claim == nil {
					t.Fatal("offline claim status invalid")
				}
				if !crash || status.Claim.State == "unknown" {
					break
				}
				select {
				case <-ctx.Done():
					t.Fatal("lease did not expire")
				case <-time.After(50 * time.Millisecond):
				}
			}
			if crash {
				if status.Claim.Reason != "lease_expired" {
					t.Fatal("wrong uncertainty reason")
				}
				restarted := start(2)
				if await(restarted).Granted {
					t.Fatal("new process received interrupted work")
				}
				if e := restarted.cmd.Wait(); e != nil {
					t.Fatalf("replacement worker failed: %s", &restarted.log)
				}
				restarted.waited = true
				if _, e = replyProcessCall(ctx, base, "", "renew", "--token-file", base.CredentialFile, "--claim-file", winner.config.ClaimFile); e == nil {
					t.Fatal("expired worker renewed")
				}
				if _, e = replyProcessCall(ctx, base, "late", "reply", "--token-file", base.CredentialFile, "--claim-file", winner.config.ClaimFile, "--reply-id", base.ReplyID, "--reply-to", base.MessageID, "--body-file", "-"); e == nil {
					t.Fatal("expired worker published a result")
				}
				rows, e := m.Replies(ctx, base.MessageID)
				if e != nil || len(rows) != 0 {
					t.Fatal("interrupted result became durable")
				}
			} else if status.Claim.State != "replied" {
				t.Fatal("winner did not atomically publish result")
			}
			if _, e = replyProcessCall(ctx, base, "", "deliver", base.MessageID); e == nil {
				t.Fatal("processing ownership allowed terminal fallback")
			}
			msg, e := m.Message(ctx, base.MessageID)
			if e != nil || msg.State != "queued" {
				t.Fatal("processing changed terminal delivery evidence")
			}
			t.Log("two separate processing workers, one effect; private claim fencing; crash uses exact owned Kill and real 2-second lease; no automatic reassignment or terminal fallback")
		})
	}
}
