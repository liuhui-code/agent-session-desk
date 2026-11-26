//go:build windows

package instance

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestWindowsInstanceReusesOnlyTheMatchingScopeAndClosesCleanly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	root := t.TempDir()
	key, err := ScopeKey("psmux", filepath.Join(root, "runtime"), filepath.Join(root, "state", "metadata.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	first, err := Listen(ctx, key, func() bool { calls.Add(1); return true })
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if second, e := Listen(ctx, key, func() bool { return false }); !errors.Is(e, ErrAlreadyRunning) {
		if second != nil {
			second.Close()
		}
		t.Fatalf("second owner accepted: %v", e)
	}
	found, restored, err := Notify(ctx, key)
	if err != nil || !found || !restored || calls.Load() != 1 {
		t.Fatalf("existing scope not reused: %v %v %v calls=%d", found, restored, err, calls.Load())
	}
	other, _ := ScopeKey("psmux", filepath.Join(root, "other"), filepath.Join(root, "state", "metadata.sqlite3"))
	if found, _, err = Notify(ctx, other); err != nil || found {
		t.Fatalf("activation crossed data scope: %v %v", found, err)
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Listen(ctx, key, func() bool { return true })
	if err != nil {
		t.Fatalf("normal quit retained instance: %v", err)
	}
	defer reopened.Close()
}

func TestWindowsInstanceCanceledLookupDoesNotClaimAnExistingOwner(t *testing.T) {
	root := t.TempDir()
	key, _ := ScopeKey("psmux", root, filepath.Join(root, "db"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	found, active, err := Notify(ctx, key)
	if found || active || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled query invented owner: %v %v %v", found, active, err)
	}
}

func TestWindowsInstanceReadyRequiresBindingAndDoesNotOverstateForeground(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	root := t.TempDir()
	key, _ := ScopeKey("psmux", root, filepath.Join(root, "db"))
	s, err := Listen(ctx, key, func() bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	short, stop := context.WithTimeout(ctx, 50*time.Millisecond)
	if err = WaitReady(short, key); !errors.Is(err, context.DeadlineExceeded) {
		stop()
		t.Fatalf("unbound endpoint was declared ready: %v", err)
	}
	stop()
	// A test-owned hidden HWND is deliberately not the foreground window.
	if err = s.BindWindow(s.hwnd); err != nil {
		t.Fatal(err)
	}
	if err = WaitReady(ctx, key); err != nil {
		t.Fatal(err)
	}
	found, active, err := Notify(ctx, key)
	if err != nil || !found || active {
		t.Fatalf("callback claimed success without actual foreground: %v %v %v", found, active, err)
	}
}
