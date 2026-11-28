//go:build windows

package ui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/diagnostics"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func TestWindowsNativeDiagnosticsKeepOriginalClientAndExcludePrivateContent(t *testing.T) {
	bundle, fixture := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_MESSAGING_FIXTURE")
	if bundle == "" || fixture == "" {
		t.Skip("explicit isolated native fixture required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	root := t.TempDir()
	id, _ := session.NewID()
	opts := backend.Options{TerminalBundle: bundle, RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-diag-" + id[:12]}
	b, err := backend.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	info, err := b.Probe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	kind, location, _ := backend.Location(opts)
	db, err := store.Open(ctx, filepath.Join(root, "state", "metadata.sqlite3"), strings.ToLower(kind+":"+location))
	if err != nil {
		t.Fatal(err)
	}
	svc := manager.New(b, db)
	original, err := svc.Start(ctx, session.Launch{Name: "PRIVATE_DIAGNOSTIC_TITLE", Executable: fixture, Dir: root, Env: []string{"ASD_FIXTURE_ENV_KEYS=ASD_DIAG_PRIVATE", "ASD_DIAG_PRIVATE=PRIVATE_DIAGNOSTIC_SENTINEL"}})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	var h *workspaceHarness
	defer func() {
		if h != nil {
			h.m.Close()
		}
		c, x := context.WithTimeout(context.Background(), 15*time.Second)
		defer x()
		if e := svc.Stop(c, original); e != nil {
			t.Error(e)
		}
		if e := db.Close(); e != nil {
			t.Error(e)
		}
	}()
	for {
		p, e := b.Preview(ctx, original, 100)
		if e == nil && strings.Contains(p.Text, "PRIVATE_DIAGNOSTIC_SENTINEL") {
			break
		}
		if ctx.Err() != nil {
			t.Fatal("private marker did not reach original native screen")
		}
		time.Sleep(50 * time.Millisecond)
	}
	h = newWorkspaceHarness(t, svc, Options{SidebarMode: "flat", DiagnosticRuntime: diagnostics.CurrentRuntime(info.Name, info.Version), ExportDiagnostics: func(ctx context.Context, r diagnostics.Report) (string, error) {
		return diagnostics.Export(ctx, filepath.Join(root, "reports"), r)
	}})
	until := func(check func() bool) {
		t.Helper()
		deadline := time.NewTimer(15 * time.Second)
		defer deadline.Stop()
		for !check() {
			select {
			case msg := <-h.messages:
				h.send(msg)
			case <-deadline.C:
				t.Fatal("native diagnostic phase did not complete")
			}
		}
	}
	until(func() bool { return h.m.workspace.conn != nil && !h.m.listBusy && h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	until(func() bool { return h.m.workspace.pending == 0 })
	conn, generation := h.m.workspace.conn, h.m.workspace.generation
	h.send(tea.KeyPressMsg{Code: 'D', Text: "D"})
	if !strings.Contains(h.m.View().Content, "Diagnostics") {
		t.Fatal("no diagnostic view")
	}
	h.send(tea.PasteMsg{Content: "DIAGNOSTIC_MUST_NOT_REACH_TERMINAL"})
	h.send(tea.KeyPressMsg{Code: 'e', Text: "e"})
	until(func() bool { return !h.m.diagnostic.exportBusy })
	if h.m.diagnostic.exportPath == "" {
		t.Fatal("diagnostic export failed")
	}
	data, err := os.ReadFile(h.m.diagnostic.exportPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []string{"PRIVATE_DIAGNOSTIC_TITLE", "PRIVATE_DIAGNOSTIC_SENTINEL", root, fixture, original.ID, original.RunID, "DIAGNOSTIC_MUST_NOT_REACH_TERMINAL"} {
		if strings.Contains(string(data), v) {
			t.Fatal("diagnostic report contained private input")
		}
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if h.m.workspace.conn != conn || h.m.workspace.generation != generation {
		t.Fatal("inspecting/exporting reconnected the original client")
	}
	rows, err := b.List(ctx)
	if err != nil || len(rows) != 1 || rows[0].ID != original.ID || rows[0].RunID != original.RunID || rows[0].BackendID != original.BackendID {
		t.Fatal("original terminal identity changed")
	}
	page, err := b.Preview(ctx, original, 100)
	if err != nil || strings.Contains(page.Text, "DIAGNOSTIC_MUST_NOT_REACH_TERMINAL") {
		t.Fatal("diagnostic input reached original terminal")
	}
	t.Log("real fixture private marker verified; sanitized export; original connection/Session/Run/backend retained; no diagnostic input written")
}
