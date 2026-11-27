package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"io"
	"os"
)

type workspace struct {
	Templates []string `toml:"templates"`
}

func LoadWorkspace(path, name string) ([]session.Launch, error) {
	if name == "" {
		return nil, fmt.Errorf("workspace name required")
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 {
		return nil, fmt.Errorf("workspace config bound")
	}
	var d document
	meta, e := toml.NewDecoder(io.LimitReader(f, (1<<20)+1)).Decode(&d)
	if e != nil || len(meta.Undecoded()) != 0 || d.Version != 1 {
		return nil, fmt.Errorf("invalid workspace config")
	}
	w, ok := d.Workspaces[name]
	if !ok || len(w.Templates) == 0 || len(w.Templates) > 8 {
		return nil, fmt.Errorf("workspace requires 1–8 templates")
	}
	var out []session.Launch
	seen := map[string]bool{}
	for _, k := range w.Templates {
		if seen[k] {
			return nil, fmt.Errorf("duplicate workspace template")
		}
		seen[k] = true
		l, e := launchFrom(d, path, k)
		if e != nil {
			return nil, e
		}
		out = append(out, l)
	}
	return out, nil
}
