package appearance_test

import (
	"context"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/appearance"
)

func TestTerminalDeclarationProducesOneKnownThemeThenStopsOnContext(t *testing.T) {
	t.Setenv("COLORFGBG", "15;0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, e := appearance.Open(ctx)
	if e != nil {
		t.Fatalf("valid terminal declaration unavailable: %v", e)
	}
	defer c.Close()
	v, e := c.Next(ctx)
	if e != nil || v.Theme != "dark" || v.Source != "terminal" || v.Err != nil {
		t.Fatalf("declared terminal theme: %+v %v", v, e)
	}
	cancel()
	if _, e = c.Next(ctx); e == nil {
		t.Fatal("cancelled observer did not stop")
	}
}
