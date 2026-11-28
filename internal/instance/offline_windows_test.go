//go:build windows

package instance

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
)

func TestWithStoppedRefusesLiveUIAndExcludesAnotherListener(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(t.TempDir())))
	live, err := Listen(ctx, key, func() bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	if err = WithStopped(ctx, key, func() error { calls++; return nil }); err == nil || calls != 0 {
		t.Fatal("live UI update accepted", err)
	}
	live.Close()
	if err = WithStopped(ctx, key, func() error {
		other, e := Listen(ctx, key, nil)
		if e == nil {
			other.Close()
			t.Fatal("new UI entered offline transaction")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	next, err := Listen(ctx, key, nil)
	if err != nil {
		t.Fatal("offline gate not released", err)
	}
	next.Close()
}
