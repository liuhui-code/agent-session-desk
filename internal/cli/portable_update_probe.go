package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/updater"
)

// Bounded output; no shell, stderr or terminal input. This runs only after
// explicit apply/rollback and program hash checks in the update engine.
type probeBuffer struct{ data []byte }

func (b *probeBuffer) Write(p []byte) (int, error) {
	if len(b.data)+len(p) > 32*1024 {
		return 0, fmt.Errorf("update probe output exceeds bound")
	}
	b.data = append(b.data, p...)
	return len(p), nil
}
func probeUpdateProgram(ctx context.Context, path string) (updater.Probe, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var b probeBuffer
	cmd := exec.CommandContext(c, path, "update", "probe")
	cmd.WaitDelay = time.Second
	cmd.Stdout = &b
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return updater.Probe{}, fmt.Errorf("program update probe failed")
	}
	var p updater.Probe
	if json.Unmarshal(b.data, &p) != nil {
		return p, fmt.Errorf("invalid program update probe")
	}
	return p, nil
}
