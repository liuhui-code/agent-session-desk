package ui

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/document"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type markdownService interface {
	ResolveMarkdown(context.Context, session.Session, string) (string, error)
}
type documentOpen struct {
	generation uint64
	target     session.Session
	ctx        context.Context
	cancel     context.CancelFunc
}
type documentResolved struct {
	operation *documentOpen
	path      string
	err       error
}
type documentOpened struct {
	operation *documentOpen
	err       error
}

func (m *Model) startDocumentClick(msg tea.MouseClickMsg) (tea.Cmd, bool) {
	w := m.workspace
	locks := tea.ModCapsLock | tea.ModNumLock | tea.ModScrollLock
	if msg.Button != tea.MouseLeft || msg.Mod & ^locks != tea.ModCtrl || !w.focused || w.conn == nil || w.busy || w.err != nil || w.transition != workspaceIdle || len(w.tabs) == 0 || w.programPointer != nil || w.copyPointer != nil {
		return nil, false
	}
	g := m.workspaceGeometry()
	x, y := msg.X-g.x, msg.Y-g.y
	width, height := min(g.width, w.screen.Width), min(g.height, w.screen.Height)
	if x < 0 || x >= width || y < 0 || y >= height || y >= len(w.screen.Rows) || w.screen.CopyPrompt != nil {
		return nil, false
	}
	line := func(row int) string {
		var out strings.Builder
		for _, run := range w.screen.Rows[row].Runs {
			if run.Flags&(1<<6) != 0 {
				out.WriteString(strings.Repeat(" ", run.Width))
				continue
			}
			out.WriteString(strings.Map(func(r rune) rune {
				if unicode.IsControl(r) {
					return -1
				}
				return r
			}, run.Text))
		}
		return out.String()
	}
	text := strings.TrimRight(line(y), " ")
	// The transport has no soft-wrap map. Never guess from a clipped/wrapped row.
	if ansi.StringWidth(text) >= width || (y > 0 && ansi.StringWidth(strings.TrimRight(line(y-1), " ")) >= width) {
		return nil, false
	}
	ref, ok := document.At(text, x)
	if !ok {
		return nil, false
	}
	w.documentPointer = true
	if w.document != nil {
		m.status = "Markdown open already in progress"
		return nil, true
	}
	service, ok := m.service.(markdownService)
	if !ok || m.openDocument == nil {
		m.status = "External Markdown opening unavailable"
		return nil, true
	}
	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Second)
	op := &documentOpen{generation: w.generation, target: w.tabs[w.active].session, ctx: ctx, cancel: cancel}
	w.document = op
	m.status = "Opening Markdown…"
	return func() tea.Msg {
		path, err := service.ResolveMarkdown(ctx, op.target, ref)
		return documentResolved{op, path, err}
	}, true
}
func (m *Model) documentEvent(msg tea.Msg) (tea.Cmd, bool) {
	w := m.workspace
	switch v := msg.(type) {
	case tea.MouseReleaseMsg:
		if w.documentPointer && (v.Button == tea.MouseLeft || v.Button == tea.MouseNone) {
			w.documentPointer = false
			return nil, true
		}
	case tea.MouseMotionMsg:
		if w.documentPointer {
			return nil, true
		}
	case documentResolved:
		op := v.operation
		if op == nil || w.document != op {
			return nil, true
		}
		if v.err == nil && (op.generation != w.generation || w.conn == nil || w.err != nil || w.transition != workspaceIdle || len(w.tabs) == 0 || !sameStopTarget(w.tabs[w.active].session, op.target)) {
			v.err = fmt.Errorf("original view changed")
		}
		if v.err != nil {
			op.cancel()
			w.document = nil
			m.status = "Could not open Markdown: " + label(v.err.Error())
			return nil, true
		}
		open := m.openDocument
		return func() tea.Msg { return documentOpened{op, open(op.ctx, v.path)} }, true
	case documentOpened:
		if v.operation == nil || w.document != v.operation {
			return nil, true
		}
		v.operation.cancel()
		w.document = nil
		if v.err != nil {
			m.status = "Could not open Markdown: " + label(v.err.Error())
		} else {
			m.status = "Markdown open requested in the external editor"
		}
		return nil, true
	}
	return nil, false
}
