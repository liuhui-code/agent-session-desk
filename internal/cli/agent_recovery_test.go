package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
)

func TestRecoveryProfileAndStatusAreOffline(t *testing.T) {
	var out, diag bytes.Buffer
	base := []string{"--backend", filepath.Join(t.TempDir(), "missing-backend"), "--runtime-dir", t.TempDir(), "--state-dir", t.TempDir()}
	for _, command := range []string{"recovery-profile", "recovery-status"} {
		out.Reset()
		diag.Reset()
		args := append(append([]string{}, base...), command, "--session", "11111111111111111111111111111111")
		err := run(context.Background(), args, bytes.NewReader(nil), &out, &diag)
		if err == nil || !bytes.Contains([]byte(err.Error()), []byte("unavailable")) {
			t.Fatalf("offline %s: %v", command, err)
		}
	}
}
