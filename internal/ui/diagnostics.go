package ui

import (
	"context"
	"errors"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/release"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/diagnostics"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type diagnosticState struct {
	journal                      diagnostics.Journal
	runtime                      diagnostics.Runtime
	export                       func(context.Context, diagnostics.Report) (string, error)
	discoveryObserved            bool
	open, exportBusy             bool
	view, exportGeneration       uint64
	target                       session.Session
	offset, width, height        int
	message, exportPath          string
	discoveryStart, connectStart time.Time
	connectGeneration            uint64
}
type diagnosticExported struct {
	generation, view uint64
	path             string
	err              error
}

func (m *Model) beginDiagnostics() tea.Cmd {
	w := m.workspace
	if w == nil || (w.focused && len(w.tabs) > 0) || w.transition != workspaceIdle {
		return nil
	}
	d := &m.diagnostic
	d.open = true
	d.view++
	d.offset = 0
	d.message = ""
	d.exportPath = ""
	d.target, _ = m.selected()
	g := m.workspaceGeometry()
	d.width, d.height = g.width, g.height
	return nil
}
func (m *Model) diagnosticReport() diagnostics.Report {
	d := &m.diagnostic
	terminal, client := "unknown", "not_open"
	if m.listErr == nil {
		for _, s := range m.sessions {
			if activityKey(s) == activityKey(d.target) {
				terminal = string(s.TerminalState)
				break
			}
		}
	}
	if w := m.workspace; w != nil && w.active >= 0 && w.active < len(w.tabs) && activityKey(w.tabs[w.active].session) == activityKey(d.target) {
		switch {
		case w.busy:
			client = "connecting"
		case w.err != nil:
			client = "disconnected"
		case w.conn != nil:
			client = "connected"
		}
	}
	r := d.runtime
	if r.OS == "" {
		r = diagnostics.CurrentRuntime("", "")
		r.ProductVersion = release.Version
	}
	identity := ""
	if d.target.ID != "" {
		identity = activityKey(d.target)
	}
	return d.journal.Snapshot(r, identity, terminal, client)
}
func (m *Model) diagnosticEvent(msg tea.Msg) (tea.Cmd, bool) {
	d := &m.diagnostic
	if result, ok := msg.(diagnosticExported); ok {
		if result.generation != d.exportGeneration {
			return nil, true
		}
		d.exportBusy = false
		if d.open && result.view == d.view {
			if result.err != nil {
				d.message = "Export failed · " + diagnostics.Cause(result.err)
				if errors.Is(result.err, diagnostics.ErrSlotsFull) {
					d.message = "Four export slots occupied; remove an old report from the diagnostics directory"
				}
			} else {
				d.message = "Report saved"
				d.exportPath = result.path
			}
		}
		return nil, true
	}
	if !d.open {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		return nil, true
	case tea.KeyReleaseMsg, tea.PasteMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
		return nil, true
	case AttentionActivate, GlobalActivate:
		if m.attentionFocus != nil {
			m.attentionFocus()
		}
		d.message = "Close diagnostics before navigating to a pending item"
		return nil, true
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			d.open = false
			d.view++
			g := m.workspaceGeometry()
			if g.width != d.width || g.height != d.height {
				return m.workspaceSize(), true
			}
			return nil, true
		case "up", "k":
			d.offset = max(0, d.offset-1)
		case "down", "j":
			d.offset++
		case "pgdown":
			d.offset += m.diagnosticPageSize()
		case "pgup":
			d.offset = max(0, d.offset-m.diagnosticPageSize())
		case "home":
			d.offset = 0
		case "end":
			d.offset = len(m.diagnosticLines())
		case "e", "E":
			if d.exportBusy {
				return nil, true
			}
			if d.export == nil {
				d.message = "Export unavailable in this host"
				return nil, true
			}
			d.exportBusy = true
			d.exportGeneration++
			d.message = "Exporting report…"
			generation, view, report, export, ctx := d.exportGeneration, d.view, m.diagnosticReport(), d.export, m.ctx
			return func() tea.Msg {
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				path, err := export(ctx, report)
				return diagnosticExported{generation, view, path, err}
			}, true
		}
		return nil, true
	}
	return nil, false
}

// Observe only accepted operation receipts. No key, paste, output or raw error
// content is retained, and stale generations cannot contaminate another Run.
func (m *Model) observeDiagnostics(msg tea.Msg) {
	d := &m.diagnostic
	switch x := msg.(type) {
	case sessionsMsg:
		if x.generation == m.listGeneration && !m.attaching {
			d.discoveryObserved = true
			elapsed := time.Duration(0)
			if !d.discoveryStart.IsZero() {
				elapsed = time.Since(d.discoveryStart)
			}
			d.journal.Record(diagnostics.Discovery, "", elapsed, x.err)
		}
	case workspaceOpened:
		w := m.workspace
		if w == nil || x.generation != w.generation || len(w.tabs) == 0 {
			return
		}
		elapsed := time.Duration(0)
		if x.generation == d.connectGeneration && !d.connectStart.IsZero() {
			elapsed = time.Since(d.connectStart)
		}
		d.journal.Record(diagnostics.Connect, activityKey(w.tabs[w.active].session), elapsed, x.err)
	case workspaceFrame:
		if w := m.workspace; w != nil && x.generation == w.generation && x.err != nil && len(w.tabs) > 0 {
			d.journal.Record(diagnostics.Stream, activityKey(w.tabs[w.active].session), 0, x.err)
		}
	case workspaceInputDone:
		if w := m.workspace; w != nil && x.generation == w.generation && x.err != nil && len(w.tabs) > 0 {
			d.journal.Record(diagnostics.Input, activityKey(w.tabs[w.active].session), 0, x.err)
		}
	case tea.WindowSizeMsg:
		d.journal.Record(diagnostics.Resize, "", 0, nil)
	}
}
func (m *Model) diagnosticSummary() string {
	report := m.diagnosticReport()
	return fmt.Sprintf("Terminal: %s · Native view: %s · Agent state: unknown", report.Terminal, report.Client)
}
