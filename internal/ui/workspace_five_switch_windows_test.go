//go:build windows

package ui

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestWindowsNativeFiveSessionWarmCycle(t *testing.T) {
	bundle, fixture := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_MESSAGING_FIXTURE")
	if bundle == "" || fixture == "" {
		t.Skip("explicit isolated native fixture required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	root := t.TempDir()
	id, _ := session.NewID()
	opts := backend.Options{RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-five-" + id[:12], TerminalBundle: bundle}
	b, err := backend.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	kind, location, _ := backend.Location(opts)
	db, err := store.Open(ctx, filepath.Join(root, "state", "metadata.sqlite3"), strings.ToLower(kind+":"+location))
	if err != nil {
		t.Fatal(err)
	}
	svc := manager.New(b, db)
	var originals []session.Session
	var h *workspaceHarness
	defer func() {
		if h != nil {
			h.m.Close()
		}
		for _, s := range originals {
			c, x := context.WithTimeout(context.Background(), 15*time.Second)
			e := svc.Stop(c, s)
			x()
			if e != nil {
				t.Errorf("original cleanup %s: %v", s.ID, e)
			}
		}
		if e := db.Close(); e != nil {
			t.Error(e)
		}
	}()
	for i := range 5 {
		s, e := svc.Start(ctx, session.Launch{Name: fmt.Sprintf("Five-%d", i), Dir: root, Executable: fixture})
		if e != nil {
			t.Fatal(e)
		}
		originals = append(originals, s)
	}
	h = newWorkspaceHarness(t, svc)
	until := func(check func() bool) {
		t.Helper()
		timer := time.NewTimer(20 * time.Second)
		defer timer.Stop()
		for !check() {
			select {
			case msg := <-h.messages:
				h.send(msg)
			case <-timer.C:
				t.Fatal("native original did not become ready")
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		}
	}
	until(func() bool {
		return h.m.workspace.conn != nil && !h.m.workspace.busy && h.m.workspace.screen.Width > 0 && h.m.workspace.pending == 0
	})
	if len(h.m.workspace.tabs) != 5 {
		t.Fatal("initial original set incomplete")
	}
	hashes := map[string][32]byte{}
	for i := range 5 {
		w := h.m.workspace
		if i != w.active {
			h.run(h.m.activateWorkspaceNow(i))
			until(func() bool { return w.conn != nil && !w.busy && w.pending == 0 })
		}
		g := h.m.workspaceGeometry()
		hashes[activityKey(w.tabs[i].session)] = sha256.Sum256([]byte(w.screen.Render(g.width, g.height)))
	}
	for cycle := range 15 {
		index := cycle % 5
		w := h.m.workspace
		until(func() bool { return w.pending == 0 })
		cmd := h.m.activateWorkspaceNow(index)
		g := h.m.workspaceGeometry()
		key := activityKey(w.tabs[index].session)
		if w.conn != nil || !w.busy || w.screen.Width == 0 || h.m.View().Cursor != nil {
			t.Fatalf("cycle %d lacked read-only original", cycle)
		}
		if sha256.Sum256([]byte(w.screen.Render(g.width, g.height))) != hashes[key] {
			t.Fatalf("cycle %d wrong original frame", cycle)
		}
		if strings.Contains(h.m.View().Content, "Connecting to the original terminal") {
			t.Fatalf("cycle %d flashed connecting", cycle)
		}
		total := 0
		for _, s := range w.snapshots {
			total += s.bytes
		}
		if len(w.snapshots) > workspaceSnapshotLimit || total > workspaceSnapshotBytes {
			t.Fatal("cache budget grew")
		}
		h.run(cmd)
		until(func() bool { return w.conn != nil && !w.busy && w.pending == 0 })
		if key != activityKey(w.tabs[w.active].session) {
			t.Fatal("connected another original")
		}
		hashes[key] = sha256.Sum256([]byte(w.screen.Render(g.width, g.height)))
	}
	t.Log("five native originals, 15 warm returns, exact per-original screen hashes, single client, unchanged cache limits")
}
