package ui

import (
	"context"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/snippets"
)

type snippetState struct {
	inflight *snippetDialog
	load     func(context.Context, session.Session) (snippets.Library, error)
	dialog   *snippetDialog
}
type snippetDialog struct {
	input                              textinput.Model
	target                             session.Session
	generation, viewID                 uint64
	cancel                             context.CancelFunc
	library                            snippets.Library
	rows                               []int
	cursor, offset                     int
	loading, preview                   bool
	returnFocused, returnSidebarHidden bool
	notice                             string
}
type snippetsLoaded struct {
	dialog  *snippetDialog
	library snippets.Library
	err     error
}

func (m *Model) snippetTargetCurrent(d *snippetDialog) bool {
	w := m.workspace
	if w == nil || w.conn == nil || w.busy || w.err != nil || w.transition != workspaceIdle || w.screen.CopyMode || w.screen.CopyPrompt != nil || m.listErr != nil || w.active < 0 || w.active >= len(w.tabs) {
		return false
	}
	tab := w.tabs[w.active]
	return w.generation == d.generation && tab.viewID == d.viewID && tab.live && sameStopTarget(tab.session, d.target) && tab.session.TerminalState == session.TerminalAvailable
}
func (m *Model) beginSnippets() tea.Cmd {
	w := m.workspace
	if w == nil || w.active < 0 || w.active >= len(w.tabs) || w.transition != workspaceIdle || m.snippets.dialog != nil {
		return nil
	}
	if m.snippets.inflight != nil {
		m.status = "Previous snippet load is finishing; wait before reopening"
		return nil
	}
	if w.pending > 0 {
		m.status = "Finish terminal input before opening snippets"
		return nil
	}
	d := &snippetDialog{target: w.tabs[w.active].session, generation: w.generation, viewID: w.tabs[w.active].viewID, returnFocused: w.focused, returnSidebarHidden: w.sidebarHidden}
	if !m.snippetTargetCurrent(d) {
		m.status = "Snippets require the active live terminal outside selection/search"
		return nil
	}
	d.input = textinput.New()
	d.input.CharLimit = searchQueryLimit
	d.input.Placeholder = "Search snippets…"
	d.input.Prompt = m.tokens.search.prompt
	d.input.SetVirtualCursor(false)
	d.input.SetStyles(searchInputStyles(m.tokens.forDisplay(m.colorProfile)))
	d.input.KeyMap.Paste.SetEnabled(false)
	m.snippets.dialog = d
	w.focused = false
	m.sizeSnippets()
	if m.snippets.load == nil {
		d.notice = "Snippet library unavailable"
		return d.input.Focus()
	}
	d.loading = true
	m.snippets.inflight = d
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	d.cancel = cancel
	load := m.snippets.load
	return tea.Batch(d.input.Focus(), func() tea.Msg { defer cancel(); lib, err := load(ctx, d.target); return snippetsLoaded{d, lib, err} })
}
func (m *Model) filterSnippets() {
	d := m.snippets.dialog
	d.rows = nil
	d.cursor = 0
	q := strings.ToLower(strings.TrimSpace(d.input.Value()))
	for i, s := range d.library.Items {
		if strings.Contains(strings.ToLower(s.Title+" "+s.ID+" "+s.Source+" "+s.Body), q) {
			d.rows = append(d.rows, i)
		}
	}
}
func (m *Model) closeSnippets(paste bool) tea.Cmd {
	d := m.snippets.dialog
	if d.cancel != nil {
		d.cancel()
	}
	m.snippets.dialog = nil
	w := m.workspace
	w.focused = d.returnFocused || paste
	w.sidebarHidden = d.returnSidebarHidden
	if w.conn != nil {
		return m.workspaceSize()
	}
	return nil
}
func (m *Model) snippetEvent(msg tea.Msg) (tea.Cmd, bool) {
	if v, ok := msg.(snippetsLoaded); ok {
		if m.snippets.inflight == v.dialog {
			m.snippets.inflight = nil
		}
		d := m.snippets.dialog
		if d == nil || v.dialog != d {
			return nil, true
		}
		d.loading = false
		if v.err != nil {
			d.notice = "Could not load snippets; check snippets.toml format and limits"
			return nil, true
		}
		d.library = v.library
		m.filterSnippets()
		return nil, true
	}
	d := m.snippets.dialog
	if d == nil {
		return nil, false
	}
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, v.Width), max(0, v.Height)
		m.sizeSnippets()
		return nil, true
	case tea.KeyReleaseMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
		return nil, true
	case tea.KeyPressMsg:
		switch v.String() {
		case "esc", "ctrl+c":
			if d.preview {
				d.preview = false
				d.offset = 0
				d.notice = ""
				return d.input.Focus(), true
			}
			return m.closeSnippets(false), true
		case "enter":
			if m.width < m.tokens.snippets.minWidth || m.height < m.tokens.snippets.minHeight {
				d.notice = "Enlarge the window before confirming a snippet"
				return nil, true
			}
			if d.loading || d.cursor < 0 || d.cursor >= len(d.rows) {
				return nil, true
			}
			if !d.preview {
				d.preview = true
				d.offset = 0
				d.input.Blur()
				return nil, true
			}
			if !m.snippetTargetCurrent(d) {
				d.notice = "Original terminal changed; close and choose again. Nothing pasted"
				return nil, true
			}
			body := d.library.Items[d.rows[d.cursor]].Body
			cmd := m.closeSnippets(true)
			m.status = "Snippet paste queued; no Enter added"
			return tea.Batch(cmd, m.workspaceWrite(workspaceInput{kind: "paste", text: body})), true
		case "up", "ctrl+k":
			if d.preview {
				d.offset = max(0, d.offset-1)
			} else {
				d.cursor = max(0, d.cursor-1)
			}
			return nil, true
		case "down", "ctrl+j":
			if d.preview {
				d.offset++
			} else {
				d.cursor = min(max(0, len(d.rows)-1), d.cursor+1)
			}
			return nil, true
		case "pgup":
			if d.preview {
				d.offset = max(0, d.offset-m.tokens.snippets.previewRows)
			}
			return nil, true
		case "pgdown":
			if d.preview {
				d.offset += m.tokens.snippets.previewRows
			}
			return nil, true
		case "tab", "shift+tab", "f6":
			return nil, true
		}
		if d.preview {
			return nil, true
		}
		if len(d.input.Value())+len(v.Text) > searchQueryLimit*4 {
			return nil, true
		}
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(v)
		m.filterSnippets()
		return cmd, true
	case tea.PasteMsg:
		if d.preview || strings.ContainsAny(v.Content, "\r\n\t") || len(d.input.Value())+len(v.Content) > searchQueryLimit*4 {
			return nil, true
		}
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(v)
		m.filterSnippets()
		return cmd, true
	}
	return nil, false
}
