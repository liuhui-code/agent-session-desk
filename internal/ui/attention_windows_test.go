//go:build windows

package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

// Uses pinned native psmux and two original fixture processes. UI keys enter
// the real model; this is not Windows Terminal desktop keyboard evidence.
func TestWindowsNativeAttentionOriginalForm(t *testing.T) {
	bundle, fixture, product := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_MESSAGING_FIXTURE"), os.Getenv("ASD_NATIVE_REPLY_CLI")
	if bundle == "" || fixture == "" || product == "" {
		t.Skip("explicit pinned bundle, fixture and product required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := t.TempDir()
	id, err := session.NewID()
	if err != nil {
		t.Fatal(err)
	}
	opts := backend.Options{TerminalBundle: bundle, RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-attn-" + id[:12]}
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
			e := m.Stop(c, v)
			stop()
			if e != nil {
				t.Errorf("owned fixture cleanup: %v", e)
			}
		}
		db.Close()
	}()
	wait := func(check func() bool) {
		t.Helper()
		for !check() {
			select {
			case <-ctx.Done():
				t.Fatal("native attention deadline")
			case <-time.After(30 * time.Millisecond):
			}
		}
	}
	globals := []string{"--terminal-bundle", bundle, "--runtime-dir", opts.RuntimeDir, "--namespace", opts.Namespace, "--state-dir", filepath.Join(root, "state")}
	call := func(input string, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, product, append(append([]string(nil), globals...), args...)...)
		cmd.Stdin = strings.NewReader(input)
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return out.Bytes(), fmt.Errorf("%s: %s", args[0], &stderr)
		}
		return out.Bytes(), nil
	}
	config := filepath.Join(root, "form.config.json")
	start := func(name string, form bool) (session.Session, string) {
		journal := filepath.Join(root, name+".jsonl")
		env := []string{"ASD_FIXTURE_REPORT_PATH=" + journal}
		if form {
			env = append(env, "ASD_FIXTURE_ATTENTION_CONFIG="+config)
		}
		v, e := m.Start(ctx, session.Launch{Name: name, Executable: fixture, Dir: root, Env: env})
		if e != nil {
			t.Fatal(e)
		}
		originals = append(originals, v)
		wait(func() bool { data, _ := os.ReadFile(journal); return bytes.Contains(data, []byte(`"event":"started"`)) })
		return v, journal
	}
	a, aLog := start("Current", false)
	target, bLog := start("Decision", true)
	credential := filepath.Join(root, "form.credential.json")
	if _, err = call("", "agent-bind", "--session", target.ID, "--run", target.RunID, "--token-file", credential); err != nil {
		t.Fatal(err)
	}
	attentionID, _ := session.NewID()
	data, _ := json.Marshal(struct {
		CLI                         string
		Globals                     []string
		CredentialFile, AttentionID string
	}{product, globals, credential, attentionID})
	if err = os.WriteFile(config, data, 0600); err != nil {
		t.Fatal(err)
	}
	wait(func() bool {
		data, _ := os.ReadFile(bLog)
		return bytes.Contains(data, []byte(`"event":"decision_waiting"`))
	})
	inputs := func(path string) []string {
		data, _ := os.ReadFile(path)
		var lines []string
		for _, line := range bytes.Split(data, []byte{'\n'}) {
			var event struct{ Event, Text string }
			if json.Unmarshal(line, &event) == nil && event.Event == "input" {
				lines = append(lines, event.Text)
			}
		}
		return lines
	}
	pids := func(path string) int {
		data, _ := os.ReadFile(path)
		var v struct{ PID int }
		json.Unmarshal(bytes.Split(data, []byte{'\n'})[0], &v)
		return v.PID
	}
	aPID, bPID := pids(aLog), pids(bLog)
	if aPID <= 0 || bPID <= 0 || aPID == bPID {
		t.Fatal("independent fixture process identities missing")
	}
	h := newWorkspaceHarness(t, m)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	// Discovery order is not an activation contract. Explicitly choose A.
	for i, tab := range h.m.workspace.tabs {
		if sameRun(tab.session, a) {
			h.run(h.m.activateWorkspace(i))
			break
		}
	}
	h.until(func() bool {
		return h.m.workspace.conn != nil && sameRun(h.m.workspace.tabs[h.m.workspace.active].session, a)
	})
	h.send(tea.KeyPressMsg{Code: 'd', Text: "DRAFT_A"})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	geometry := h.m.workspaceGeometry()
	h.run(h.m.loadSessions())
	h.until(func() bool { return len(h.m.attention.items) == 1 })
	if !sameRun(h.m.workspace.tabs[h.m.workspace.active].session, a) || h.m.workspaceGeometry() != geometry || !strings.Contains(ansi.Strip(h.m.View().Content), "need attention") {
		t.Fatal("passive reminder stole focus or resized terminal")
	}
	if len(inputs(bLog)) != 0 {
		t.Fatal("pending event answered native form")
	}
	msgID, _ := session.NewID()
	if _, err = call("y\n", "send", "--request-id", msgID, "--to", target.ID, "--run", target.RunID, "--body-file", "-", "--deliver"); err == nil {
		t.Fatal("automatic input entered native decision form")
	}
	if len(inputs(bLog)) != 0 {
		t.Fatal("blocked delivery still reached form")
	}
	h.send(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	h.until(func() bool {
		return h.m.workspace.conn != nil && sameRun(h.m.workspace.tabs[h.m.workspace.active].session, target) && !h.m.attention.busy
	})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "[y/N]") })
	item, err := db.Attention(ctx, attentionID)
	if err != nil || item.ViewedAt == nil || item.State != "pending" || len(inputs(bLog)) != 0 {
		t.Fatal("navigation answered or resolved original form", err)
	}
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	wait(func() bool {
		data, _ := os.ReadFile(bLog)
		return bytes.Contains(data, []byte(`"event":"decision_resolved"`))
	})
	h.run(h.m.loadSessions())
	h.until(func() bool { return len(h.m.attention.items) == 0 })
	if lines := inputs(bLog); len(lines) != 1 || lines[0] != "y" {
		t.Fatalf("native decision input %v", lines)
	}
	// Return by identity; retained tab order is independent of session order.
	for i, tab := range h.m.workspace.tabs {
		if sameRun(tab.session, a) {
			h.run(h.m.activateWorkspace(i))
			break
		}
	}
	h.until(func() bool {
		return h.m.workspace.conn != nil && sameRun(h.m.workspace.tabs[h.m.workspace.active].session, a)
	})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "DRAFT_A") })
	if len(inputs(aLog)) != 0 || pids(aLog) != aPID || pids(bLog) != bPID {
		t.Fatal("original process/draft replaced or submitted")
	}
	view, err := m.Workspace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, original := range originals {
		found := false
		for _, live := range view.Sessions {
			if sameRun(original, live) && original.BackendID == live.BackendID {
				found = true
			}
		}
		if !found {
			t.Fatal("original Run lost")
		}
	}
	h.m.Close()
	t.Logf("pending -> reminder without focus/size change -> Alt+N/viewed still pending -> explicit y/Enter -> resolved; original fixture PIDs %d/%d and Session/Run retained, A draft unsubmitted", aPID, bPID)
}
