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
	err     error
}

func (m *Model) openLaunchDialog() tea.Cmd {
	if _, ok := m.service.(launchService); !ok {
		m.status = "当前接入不支持创建终端"
		return nil
	}
	d := &launchDialog{firstSession: len(m.sessions) == 0}
	group := m.launchGroup
	m.launchGroup = session.Group{}
	d.groupID = group.ID
	placeholders := []string{"可留空，使用程序名", "例如 pwsh.exe、cmd.exe 或 agent 路径", "可留空，使用当前工作目录", `例如 ["-NoLogo","-NoProfile"]`}
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
				d.err = fmt.Errorf("启动结果需核对 Session %s / Run %s：%w；不会自动重试", result.session.ID, result.session.RunID, result.err)
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
		return m.acceptLaunchedSession(result.session), true
	}
	d := m.launch
	if d == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		m.sizeLaunchDialog()
		return nil, true
	case tea.KeyReleaseMsg:
		return nil, true
	case tea.KeyPressMsg:
		if d.busy {
			return nil, true
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
			d.err = fmt.Errorf("字段超过字节上限")
			return nil, true
		}
		var cmd tea.Cmd
		d.fields[d.focus], cmd = d.fields[d.focus].Update(msg)
		return cmd, true
	case tea.PasteMsg:
		if d.busy || d.focus >= len(d.fields) {
			return nil, true
		}
		if strings.ContainsAny(msg.Content, "\r\n\t") {
			d.err = fmt.Errorf("此字段需要单行文本；终端内仍可多行粘贴")
			return nil, true
		}
		if len(d.fields[d.focus].Value())+len(msg.Content) > launchFieldLimit {
			d.err = fmt.Errorf("字段超过字节上限")
			return nil, true
		}
		var cmd tea.Cmd
		d.fields[d.focus], cmd = d.fields[d.focus].Update(msg)
		return cmd, true
	}
	return nil, false
}

func (m *Model) focusLaunchField(index int) tea.Cmd {
	d := m.launch
	for i := range d.fields {
		d.fields[i].Blur()
	}
	d.focus = index
	if index < len(d.fields) {
		return d.fields[index].Focus()
	}
	return nil
}

func (m *Model) cancelLaunchDialog() tea.Cmd {
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
			d.err = fmt.Errorf("参数需要 JSON 字符串数组")
			return nil
		}
		if err := json.Unmarshal([]byte(raw), &l.Args); err != nil {
			d.err = fmt.Errorf("参数 JSON 无效：%w", err)
			return nil
		}
	}
	if err := l.Validate(); err != nil {
		d.err = err
		return nil
	}
	d.busy = true
	d.err = nil
	service, ctx := m.service.(launchService), m.ctx
	return func() tea.Msg { s, err := service.Start(ctx, l); return launchResult{d, s, err} }
}
