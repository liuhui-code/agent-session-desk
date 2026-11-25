// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/group_dialog.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: stable v2 input, product metadata labels and exact Session/Run
// binding. See third_party/agent-deck/LICENSE.
package ui

import (
	"context"
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type renameService interface {
	Rename(context.Context, session.Session, string) (session.Session, error)
}
type sessionRename struct {
	input           textinput.Model
	target          session.Session
	busy, submitted bool
	err             error
}
type renameResult struct {
	dialog  *sessionRename
	session session.Session
	err     error
}

func (m *Model) beginSessionRename() tea.Cmd {
	if _, ok := m.service.(renameService); !ok || m.workspace.transition != workspaceIdle {
		return nil
	}
	target, ok := m.selected()
	if !ok {
		return nil
	}
	m.renameTarget = &target
	return m.navigateWorkspace(workspaceRename)
}

func (m *Model) openSessionRename() tea.Cmd {
	target := m.renameTarget
	m.renameTarget = nil
	if target == nil {
		return m.connectWorkspace()
	}
	d := &sessionRename{input: textinput.New(), target: *target}
	d.input.CharLimit = session.MaxTitleInputBytes
	d.input.Prompt = m.tokens.rename.prompt
	d.input.SetVirtualCursor(false)
	d.input.SetStyles(m.styles.renameInput)
	d.input.KeyMap.Paste.SetEnabled(false)
	d.input.SetValue(target.Title)
	d.input.CursorEnd()
	m.rename = d
	m.workspace.focused = false
	m.sizeSessionRename()
	return d.input.Focus()
}

func (m *Model) sessionRenameEvent(msg tea.Msg) (tea.Cmd, bool) {
	if result, ok := msg.(renameResult); ok {
		if m.rename != result.dialog {
			return nil, true
		}
		d := m.rename
		d.busy, d.err = false, result.err
		if result.err != nil {
			d.err = fmt.Errorf("Go back and refresh the name before editing: %w", result.err)
			return nil, true
		}
		m.listGeneration++ // a pre-commit discovery must not restore the old label
		if m.listCancel != nil {
			m.listCancel()
		}
		m.acceptSessionName(result.session)
		m.rename = nil
		m.status = ""
		// The new name also changes discovery order and its reviewed stamp.
		// Reconcile both from the service for subsequent reviewed moves.
		return tea.Batch(m.connectWorkspace(), m.loadSessions()), true
	}
	d := m.rename
	if d == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		m.sizeSessionRename()
		return nil, true
	case tea.KeyReleaseMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
		return nil, true
	case tea.KeyPressMsg:
		if d.busy {
			return nil, true
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			m.rename = nil
			return tea.Batch(m.connectWorkspace(), m.loadSessions()), true
		case "enter":
			return m.submitSessionRename(), true
		case "tab", "shift+tab", "ctrl+pgup", "ctrl+pgdown", "f6":
			return nil, true
		}
		if d.submitted || !renameInputSafe(d.input.Value(), msg.Text) {
			return nil, true
		}
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(msg)
		d.err = nil
		return cmd, true
	case tea.PasteMsg:
		if d.busy || d.submitted {
			return nil, true
		}
		if !renameInputSafe(d.input.Value(), msg.Content) {
			d.err = fmt.Errorf("Name must be bounded single-line text")
			return nil, true
		}
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(msg)
		d.err = nil
		return cmd, true
	}
	return nil, false
}

func renameInputSafe(current, added string) bool {
	if len(current)+len(added) > session.MaxTitleInputBytes || !utf8.ValidString(added) {
		return false
	}
	for _, r := range added {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return false
		}
	}
	return true
}

func (m *Model) submitSessionRename() tea.Cmd {
	d := m.rename
	if d.submitted {
		return nil
	}
	name, err := session.NormalizeRenameTitle(d.input.Value())
	if err != nil {
		d.err = err
		return nil
	}
	found := false
	for _, s := range m.sessions {
		if sameRun(s, d.target) && s.BackendID == d.target.BackendID && s.Command == d.target.Command {
			found = true
			break
		}
	}
	if !found {
		d.err = fmt.Errorf("Session/Run changed; go back and select again")
		return nil
	}
	d.busy, d.submitted, d.err = true, true, nil
	service, ctx := m.service.(renameService), m.ctx
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		v, err := service.Rename(ctx, d.target, name)
		return renameResult{d, v, err}
	}
}

func (m *Model) acceptSessionName(named session.Session) {
	for i := range m.sessions {
		v := &m.sessions[i]
		if v.ID == named.ID && v.TitleRevision <= named.TitleRevision {
			v.Title, v.TitleRevision = named.Title, named.TitleRevision
		}
	}
	for i := range m.workspace.tabs {
		v := &m.workspace.tabs[i].session
		if v.ID == named.ID && v.TitleRevision <= named.TitleRevision {
			v.Title, v.TitleRevision = named.Title, named.TitleRevision
		}
	}
}
