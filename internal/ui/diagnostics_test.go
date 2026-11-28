package ui

import (
	"charm.land/lipgloss/v2"
	"context"
	"errors"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/diagnostics"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestDiagnosticsInspectWithoutChangingNativeClientOrForwardingInput(t *testing.T) {
	svc := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, svc)
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.listBusy })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	conn := h.m.workspace.conn
	generation := h.m.workspace.generation
	h.send(tea.KeyPressMsg{Code: 'D', Text: "D"})
	if !strings.Contains(h.m.View().Content, "Diagnostics") {
		t.Fatal("management D did not open diagnostics")
	}
	h.send(tea.PasteMsg{Content: "SECRET paste must not reach terminal"})
	h.send(tea.KeyPressMsg{Code: 'j', Mod: tea.ModAlt})
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if h.m.closed || h.m.workspace.active != 0 || h.m.workspace.conn != conn || h.m.workspace.generation != generation {
		t.Fatal("read-only diagnostics changed the original native view")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if strings.Contains(h.m.View().Content, "Diagnostics") || h.m.workspace.conn != conn {
		t.Fatal("closing diagnostics reconnected or retained the modal")
	}
}

func TestDiagnosticExportSurvivesClosingWithoutChangingOrReopeningView(t *testing.T) {
	svc := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	report := make(chan diagnostics.Report, 1)
	gate := make(chan struct{})
	h := newWorkspaceHarness(t, svc, Options{SidebarMode: "flat", ExportDiagnostics: func(ctx context.Context, r diagnostics.Report) (string, error) {
		report <- r
		select {
		case <-gate:
			return "diagnostic-1.json", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}})
	h.until(func() bool { return h.m.workspace.conn != nil && !h.m.listBusy })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: 'D', Text: "D"})
	h.send(tea.KeyPressMsg{Code: 'e', Text: "e"})
	select {
	case r := <-report:
		data, err := r.Encode()
		if err != nil || !strings.Contains(string(data), "events") {
			t.Fatal("invalid export", err)
		}
	case <-time.After(time.Second):
		t.Fatal("no export")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	conn := h.m.workspace.conn
	close(gate)
	h.until(func() bool { return !h.m.diagnostic.exportBusy })
	if strings.Contains(h.m.View().Content, "Diagnostics") || h.m.workspace.conn != conn {
		t.Fatal("late export reopened or reconnected the workspace")
	}
}

func TestDiagnosticsRemainBoundedAtNarrowWindowSizes(t *testing.T) {
	for _, size := range [][2]int{{160, 45}, {120, 32}, {63, 16}, {49, 12}, {12, 4}, {1, 1}} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			h, _ := attentionUIFixture(t)
			h.m.workspace.focused = false
			h.send(tea.KeyPressMsg{Code: 'D', Text: "D"})
			h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
			view := h.m.View()
			if view.Cursor != nil {
				t.Fatal("diagnostics exposed a terminal cursor")
			}
			if lipgloss.Width(view.Content) > size[0] || lipgloss.Height(view.Content) > size[1] {
				t.Fatal("diagnostics exceeded window boundary")
			}
		})
	}
}

func TestDiagnosticsIgnoreStaleReceiptsAndExcludeRawFailureText(t *testing.T) {
	h, _ := attentionUIFixture(t)
	w := h.m.workspace
	h.m.workspace.focused = false
	h.send(tea.KeyPressMsg{Code: 'D', Text: "D"})
	before := len(h.m.diagnosticReport().Events)
	h.send(workspaceFrame{generation: w.generation + 1, err: errors.New("STALE_SECRET")})
	if len(h.m.diagnosticReport().Events) != before {
		t.Fatal("stale Run contaminated diagnostics")
	}
	h.send(workspaceFrame{generation: w.generation, err: errors.New("RAW_SECRET C:\\private\\file env=token")})
	report := h.m.diagnosticReport()
	data, err := report.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if report.Client != "disconnected" || !strings.Contains(string(data), "operation_failed") {
		t.Fatal("fresh failure was not observable")
	}
	for _, v := range []string{"RAW_SECRET", "STALE_SECRET", "private", "env=token"} {
		if strings.Contains(string(data), v) || strings.Contains(h.m.View().Content, v) {
			t.Fatal("raw failure leaked")
		}
	}
}
