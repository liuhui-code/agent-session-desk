package config

import (
	"os"
	"testing"
)

func TestWorkspace(t *testing.T) {
	p := t.TempDir() + "/x"
	os.WriteFile(p, []byte("schema_version=1\n[templates.a]\nexecutable='cmd.exe'\n[workspaces.default]\ntemplates=['a']"), 0600)
	x, e := LoadWorkspace(p, "default")
	if e != nil || len(x) != 1 {
		t.Fatal(e)
	}
	if _, e = LoadWorkspace(p, "bad"); e == nil {
		t.Fatal("missing")
	}
}
