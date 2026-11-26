package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/liuhui-code/agent-session-desk/internal/pet"
)

func petDemo(ctx context.Context, args []string, errorOutput io.Writer) error {
	f := flag.NewFlagSet("pet-demo", flag.ContinueOnError)
	f.SetOutput(errorOutput)
	state := f.String("state", "idle", "demo state: idle, approval, question, unknown")
	count := f.Int("count", 1, "synthetic request count (1-99)")
	duration := f.Duration("duration", 0, "auto-close after this duration, at most 5m")
	report := f.String("report", "", "optional new JSON diagnostics file")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("pet-demo accepts no positional arguments")
	}
	return pet.Run(ctx, pet.Options{State: pet.Kind(*state), Count: *count, Duration: *duration, Report: *report})
}
