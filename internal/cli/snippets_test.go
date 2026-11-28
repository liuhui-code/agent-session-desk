package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

type snippetDirectoryFixture struct {
	dir    string
	err    error
	target session.Session
}

func (s *snippetDirectoryFixture) CurrentDirectory(_ context.Context, target session.Session) (string, error) {
	s.target = target
	return s.dir, s.err
}
func TestSnippetOptionsUseOriginalPaneAndMetadataDirectory(t *testing.T) {
	root := t.TempDir()
	project := t.TempDir()
	if err := os.Mkdir(filepath.Join(project, ".asd"), 0700); err != nil {
		t.Fatal(err)
	}
	data := []byte("schema_version=1\n[snippets.review]\ntitle='Review'\nbody='literal'\n")
	for _, p := range []string{filepath.Join(root, "snippets.toml"), filepath.Join(project, ".asd", "snippets.toml")} {
		if err := os.WriteFile(p, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	source := &snippetDirectoryFixture{dir: project}
	opts := withSnippets(ui.Options{}, source, filepath.Join(root, "metadata.sqlite3"))
	original := session.Session{ID: "original", RunID: "run", BackendID: "pane"}
	lib, err := opts.LoadSnippets(context.Background(), original)
	if err != nil || len(lib.Items) != 2 || source.target != original || lib.ProjectDir != project {
		t.Fatalf("%+v %v", lib, err)
	}
	source.err = errors.New("synthetic unavailable")
	source.dir = ""
	lib, err = opts.LoadSnippets(context.Background(), original)
	if err != nil || len(lib.Items) != 1 || lib.Items[0].Source != "Global" || lib.ProjectDir != "" {
		t.Fatalf("global fallback: %+v %v", lib, err)
	}
}
