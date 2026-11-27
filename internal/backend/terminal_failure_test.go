package backend

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A failed concrete-pointer constructor must not cross the Connection
// interface boundary as a typed nil: view cleanup legitimately closes a
// non-nil connection, including stale asynchronous open results.
func TestFailedTerminalHandshakeReturnsNilConnection(t *testing.T) {
	b := newFake(t)
	target := startFake(t, b)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	base := filepath.Join(b.dataDir, b.namespace+"__"+target.BackendID)
	for suffix, value := range map[string]string{
		".port": fmt.Sprint(listener.Addr().(*net.TCPAddr).Port),
		".key":  "synthetic-handshake-key",
	} {
		if err := os.WriteFile(base+suffix, []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		peer, err := listener.Accept()
		if err != nil {
			return
		}
		defer peer.Close()
		_ = peer.SetDeadline(time.Now().Add(5 * time.Second))
		reader := bufio.NewReader(peer)
		if _, err := reader.ReadString('\n'); err != nil {
			return
		}
		if _, err := fmt.Fprint(peer, "OK\n"); err != nil {
			return
		}
		// The connection drops after authentication and before the first
		// persistent screen, matching the native fault-injection boundary.
		_, _ = reader.ReadString('\n')
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	before := len(commandLog(t))
	connection, err := b.OpenTerminal(ctx, target, 80, 24)
	if err == nil {
		t.Fatal("disconnected handshake unexpectedly succeeded")
	}
	if connection != nil {
		t.Fatalf("failed open returned non-nil Connection (%T); caller cleanup can panic", connection)
	}
	checks := 0
	for _, args := range commandLog(t)[before:] {
		for _, arg := range args {
			if arg == metadataOption {
				checks++
			}
		}
	}
	if checks != 1 {
		t.Fatalf("one open revalidated the same Run %d times", checks)
	}
	select {
	case <-finished:
	case <-ctx.Done():
		t.Fatal("handshake peer did not finish")
	}
}
