//go:build windows

package pet

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

func TestWindowsIdleBlinkStillPaintsOpenAndClosedFrames(t *testing.T) {
	if os.Getenv("ASD_NATIVE_PET_INTERACTIVE") != "1" {
		t.Skip("explicit interactive desktop required")
	}
	tokens := ui.PetDesign()
	report := filepath.Join(t.TempDir(), "blink.json")
	duration := time.Duration(tokens.OpenMS+tokens.BlinkMS)*time.Millisecond + 500*time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), duration+5*time.Second)
	defer cancel()
	companion := NewCompanion(func(Activation) {})
	companion.Publish(nil, true)
	if err := Run(ctx, Options{Live: companion, Duration: duration, Report: report}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Paints int  `json:"paints"`
		State  Kind `json:"state"`
	}
	if err = json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Paints < 3 || result.State != Idle {
		t.Fatalf("blink timer was removed or changed state: %+v", result)
	}
}
