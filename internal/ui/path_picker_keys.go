// Adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: newdialog.go and group_dialog.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0; generic CLI fields and v2 messages.
package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/pathsuggest"
)

func (m *Model) pathPickerKey(p *pathPicker, key string, text string) (tea.Cmd, bool) {
	if key == "ctrl+c" || key == "ctrl+s" {
		return nil, false
	}
	if p == nil || !m.pathPopupAvailable() {
		return nil, false
	}
	if !p.launch {
		if p.hidden || len(p.matches) == 0 {
			return nil, false
		}
		switch key {
		case "up", "ctrl+p":
			p.navigated = true
			p.selected = (p.selected + len(p.matches) - 1) % len(p.matches)
			return nil, true
		case "down", "ctrl+n":
			p.navigated = true
			p.selected = (p.selected + 1) % len(p.matches)
			return nil, true
		case "tab", "enter":
			m.setPathValue(p, p.matches[p.selected])
			p.hidden = true
			return nil, true
		case "esc":
			p.hidden = true
			return nil, true
		}
		return nil, false
	}
	if p.active && !p.hidden {
		total := len(p.matches) + 1
		switch key {
		case "down", "j", "ctrl+n":
			p.navigated = true
			p.selected = (p.selected + 1) % total
		case "up", "k", "ctrl+p":
			p.navigated = true
			p.selected = (p.selected + total - 1) % total
		case "left", "h", "esc":
			p.active = false
			return m.focusPathInput(p), true
		case "enter", "space":
			p.hidden, p.active, p.soft = true, false, false
			if p.selected > 0 {
				m.setPathValue(p, p.matches[p.selected-1])
				if key == "enter" {
					return m.focusLaunchField(3), true
				}
			}
			return m.focusPathInput(p), true
		case "tab", "shift+tab":
			p.hidden, p.active = true, false
			step := 1
			if key == "shift+tab" {
				step = -1
			}
			return m.focusLaunchField(2 + step), true
		}
		return nil, true
	}
	if p.soft {
		switch key {
		case "up":
			return m.focusLaunchField(1), true
		case "down":
			return m.focusLaunchField(3), true
		case "left":
			p.soft = false
			return m.focusPathInput(p), true
		case "backspace", "delete":
			p.soft = false
			m.setPathValue(p, "")
			cmd := m.focusPathInput(p)
			return tea.Batch(cmd, m.pathEdited(p, "")), true
		case "right", "space":
			p.active, p.hidden, p.soft = true, false, false
			m.launch.fields[2].Blur()
			return nil, true
		}
		if text != "" {
			m.setPathValue(p, "")
			p.soft = false
			m.focusPathInput(p)
			return nil, false
		}
	}
	if len(p.matches) > 0 && (key == "ctrl+n" || key == "ctrl+p") {
		p.soft, p.hidden = false, false
		p.navigated = true
		total := len(p.matches) + 1
		delta := 1
		if key == "ctrl+p" {
			delta = total - 1
		}
		p.selected = (p.selected + delta) % total
		return m.focusPathInput(p), true
	}
	if p.navigated && p.selected > 0 && (key == "tab" || key == "enter") {
		m.setPathValue(p, p.matches[p.selected-1])
		p.hidden = true
		return m.focusLaunchField(3), true
	}
	if !p.hidden && !p.soft && len(p.matches) > 0 {
		switch key {
		case "up", "down":
			p.active = true
			p.navigated = true
			m.launch.fields[2].Blur()
			if key == "down" || key == "ctrl+n" {
				p.selected = 1
			} else {
				p.selected = len(p.matches)
			}
			return nil, true
		}
	}
	if key == "esc" && !p.hidden {
		p.hidden = true
		return nil, true
	}
	if key == "tab" {
		if len(p.cycle) > 0 {
			return m.completePathTab(p, p.cycle), true
		}
		if m.pathValue(p) != "" && pathsuggest.NativeQuery(m.pathValue(p)) && !p.exact {
			if p.ready && p.scanned {
				return m.completePathTab(p, p.disk), true
			}
			p.tabPending = true
			cmd := m.requestPaths(p, m.pathValue(p), true)
			return cmd, true
		}
	}
	return nil, false
}
func (m *Model) completePathTab(p *pathPicker, matches []string) tea.Cmd {
	if p.exact {
		return m.focusLaunchField(3)
	}
	if len(matches) == 0 {
		return nil
	}
	if len(p.cycle) == 0 {
		p.cycle = append([]string(nil), matches...)
		p.cycleIndex = -1
	}
	p.cycleIndex = (p.cycleIndex + 1) % len(p.cycle)
	m.setPathValue(p, p.cycle[p.cycleIndex])
	p.matches = p.cycle
	p.selected = p.cycleIndex + 1
	p.hidden = false
	return nil
}

// Preserve events typed after an explicit asynchronous Tab, in field order.
// Escape cancels this continuation immediately; queued input never reaches a terminal.
func (m *Model) deferPathTab(msg tea.Msg) bool {
	d := m.launch
	if d == nil || d.focus != 2 || !d.paths.tabPending {
		return false
	}
	p := d.paths
	size := 0
	switch v := msg.(type) {
	case tea.KeyPressMsg:
		if v.String() == "esc" || v.String() == "ctrl+c" {
			p.tabPending = false
			p.deferred = nil
			p.deferredBytes = 0
			return false
		}
		size = len(v.Text)
	case tea.PasteMsg:
		size = len(v.Content)
	default:
		return false
	}
	if len(p.deferred) >= 16 || p.deferredBytes+size > launchFieldLimit {
		d.err = fmt.Errorf("Directory completion pending; buffered input limit reached")
		return true
	}
	p.deferred = append(p.deferred, msg)
	p.deferredBytes += size
	return true
}
