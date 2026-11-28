package keybindings

import "testing"
import "os"

func TestOverrides(t *testing.T) {
	m := Map{map[string]string{"workspace.next": "alt+x"}}
	if e := m.Validate(); e != nil {
		t.Fatal(e)
	}
	k, ok, _ := m.Translate("workspace", "alt+x")
	if k != "alt+j" || !ok {
		t.Fatal(k)
	}
	_, _, off := m.Translate("workspace", "alt+j")
	if !off {
		t.Fatal("old active")
	}
	m.Overrides["workspace.back"] = "alt+x"
	if m.Validate() == nil {
		t.Fatal("collision")
	}
}
func TestDisable(t *testing.T) {
	m := Map{map[string]string{"global.wake": ""}}
	if m.Wake() != "" {
		t.Fatal("disable")
	}
}

func TestLoadRejectsUnknownAndConflicts(t *testing.T) {
	for _, s := range []string{"schema_version=2", "schema_version=1\n[workspace]\nwat='alt+x'", "schema_version=1\n[workspace]\nnext='alt+k'", "schema_version=1\n[global]\nwake='x'"} {
		p := t.TempDir() + "/keys"
		if e := os.WriteFile(p, []byte(s), 0600); e != nil {
			t.Fatal(e)
		}
		if _, e := Load(p); e == nil {
			t.Fatal("invalid accepted", s)
		}
	}
}
func TestLoadValid(t *testing.T) {
	p := t.TempDir() + "/k"
	os.WriteFile(p, []byte("schema_version=1\n[global]\nwake='ctrl+alt+f9'\n[workspace]\nnext='alt+x'\nprevious='alt+z'\n[management]\nopen_project='f8'"), 0600)
	if _, e := Load(p); e != nil {
		t.Fatal(e)
	}
}

func TestSnippetShortcutCanBeReplacedAndDisabled(t *testing.T) {
	m := Map{Overrides: map[string]string{"workspace.snippets": "alt+i"}}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	if k, mapped, off := m.Translate("workspace", "alt+i"); k != "f7" || !mapped || off {
		t.Fatal("custom snippet shortcut not routed")
	}
	if _, _, off := m.Translate("workspace", "f7"); !off {
		t.Fatal("old shortcut not returned to original terminal")
	}
	m.Overrides["workspace.snippets"] = ""
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, _, off := m.Translate("workspace", "f7"); !off {
		t.Fatal("snippet shortcut not disabled")
	}
}
