package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/keybindings"
	"strings"
)

func keyMessage(s string) tea.KeyPressMsg {
	parts := strings.Split(s, "+")
	if s == "+" {
		parts = []string{"+"}
	}
	k := tea.KeyPressMsg{}
	for _, v := range parts[:len(parts)-1] {
		switch v {
		case "ctrl":
			k.Mod |= tea.ModCtrl
		case "alt":
			k.Mod |= tea.ModAlt
		case "shift":
			k.Mod |= tea.ModShift
		}
	}
	name := parts[len(parts)-1]
	codes := map[string]rune{"space": tea.KeySpace, "enter": tea.KeyEnter, "tab": tea.KeyTab, "esc": tea.KeyEscape, "backspace": tea.KeyBackspace, "up": tea.KeyUp, "down": tea.KeyDown, "left": tea.KeyLeft, "right": tea.KeyRight, "home": tea.KeyHome, "end": tea.KeyEnd, "pgup": tea.KeyPgUp, "pgdown": tea.KeyPgDown, "f1": tea.KeyF1, "f2": tea.KeyF2, "f3": tea.KeyF3, "f4": tea.KeyF4, "f5": tea.KeyF5, "f6": tea.KeyF6, "f7": tea.KeyF7, "f8": tea.KeyF8, "f9": tea.KeyF9, "f10": tea.KeyF10, "f11": tea.KeyF11}
	if n, ok := codes[name]; ok {
		k.Code = n
	} else {
		for _, r := range name {
			k.Code = r
			break
		}
	}
	return k
}
func (m *Model) resolveWorkspaceKey(msg tea.KeyPressMsg) (tea.KeyPressMsg, bool) {
	w := m.workspace
	scope := "management"
	if w.focused && len(w.tabs) > 0 {
		scope = "workspace"
	}
	key, e := keybindings.Normalize(msg.String())
	if e != nil {
		return msg, false
	}
	canonical, mapped, off := w.bindings.Translate(scope, key)
	if mapped {
		return keyMessage(canonical), false
	}
	return msg, off
}
func (m *Model) configuredHelp() []helpSection {
	if m.workspace == nil || len(m.workspace.bindings.Overrides) == 0 {
		return workspaceHelpSections()
	}
	out := []helpSection{workspaceHelpSections()[0]}
	for _, scope := range []string{"global", "workspace", "management"} {
		s := helpSection{title: scope}
		for _, d := range keybindings.Definitions {
			if d.Scope != scope {
				continue
			}
			keys := m.workspace.bindings.Keys(d)
			text := strings.Join(keys, " / ")
			if len(keys) == 0 {
				text = "Disabled"
			}
			s.items = append(s.items, [2]string{text, d.Action})
		}
		out = append(out, s)
	}
	return append(out, helpSection{title: "Modal controls", items: [][2]string{{"Esc", "Back"}, {"Tab", "Cycle fields / help"}, {"Enter", "Confirm"}}})
}

func (m *Model) keyLabel(scope, action string) string {
	for _, d := range keybindings.Definitions {
		if d.Scope == scope && d.Action == action {
			v := m.workspace.bindings.Keys(d)
			if len(v) == 0 {
				return "—"
			}
			return strings.Join(v, "/")
		}
	}
	return "—"
}
