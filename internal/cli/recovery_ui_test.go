package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/recovery"
)

func TestRecoveryUIContextCapacityPreservesOccupiedSlots(t *testing.T) {
	root := t.TempDir()
	for i := 1; i <= recoveryUIContextSlots; i++ {
		// Occupied entries of any type remain occupied; do not follow or overwrite.
		name := filepath.Join(root, fmt.Sprintf("recovery-context-%02d", i))
		if err := os.WriteFile(name, []byte("preserve"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	_, err := prepareUIRecoveryContext(backend.Options{RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-bounded"}, root, recovery.Profile{}, "")
	if err == nil || !strings.Contains(err.Error(), "slots occupied") {
		t.Fatal("capacity not enforced", err)
	}
	entries, e := os.ReadDir(root)
	if e != nil || len(entries) != recoveryUIContextSlots {
		t.Fatal("occupied context mutated", e)
	}
}
