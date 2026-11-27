//go:build windows

package backend

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWindowsNativeSwitchLatencyProbe(t *testing.T) {
	bundle, fixture := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_MESSAGING_FIXTURE")
	if bundle == "" || fixture == "" {
		t.Skip("explicit native fixture required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := t.TempDir()
	id, _ := session.NewID()
	b, err := New(Options{RuntimeDir: filepath.Join(root, "r"), Namespace: "asd-switch-" + id[:12], TerminalBundle: bundle})
	if err != nil {
		t.Fatal(err)
	}
	var targets []session.Session
	defer func() {
		for _, s := range targets {
			c, x := context.WithTimeout(context.Background(), 10*time.Second)
			e := b.Stop(c, s)
			x()
			if e != nil {
				t.Error(e)
			}
		}
	}()
	for range 2 {
		s, e := b.Start(ctx, session.Launch{Name: "Synthetic switch", Dir: root, Executable: fixture})
		if e != nil {
			t.Fatal(e)
		}
		targets = append(targets, s)
	}
	measure := func(name string, f func() error) {
		t.Helper()
		start := time.Now()
		if e := f(); e != nil {
			t.Fatal(name, e)
		}
		t.Logf("%s elapsed=%s", name, time.Since(start))
	}
	measure("probe", func() error { _, e := b.Probe(ctx); return e })
	measure("validate", func() error { _, e := b.validateTarget(ctx, targets[0]); return e })
	measure("attach_checks", func() error { _, e := b.Attach(ctx, targets[0]); return e })
	next := 0
	open := func() error {
		s := targets[next%len(targets)]
		next++
		c, e := b.OpenTerminal(ctx, s, 80, 24)
		if e != nil {
			return e
		}
		defer c.Close()
		_, e = c.Next(ctx)
		return e
	}
	for range 3 {
		measure("idle_open_to_first_frame", open)
	}
	load, stop := context.WithCancel(ctx)
	var wg sync.WaitGroup
	for _, kind := range []string{"preview", "discovery"} {
		wg.Add(1)
		go func(kind string) {
			defer wg.Done()
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-load.Done():
					return
				case <-ticker.C:
					if kind == "discovery" {
						_, _ = b.List(load)
					} else {
						for _, s := range targets {
							_, _ = b.Preview(load, s, 100)
						}
					}
				}
			}
		}(kind)
	}
	defer func() { stop(); wg.Wait() }()
	time.Sleep(1100 * time.Millisecond)
	for range 3 {
		measure("loaded_open_to_first_frame", open)
	}
	stop()
	wg.Wait()
}
