package keybindings

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"os"
	"strings"
)

func Load(path string) (Map, error) {
	m := Map{Overrides: map[string]string{}}
	f, e := os.Open(path)
	if os.IsNotExist(e) {
		return m, nil
	}
	if e != nil {
		return m, e
	}
	defer f.Close()
	i, e := f.Stat()
	if e != nil || !i.Mode().IsRegular() {
		return m, fmt.Errorf("keybindings must be regular file")
	}
	b, e := io.ReadAll(io.LimitReader(f, 65537))
	if e != nil || len(b) > 65536 {
		return m, fmt.Errorf("keybindings exceeds 64 KiB")
	}
	var d struct {
		Version                       int `toml:"schema_version"`
		Global, Workspace, Management map[string]string
	}
	meta, e := toml.Decode(string(b), &d)
	if e != nil || len(meta.Undecoded()) != 0 || d.Version != 1 {
		return m, fmt.Errorf("invalid keybindings v1")
	}
	for s, entries := range map[string]map[string]string{"global": d.Global, "workspace": d.Workspace, "management": d.Management} {
		for a, k := range entries {
			if k != "" {
				k, e = Normalize(k)
				if e != nil {
					return m, fmt.Errorf("invalid key: %s.%s", s, a)
				}
			}
			m.Overrides[s+"."+a] = k
		}
	}
	if k := m.Wake(); !ValidGlobal(k) {
		return m, fmt.Errorf("global wake requires ctrl or alt")
	}
	return m, m.Validate()
}
func hasModifier(k string) bool { return len(k) > 4 && (k[:5] == "ctrl+" || k[:4] == "alt+") }

func ValidGlobal(k string) bool {
	if k == "" {
		return true
	}
	if !hasModifier(k) {
		return false
	}
	parts := strings.Split(k, "+")
	last := parts[len(parts)-1]
	if last == "space" {
		return true
	}
	if len(last) == 1 && ((last[0] >= 'a' && last[0] <= 'z') || (last[0] >= '0' && last[0] <= '9')) {
		return true
	}
	for _, v := range []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11"} {
		if last == v {
			return true
		}
	}
	return false
}
