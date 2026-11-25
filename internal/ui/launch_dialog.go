// Layout adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/newdialog.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: generic CLI fields, stable v2 text inputs, token styles and bounded
// viewport; no brand/provider or worktree dependency. See THIRD_PARTY_NOTICES.md.
package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/pathsuggest"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const launchFieldLimit = 64 * 1024 // defensive form input bound, not a style value

type launchService interface {
	Start(context.Context, session.Launch) (session.Session, error)
}
type launchDialog struct {
	paths           *pathPicker
	groupID         string
	firstSession    bool
	fields          []textinput.Model
	focus           int
	busy, uncertain bool
	err             error
}
type launchResult struct {
	dialog  *launchDialog
	session session.Session
	dir     string
	err     error
}

func (m *Model) openLaunchDialog() tea.Cmd {
	if _, ok := m.service.(launchService); !ok {
		m.status = "This backend cannot create terminals"
		return nil
	}
	d := &launchDialog{firstSession: len(m.sessions) == 0, paths: m.newPathPicker(true)}
	group := m.launchGroup
	m.launchGroup = session.Group{}
	d.groupID = group.ID
	placeholders := []string{"Optional; use the program name", "e.g. pwsh.exe, cmd.exe or an agent executable", "Optional; use the current directory", `例如 ["-NoLogo","-NoProfile"]`, `例如 ["ASD_MODE=dev"]`}
	for _, placeholder := range placeholders {
		field := textinput.New()
		field.CharLimit = launchFieldLimit
		field.Placeholder = placeholder
		field.Prompt = m.tokens.marks.selected + strings.Repeat(" ", m.tokens.layout.menuGap)
		field.SetVirtualCursor(false)
		field.SetStyles(m.styles.formInput)
		field.KeyMap.Paste.SetEnabled(false) // outer terminal provides explicit PasteMsg
		d.fields = append(d.fields, field)
	}
	d.fields[2].SetValue(group.DefaultDir)
	m.launch = d
	m.workspace.focused = false
	m.status = ""
	if m.workspace.pending == 0 && !m.workspace.busy {
		m.workspace.disconnect()
	}
	m.sizeLaunchDialog()
	return d.fields[0].Focus()
}

func (m *Model) launchEvent(msg tea.Msg) (tea.Cmd, bool) {
	if result, ok := msg.(launchResult); ok {
		if m.launch != result.dialog {
			return nil, true
		}
		d := m.launch
		d.busy = false
		d.err = result.err
		if result.err != nil {
			d.uncertain = result.session.ID != ""
			if d.uncertain {
				d.err = fmt.Errorf("Verify launch outcome for Session %s / Run %s: %w; no automatic retry", result.session.ID, result.session.RunID, result.err)
			}
			return nil, true
		}
		// A discovery snapshot requested before Start may omit the new Run.
		// Cancel/invalidate it before adding the acknowledged live terminal.
		m.listGeneration++
		if m.listCancel != nil {
			m.listCancel()
		}
		m.launch = nil
		return tea.Batch(m.acceptLaunchedSession(result.session), m.rememberDirectory(result.dir)), true
	}
	d := m.launch
	if d == nil {
		return nil, false
	}
	if m.deferPathTab(msg) {
		return nil, true
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		m.sizeLaunchDialog()
		return m.reconcilePathViewport(), true
	case tea.KeyReleaseMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg, tea.MouseWheelMsg:
		return nil, true
	case tea.KeyPressMsg:
		if d.busy {
			return nil, true
		}
		if d.focus == 2 {
			if cmd, ok := m.pathPickerKey(d.paths, msg.String(), msg.Text); ok {
				return cmd, true
			}
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			return m.cancelLaunchDialog(), true
		case "ctrl+s":
			return m.submitLaunchDialog(), true
		case "tab", "shift+tab":
			step := 1
			if msg.String() == "shift+tab" {
				step = -1
			}
			return m.focusLaunchField((d.focus + step + len(d.fields) + 1) % (len(d.fields) + 1)), true
		case "enter":
			if d.focus >= len(d.fields) {
				return m.submitLaunchDialog(), true
			}
			return m.focusLaunchField(d.focus + 1), true
		}
		if d.focus >= len(d.fields) {
			return nil, true
		}
		if len(d.fields[d.focus].Value())+len(msg.Text) > launchFieldLimit {
			d.err = fmt.Errorf("Field exceeds the byte limit")
			return nil, true
		}
		var cmd tea.Cmd
		before := d.fields[d.focus].Value()
		d.fields[d.focus], cmd = d.fields[d.focus].Update(msg)
		if d.focus == 2 && d.fields[2].Value() != before {
			cmd = tea.Batch(cmd, m.pathEdited(d.paths, d.fields[2].Value()))
		}
		return cmd, true
	case tea.PasteMsg:
		if d.busy || d.focus >= len(d.fields) {
			return nil, true
		}
		if strings.ContainsAny(msg.Content, "\r\n\t") {
			d.err = fmt.Errorf("This field requires one line; multiline paste is still supported in the terminal")
			return nil, true
		}
		if len(d.fields[d.focus].Value())+len(msg.Content) > launchFieldLimit {
			d.err = fmt.Errorf("Field exceeds the byte limit")
			return nil, true
		}
		var cmd tea.Cmd
		if d.focus == 2 && d.paths.soft {
			d.fields[2].SetValue("")
			d.paths.soft = false
			d.fields[2].Focus()
		}
		before := d.fields[d.focus].Value()
		d.fields[d.focus], cmd = d.fields[d.focus].Update(msg)
		if d.focus == 2 && d.fields[2].Value() != before {
			cmd = tea.Batch(cmd, m.pathEdited(d.paths, d.fields[2].Value()))
		}
		return cmd, true
	}
	return nil, false
}

func (m *Model) focusLaunchField(index int) tea.Cmd {
	d := m.launch
	if d.focus == 2 && index != 2 {
		m.cancelPathPicker(d.paths)
	}
	for i := range d.fields {
		d.fields[i].Blur()
	}
	d.focus = index
	if index < len(d.fields) {
		cmd := d.fields[index].Focus()
		if index == 2 {
			d.paths.soft = d.fields[2].Value() != "" && m.pathPopupAvailable()
			if d.paths.soft {
				d.fields[2].Blur()
				cmd = nil
			}
			d.paths.active = false
			d.paths.hidden = false
			cmd = tea.Batch(cmd, m.requestPaths(d.paths, d.fields[2].Value()))
		}
		return cmd
	}
	return nil
}

func (m *Model) cancelLaunchDialog() tea.Cmd {
	m.cancelPathPicker(m.launch.paths)
	m.launch = nil
	w := m.workspace
	if w.conn != nil {
		if !w.readBusy {
			return m.readWorkspace()
		}
		return nil
	}
	if w.busy {
		return nil
	}
	return m.connectWorkspace()
}

func (m *Model) submitLaunchDialog() tea.Cmd {
	d := m.launch
	if d.busy || d.uncertain {
		return nil
	}
	l := session.Launch{GroupID: d.groupID, Name: d.fields[0].Value(), Executable: d.fields[1].Value(), Dir: d.fields[2].Value()}
	raw := strings.TrimSpace(d.fields[3].Value())
	if raw != "" {
		if !strings.HasPrefix(raw, "[") {
			d.err = fmt.Errorf("Arguments must be a JSON string array")
			return nil
		}
		if err := json.Unmarshal([]byte(raw), &l.Args); err != nil {
			d.err = fmt.Errorf("Invalid arguments JSON: %w", err)
			return nil
		}
	}
	raw = strings.TrimSpace(d.fields[4].Value())
	if raw != "" {
		if !strings.HasPrefix(raw, "[") {
			d.err = fmt.Errorf("Environment overrides must be a JSON string array")
			return nil
		}
		if err := json.Unmarshal([]byte(raw), &l.Env); err != nil {
			d.err = fmt.Errorf("Invalid environment JSON")
			return nil
		}
	}
	if err := l.Validate(); err != nil {
		d.err = err
		return nil
	}
	m.cancelPathPicker(d.paths)
	d.busy = true
	d.err = nil
	service, ctx := m.service.(launchService), m.ctx
	return func() tea.Msg {
		var err error
		l.Dir, err = pathsuggest.ExpandHome(l.Dir)
		if err != nil {
			return launchResult{dialog: d, err: err}
		}
		s, err := service.Start(ctx, l)
		if err != nil && len(l.Env) > 0 {
			err = environmentLaunchError(err)
		}
		dir := ""
		if l.Dir != "" {
			dir, _ = filepath.Abs(l.Dir)
		}
		return launchResult{dialog: d, session: s, dir: dir, err: err}
	}
}

// ResolveLaunch and validateTransport use these stable application prefixes.
// Keep field location while omitting paths, values and arbitrary backend stderr.
func environmentLaunchError(err error) error {
	text := err.Error()
	switch {
	case strings.HasPrefix(text, "executable not found or not executable:"):
		return fmt.Errorf("Program unavailable; check its path and PATH/PATHEXT overrides")
	case strings.HasPrefix(text, "working directory"):
		return fmt.Errorf("Working directory unavailable; check the path and permissions")
	case strings.HasPrefix(text, "environment key "):
		return fmt.Errorf("Environment overrides violate terminal rules; check names and values")
	default:
		return fmt.Errorf("Launch failed; check program, arguments, directory and environment")
	}
}
