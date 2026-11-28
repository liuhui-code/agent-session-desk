package snippets_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/snippets"
)

func TestLoadKeepsProjectAndGlobalLiteralBodiesDistinct(t *testing.T) {
	root := t.TempDir()
	global := filepath.Join(root, "global.toml")
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(filepath.Join(project, ".asd"), 0700); err != nil {
		t.Fatal(err)
	}
	body := "schema_version = 1\n[snippets.review]\ntitle = 'Review'\nbody = '''Review 中文 ${HOME} `git status`'''\n"
	if err := os.WriteFile(global, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".asd", "snippets.toml"), []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := snippets.Load(context.Background(), global, project)
	if err != nil || len(got.Items) != 2 {
		t.Fatalf("%+v %v", got, err)
	}
	if got.Items[0].Source != "Project" || got.Items[1].Source != "Global" || got.Items[0].Body != "Review 中文 ${HOME} `git status`" {
		t.Fatalf("lost literal scope/body: %+v", got)
	}
}

func TestLoadRejectsInvalidOrUnboundedFilesWithoutPartialLibrary(t *testing.T) {
	for _, text := range []string{
		"schema_version=2\n", "schema_version=1\nUnknown=1\n", "Schema_Version=1\n",
		"schema_version=1\n[snippets.x]\nTitle='x'\nbody='ok'\n",
		"schema_version=1\n[snippets.x]\ntitle='x'\nbody=\"\\u001b[31m\"\n",
		"schema_version=1\n[snippets.x]\ntitle='x'\nbody=''\n",
	} {
		t.Run(text, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "snippets.toml")
			os.WriteFile(p, []byte(text), 0600)
			lib, err := snippets.Load(context.Background(), p, "")
			if err == nil || len(lib.Items) != 0 {
				t.Fatalf("invalid file accepted: %+v %v", lib, err)
			}
		})
	}
}

func TestLoadMissingFilesAndCancellation(t *testing.T) {
	root := t.TempDir()
	lib, err := snippets.Load(context.Background(), filepath.Join(root, "missing.toml"), root)
	if err != nil || len(lib.Items) != 0 {
		t.Fatalf("missing optional file: %+v %v", lib, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := snippets.Load(ctx, "", ""); err == nil {
		t.Fatal("cancelled load accepted")
	}
}

func TestLoadExactBodyAndItemBounds(t *testing.T) {
	for _, n := range []int{snippets.MaxBodyBytes - 1, snippets.MaxBodyBytes, snippets.MaxBodyBytes + 1} {
		p := filepath.Join(t.TempDir(), "s.toml")
		text := "schema_version=1\n[snippets.x]\ntitle='x'\nbody='" + strings.Repeat("x", n) + "'\n"
		if err := os.WriteFile(p, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := snippets.Load(context.Background(), p, "")
		if (err == nil) != (n <= snippets.MaxBodyBytes) {
			t.Fatalf("body %d: %v", n, err)
		}
	}
	for _, n := range []int{snippets.MaxItems - 1, snippets.MaxItems, snippets.MaxItems + 1} {
		p := filepath.Join(t.TempDir(), "s.toml")
		text := "schema_version=1\n"
		for i := 0; i < n; i++ {
			text += fmt.Sprintf("[snippets.s%d]\ntitle='x'\nbody='ok'\n", i)
		}
		if err := os.WriteFile(p, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := snippets.Load(context.Background(), p, "")
		if (err == nil) != (n <= snippets.MaxItems) {
			t.Fatalf("items %d: %v", n, err)
		}
	}
}

func TestLoadRejectsOversizedFileAndDirectory(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "s.toml")
	if err := os.WriteFile(p, []byte(strings.Repeat("#", snippets.MaxFileBytes+1)), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, p} {
		if _, err := snippets.Load(context.Background(), path, ""); err == nil {
			t.Fatalf("invalid source %q", path)
		}
	}
}
