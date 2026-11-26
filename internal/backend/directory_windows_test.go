//go:build windows

package backend

import (
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsPsmuxNativeMarkdownDirectory(t *testing.T) {
	h := newNativeHarness(t)
	launch, report := h.launch("Markdown original cwd")
	original := h.start(launch)
	started := fixtureEvent(t, h.ctx, report, "started", "")
	got, err := h.b.CurrentDirectory(h.ctx, original)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(filepath.Clean(got), filepath.Clean(started.Cwd)) {
		t.Fatalf("pane directory %q != fixture cwd %q", got, started.Cwd)
	}
	wrong := original
	wrong.RunID = "ffffffffffffffffffffffffffffffff"
	if _, err := h.b.CurrentDirectory(h.ctx, wrong); err == nil {
		t.Fatal("different Run used original directory")
	}
	if original.TerminalState != session.TerminalAvailable || !filepath.IsAbs(got) {
		t.Fatal("invalid native directory")
	}
	t.Log("original fixture reports the same absolute directory; changed Run rejected; changing shell cwd is not covered")
}
