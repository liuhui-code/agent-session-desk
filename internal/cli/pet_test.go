package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPetDemoInvalidOptionsDoNotTouchBackendOrState(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	for _, args := range [][]string{{"pet-demo", "--state", "allow"}, {"pet-demo", "--count", "0"}, {"pet-demo", "--duration", "6m"}, {"pet-demo", "unexpected"}, {"--state-dir", state, "pet-demo"}} {
		var out, stderr bytes.Buffer
		if code := Run(context.Background(), args, nil, &out, &stderr); code == 0 {
			t.Fatal("invalid pet command accepted")
		}
	}
	if _, err := os.Stat(state); !os.IsNotExist(err) {
		t.Fatal("pet demo initialized backend state")
	}
}
