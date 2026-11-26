//go:build windows

package instance

import (
	"context"
	"errors"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWindowsInstanceConcurrentStartupKeepsOneOwner(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	root := t.TempDir()
	key, _ := ScopeKey("psmux", root, filepath.Join(root, "db"))
	var wg sync.WaitGroup
	results := make(chan *Server, 5)
	failures := make(chan error, 5)
	for range 5 {
		wg.Go(func() {
			s, e := Listen(ctx, key, func() bool { return true })
			if e != nil {
				failures <- e
			} else {
				results <- s
			}
		})
	}
	wg.Wait()
	close(results)
	close(failures)
	owners := 0
	for s := range results {
		owners++
		defer s.Close()
	}
	if owners != 1 {
		t.Fatalf("five concurrent launches created %d owners", owners)
	}
	for err := range failures {
		if !errors.Is(err, ErrAlreadyRunning) {
			t.Fatal(err)
		}
	}
}

func TestWindowsInstanceKilledOwnerReleasesAnAbandonedMutex(t *testing.T) {
	root := t.TempDir()
	key, _ := ScopeKey("psmux", root, filepath.Join(root, "db"))
	ready := filepath.Join(root, "ready")
	exe, _ := os.Executable()
	cmd := exec.Command(exe, "-test.run", "^TestInstanceProcessHelper$", "-test.timeout", "30s")
	cmd.Env = append(os.Environ(), "ASD_INSTANCE_HELPER_KEY="+key, "ASD_INSTANCE_HELPER_READY="+ready)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.ProcessState == nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("owned helper never became ready")
		}
		time.Sleep(10 * time.Millisecond)
	}
	name, _ := objectName(key)
	retained, err := openMutex(name)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(retained)
	if err = cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	cmd.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s, err := Listen(ctx, key, func() bool { return true })
	if err != nil {
		t.Fatalf("killed helper left fake running state: %v", err)
	}
	defer s.Close()
}

func TestWindowsInstanceUnresponsiveOwnerDoesNotPermitAnotherUI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	root := t.TempDir()
	key, _ := ScopeKey("psmux", root, filepath.Join(root, "db"))
	gate := make(chan struct{})
	var entered atomic.Bool
	s, err := Listen(ctx, key, func() bool { entered.Store(true); <-gate; return false })
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	defer close(gate)
	short, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()
	found, activated, err := Notify(short, key)
	if !found || activated || err == nil || !entered.Load() {
		t.Fatalf("unresponsive owner was treated as absent/success: %v %v %v", found, activated, err)
	}
	if second, e := Listen(ctx, key, func() bool { return true }); !errors.Is(e, ErrAlreadyRunning) {
		if second != nil {
			second.Close()
		}
		t.Fatalf("timeout authorized a second UI: %v", e)
	}
}

func TestInstanceProcessHelper(t *testing.T) {
	key := os.Getenv("ASD_INSTANCE_HELPER_KEY")
	if key == "" {
		t.Skip("subprocess helper")
	}
	s, err := Listen(context.Background(), key, func() bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err = os.WriteFile(os.Getenv("ASD_INSTANCE_HELPER_READY"), []byte("ready"), 0600); err != nil {
		t.Fatal(err)
	}
	select {}
}
