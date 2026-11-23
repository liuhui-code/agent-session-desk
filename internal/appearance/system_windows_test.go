//go:build windows

package appearance

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows/registry"
)

// This test changes only a freshly created task-owned key, never the OS theme.
func TestWindowsAppearanceIsolatedKeyNotificationsAndCleanup(t *testing.T) {
	if os.Getenv("ASD_NATIVE_APPEARANCE") != "1" {
		t.Skip("opt-in isolated native registry test")
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	path := `Software\AgentSessionDeskAppearanceTest-` + hex.EncodeToString(nonce[:])
	k, existed, err := registry.CreateKey(registry.CURRENT_USER, path, registry.ALL_ACCESS)
	if err != nil {
		t.Fatal(err)
	}
	if existed {
		_ = k.Close()
		t.Fatal("refusing existing test key")
	}
	t.Cleanup(func() {
		if err := k.Close(); err != nil {
			t.Error(err)
		}
		if err := registry.DeleteKey(registry.CURRENT_USER, path); err != nil {
			t.Error(err)
		}
	})
	if err = k.SetDWordValue("AppsUseLightTheme", 0); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, err := openWindowsAt(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	})
	next := func() Observation {
		t.Helper()
		readCtx, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		v, err := c.Next(readCtx)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	if v := next(); v.Theme != "dark" || v.Err != nil || v.Source != "windows" {
		t.Fatalf("initial: %+v", v)
	}
	if err = k.SetDWordValue("AppsUseLightTheme", 1); err != nil {
		t.Fatal(err)
	}
	if v := next(); v.Theme != "light" || v.Err != nil {
		t.Fatalf("notification: %+v", v)
	}
	if err = k.SetStringValue("AppsUseLightTheme", "PRIVATE-SYNTHETIC"); err != nil {
		t.Fatal(err)
	}
	if v := next(); v.Err == nil || v.Theme != "" || strings.Contains(v.Err.Error(), "PRIVATE-SYNTHETIC") {
		t.Fatal("invalid type accepted or exposed")
	}
	if err = k.SetDWordValue("AppsUseLightTheme", 9); err != nil {
		t.Fatal(err)
	}
	if v := next(); v.Err == nil || v.Theme != "" {
		t.Fatal("invalid DWORD accepted")
	}
	if err = k.SetDWordValue("AppsUseLightTheme", 0); err != nil {
		t.Fatal(err)
	}
	if v := next(); v.Theme != "dark" || v.Err != nil {
		t.Fatalf("recovery: %+v", v)
	}
	cancel()
	closed := make(chan error, 1)
	go func() { closed <- c.Close() }()
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("native observer did not stop")
	}
	if _, err = c.Next(context.Background()); err == nil {
		t.Fatal("closed observer remained readable")
	}
	if err = c.Close(); err != nil {
		t.Fatal("second close failed", err)
	}
}
