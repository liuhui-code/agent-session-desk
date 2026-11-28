package keybindings

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Definition struct {
	Scope, Action string
	Keys          []string
}

var Definitions = []Definition{
	{"global", "wake", []string{"ctrl+alt+space"}},
	{"workspace", "snippets", []string{"f7"}},
	{"workspace", "previous", []string{"alt+p"}}, {"workspace", "next", []string{"alt+j", "ctrl+pgdown"}}, {"workspace", "back", []string{"alt+k", "ctrl+pgup"}},
	{"workspace", "attention", []string{"alt+a"}}, {"workspace", "next_attention", []string{"alt+n"}}, {"workspace", "toggle_sidebar", []string{"ctrl+s"}}, {"workspace", "rename", []string{"f2"}}, {"workspace", "toggle_focus", []string{"f6"}},
	{"management", "diagnostics", []string{"D"}}, {"management", "recovery", []string{"R"}},
	{"management", "search", []string{"/"}}, {"management", "group_mode", []string{"alt+v"}},
	{"management", "open_project", []string{"W"}}, {"management", "check_release", []string{"U"}}, {"management", "attention", []string{"t"}}, {"management", "copy_visible", []string{"V"}}, {"management", "delete", []string{"d"}},
	{"management", "move_up", []string{"shift+up", "ctrl+up", "+", "K"}}, {"management", "move_down", []string{"shift+down", "ctrl+down", "-", "J"}},
	{"management", "copy_mode", []string{"v"}}, {"management", "tab_left", []string{"["}}, {"management", "tab_right", []string{"]"}}, {"management", "quit", []string{"q", "ctrl+c"}},
	{"management", "up", []string{"up", "k"}}, {"management", "down", []string{"down", "j"}}, {"management", "group", []string{"g"}}, {"management", "first", []string{"home"}}, {"management", "last", []string{"end", "G"}},
	{"management", "move_group", []string{"M"}}, {"management", "open", []string{"enter"}}, {"management", "rename", []string{"r"}}, {"management", "refresh", []string{"ctrl+r"}},
	{"management", "attach", []string{"a"}}, {"management", "history", []string{"h"}}, {"management", "new", []string{"n"}}, {"management", "new_dialog", []string{"N"}},
	{"management", "help", []string{"?"}}, {"management", "settings", []string{"S"}},
}

type Map struct{ Overrides map[string]string }

func ID(d Definition) string { return d.Scope + "." + d.Action }
func Normalize(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	parts := strings.Split(s, "+")
	if s == "+" {
		return s, nil
	}
	key := parts[len(parts)-1]
	mods := map[string]bool{}
	for _, p := range parts[:len(parts)-1] {
		p = strings.ToLower(p)
		if (p != "ctrl" && p != "alt" && p != "shift") || mods[p] {
			return "", fmt.Errorf("invalid modifiers")
		}
		mods[p] = true
	}
	if utf8.RuneCountInString(key) == 1 {
		r, _ := utf8.DecodeRuneInString(key)
		if r < 33 || r > 126 {
			return "", fmt.Errorf("key must be ASCII")
		}
		if mods["shift"] && !mods["ctrl"] && !mods["alt"] && r >= 'a' && r <= 'z' {
			key = strings.ToUpper(key)
			delete(mods, "shift")
		}
		if mods["ctrl"] || mods["alt"] {
			key = strings.ToLower(key)
		}
	} else {
		key = strings.ToLower(key)
		valid := false
		for _, v := range []string{"space", "enter", "tab", "esc", "backspace", "up", "down", "left", "right", "home", "end", "pgup", "pgdown", "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11"} {
			valid = valid || key == v
		}
		if !valid {
			return "", fmt.Errorf("unknown key")
		}
	}
	out := ""
	for _, m := range []string{"ctrl", "alt", "shift"} {
		if mods[m] {
			out += m + "+"
		}
	}
	return out + key, nil
}
func (m Map) Keys(d Definition) []string {
	v, ok := m.Overrides[ID(d)]
	if !ok {
		return d.Keys
	}
	if v == "" {
		return nil
	}
	return []string{v}
}
func (m Map) Validate() error {
	known := map[string]bool{}
	for _, d := range Definitions {
		known[ID(d)] = true
	}
	for k, v := range m.Overrides {
		if !known[k] {
			return fmt.Errorf("unknown action %s", k)
		}
		if _, e := Normalize(v); e != nil {
			return fmt.Errorf("invalid key for %s", k)
		}
	}
	for _, scope := range []string{"workspace", "management"} {
		seen := map[string]string{}
		for _, d := range Definitions {
			if d.Scope != "global" && d.Scope != "workspace" && d.Scope != scope {
				continue
			}
			for _, k := range m.Keys(d) {
				if k == "" {
					continue
				}
				if prior := seen[k]; prior != "" {
					return fmt.Errorf("key conflict: %s / %s", prior, ID(d))
				}
				seen[k] = ID(d)
			}
		}
	}
	return nil
}
func (m Map) Translate(scope, key string) (string, bool, bool) { // canonical, mapped, suppressed
	for _, d := range Definitions {
		if d.Scope != "workspace" && d.Scope != scope {
			continue
		}
		if v, ok := m.Overrides[ID(d)]; ok && v != "" && v == key {
			return d.Keys[0], true, false
		}
	}
	for _, d := range Definitions {
		if d.Scope != "workspace" && d.Scope != scope {
			continue
		}
		if _, ok := m.Overrides[ID(d)]; ok {
			for _, k := range d.Keys {
				if key == k {
					return "", false, true
				}
			}
		}
	}
	return key, false, false
}
func (m Map) Wake() string {
	d := Definitions[0]
	x := m.Keys(d)
	if len(x) == 0 {
		return ""
	}
	return x[0]
}
