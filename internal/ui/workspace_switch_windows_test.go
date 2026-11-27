//go:build windows

package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWindowsNativeWarmSessionSwitch(t *testing.T) {
	bundle, fixture := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_MESSAGING_FIXTURE")
	if bundle == "" || fixture == "" {
		t.Skip("explicit native fixture required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Second)
	defer cancel()
	root := t.TempDir()
	id, _ := session.NewID()
	opts := backend.Options{RuntimeDir: filepath.Join(root, "r"), Namespace: "asd-warm-" + id[:12], TerminalBundle: bundle}
	b, err := backend.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	k, l, _ := backend.Location(opts)
	db, err := store.Open(ctx, filepath.Join(root, "state", "metadata.sqlite3"), strings.ToLower(k+":"+l))
	if err != nil {
		t.Fatal(err)
	}
	m := manager.New(b, db)
	var targets []session.Session
	defer func() {
		for _, s := range targets {
			c, x := context.WithTimeout(context.Background(), 10*time.Second)
			e := m.Stop(c, s)
			x()
			if e != nil {
				t.Error(e)
			}
		}
		db.Close()
	}()
	for _, name := range []string{"A", "B"} {
		s, e := m.Start(ctx, session.Launch{Name: "Switch " + name, Dir: root, Executable: fixture})
		if e != nil {
			t.Fatal(e)
		}
		targets = append(targets, s)
		c, e := b.OpenTerminal(ctx, s, 80, 24)
		if e != nil {
			t.Fatal(e)
		}
		e = c.SendText(ctx, "SWITCH_"+name)
		if e == nil {
			e = c.SendKey(ctx, "Enter")
		}
		c.Close()
		if e != nil {
			t.Fatal(e)
		}
	}
	h := newWorkspaceHarness(t, m)
	wait := func(check func() bool) {
		t.Helper()
		deadline := time.NewTimer(12 * time.Second)
		defer deadline.Stop()
		for !check() {
			select {
			case msg := <-h.messages:
				h.send(msg)
			case <-deadline.C:
				t.Fatal("native switch deadline")
			}
		}
	}
	wait(func() bool {
		return !h.m.workspace.busy && h.m.workspace.conn != nil && strings.Contains(h.m.View().Content, "SWITCH_A")
	})
	h.run(h.m.activateWorkspace(1))
	wait(func() bool {
		return !h.m.workspace.busy && h.m.workspace.conn != nil && strings.Contains(h.m.View().Content, "SWITCH_B")
	})
	for round := range 6 {
		index := round % 2
		marker := []string{"SWITCH_A", "SWITCH_B"}[index]
		other := []string{"SWITCH_B", "SWITCH_A"}[index]
		started := time.Now()
		cmd := h.m.activateWorkspace(index)
		view := h.m.View()
		if cmd == nil || !h.m.workspace.busy || h.m.workspace.conn != nil || view.Cursor != nil || !strings.Contains(view.Content, marker) || strings.Contains(view.Content, other) || !strings.Contains(view.Content, "last observed") {
			t.Fatal("warm switch did not expose only its target's read-only snapshot")
		}
		visible := time.Since(started)
		h.run(cmd)
		wait(func() bool {
			return !h.m.workspace.busy && h.m.workspace.conn != nil && strings.Contains(h.m.View().Content, marker)
		})
		total := 0
		for _, entry := range h.m.workspace.snapshots {
			total += entry.bytes
		}
		if total > workspaceSnapshotBytes || len(h.m.workspace.snapshots) > workspaceSnapshotLimit {
			t.Fatal("cache limit")
		}
		t.Logf("warm target=%d visible=%s connected=%s retained_charge=%d", index, visible, time.Since(started), total)
	}
	h.send(tea.KeyPressMsg{Code: 'x', Text: "UNSUBMITTED_DRAFT"})
	wait(func() bool { return h.m.workspace.pending == 0 })
	h.m.Close()
	p, e := b.Preview(ctx, targets[1], 100)
	if e != nil || !strings.Contains(p.Text, "UNSUBMITTED_DRAFT") {
		t.Fatal("original draft was lost")
	}
	rows, e := m.List(ctx)
	if e != nil || len(rows) != 2 {
		t.Fatal("original Runs missing")
	}
	for _, s := range rows {
		if !sameRun(s, targets[0]) && !sameRun(s, targets[1]) {
			t.Fatal("Run replaced")
		}
	}
	t.Log("two original Runs retained; six warm switches; no draft submission or automatic input replay")
}
