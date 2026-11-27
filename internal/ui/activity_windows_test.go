//go:build windows

package ui

import (
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

func TestWindowsNativeActivityChangesAndSilence(t *testing.T) {
	bundle, fixture := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_MESSAGING_FIXTURE")
	if bundle == "" || fixture == "" {
		t.Skip("explicit native fixture required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	root := t.TempDir()
	id, _ := session.NewID()
	opts := backend.Options{RuntimeDir: filepath.Join(root, "r"), Namespace: "asd-act-" + id[:12], TerminalBundle: bundle}
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
	v, err := m.Start(ctx, session.Launch{Name: "Activity", Dir: root, Executable: fixture})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	defer func() {
		c, x := context.WithTimeout(context.Background(), 10*time.Second)
		defer x()
		if e := m.Stop(c, v); e != nil {
			t.Error(e)
		}
		db.Close()
	}()
	h := newWorkspaceHarness(t, m)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.workspace.busy })
	sample := func() { h.run(h.m.sampleActivity()); h.until(func() bool { return !h.m.activity.busy }) }
	sample()
	if h.m.activityMark(v) != h.m.tokens.activityIdle {
		t.Fatal("baseline")
	}
	c, err := b.OpenTerminal(ctx, v, 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err = c.SendText(ctx, "ACTIVITY"); err != nil {
		t.Fatal(err)
	}
	if err = c.SendKey(ctx, "Enter"); err != nil {
		t.Fatal(err)
	}
	// Observe the independent native page before sampling; writes aren't output.
	for {
		p, _ := b.Preview(ctx, v, 100)
		if strings.Contains(p.Text, "ACTIVITY") {
			break
		}
		if ctx.Err() != nil {
			t.Fatal("output deadline")
		}
	}
	sample()
	if !h.m.activity.entries[activityKey(v)].moving(time.Now()) {
		t.Fatal("output not active")
	}
	time.Sleep(activityQuiet + 100*time.Millisecond)
	sample()
	if h.m.activityMark(v) != h.m.tokens.activityIdle {
		t.Fatal("silence")
	}
	t.Log("real fixture -> capture baseline -> native output -> moving -> silence; no agent-state inference")
}
