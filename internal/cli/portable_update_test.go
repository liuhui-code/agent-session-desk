package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/liuhui-code/agent-session-desk/internal/updater"
	"testing"
)

func TestUpdateProbeReadsNoBackendOrPrivateState(t *testing.T) {
	var out, diag bytes.Buffer
	if code := Run(context.Background(), []string{"update", "probe"}, nil, &out, &diag); code != 0 {
		t.Fatal(code, diag.String())
	}
	var p updater.Probe
	if json.Unmarshal(out.Bytes(), &p) != nil || p.Protocol != 1 || p.Compatibility.Metadata != 11 {
		t.Fatal("invalid update capability", out.String())
	}
}
func TestUpdateApplyRequiresExplicitClosedConfirmation(t *testing.T) {
	var out, diag bytes.Buffer
	if code := Run(context.Background(), []string{"update", "apply", "--install-dir", t.TempDir()}, nil, &out, &diag); code == 0 {
		t.Fatal("apply without confirmation accepted")
	}
	if !bytes.Contains(diag.Bytes(), []byte("confirm-closed")) {
		t.Fatal("unclear refusal", diag.String())
	}
}
