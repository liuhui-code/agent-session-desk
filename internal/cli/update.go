package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/release"
	"io"
)

func checkUpdate(ctx context.Context, args []string, out io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("check-update takes no arguments")
	}
	v, e := release.Check(ctx)
	if e != nil {
		return e
	}
	return json.NewEncoder(out).Encode(struct {
		Current string `json:"current"`
		release.Info
	}{release.Build(), v})
}
