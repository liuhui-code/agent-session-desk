package session

import (
	"strings"
	"unicode"
)

// Keep a bounded basename only. Shell integration may return an entire command;
// arguments/control text are not a display identity and never reach the UI.
func ProgramLabel(v string) string {
	if strings.ContainsAny(v, "\x00\r\n") {
		return ""
	}
	v = strings.TrimSpace(v)
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		v = v[1 : len(v)-1]
	}
	v = v[strings.LastIndexAny(v, `/\`)+1:]
	if len(v) == 0 || len(v) > 96 {
		return ""
	}
	for _, r := range v {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '.' && r != '_' && r != '-' {
			return ""
		}
	}
	if strings.HasSuffix(strings.ToLower(v), ".exe") {
		v = v[:len(v)-4]
	}
	switch strings.ToLower(v) {
	case "", ".", "..", "shell", "pwsh", "powershell", "cmd", "sh", "bash", "zsh", "fish", "nu", "conhost", "openconsole", "windowsterminal", "wt", "node", "nodejs":
		return ""
	}
	return v
}
