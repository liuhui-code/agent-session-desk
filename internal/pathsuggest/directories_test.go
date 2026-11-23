package pathsuggest

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDirectoryCompletionAndFuzzyCorpus(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"Alpha", "alpine", ".hidden", "beta"} {
		if err := os.Mkdir(filepath.Join(root, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	got := Complete(context.Background(), root+string(os.PathSeparator)+"al")
	want := []string{filepath.Join(root, "Alpha"), filepath.Join(root, "alpine")}
	if !reflect.DeepEqual(got.Paths, want) || got.Exact {
		t.Fatalf("prefix completion: %#v", got)
	}
	exact := Complete(context.Background(), want[0])
	if !exact.Exact || !reflect.DeepEqual(exact.Paths, want[:1]) {
		t.Fatalf("exact: %#v", exact)
	}
	filtered := Filter([]string{"/work/Other", "/work/MyProject", "/work/MyPrototype"}, "MYPR", nil)
	if len(filtered) != 2 || filtered[0] != "/work/MyProject" {
		t.Fatalf("case-insensitive fuzzy ordering: %v", filtered)
	}
}
