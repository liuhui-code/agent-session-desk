package instance

import (
	"path/filepath"
	"testing"
)

func TestScopeKeyReusesEquivalentPathsAndSeparatesDataScopes(t *testing.T) {
	root := t.TempDir()
	runtimePath := filepath.Join(root, "runtime", "asd-test")
	state := filepath.Join(root, "state", "metadata.sqlite3")
	key, err := ScopeKey("psmux", runtimePath, state)
	if err != nil {
		t.Fatal(err)
	}
	same, err := ScopeKey("psmux", filepath.Join(root, "runtime", ".", "asd-test"), state)
	if err != nil || same != key || len(key) != 64 {
		t.Fatalf("equivalent scope differs: %s %s %v", key, same, err)
	}
	for _, paths := range [][3]string{
		{"psmux", filepath.Join(root, "runtime", "asd-other"), state},
		{"psmux", runtimePath, filepath.Join(root, "other", "metadata.sqlite3")},
		{"tmux", runtimePath, state},
	} {
		other, e := ScopeKey(paths[0], paths[1], paths[2])
		if e != nil || other == key {
			t.Fatalf("different data scope reused: %v %v", paths, e)
		}
	}
}
