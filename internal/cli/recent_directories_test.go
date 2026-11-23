package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
)

func TestRecentDirectoryCallbacksUseBackendScopeAndKeepLaunchAcknowledgement(t *testing.T) {
	root := t.TempDir()
	a := backend.Options{RuntimeDir: t.TempDir(), Namespace: "asd-recents-a"}
	b := a
	b.Namespace = "asd-recents-b"
	optionsA, optionsB := loadUIOptions(a, root), loadUIOptions(b, root)
	dir := filepath.Join(t.TempDir(), "synthetic directory")
	var warnings bytes.Buffer
	rememberLaunchDirectory(context.Background(), a, root, dir, &warnings)
	if warnings.Len() != 0 {
		t.Fatal(warnings.String())
	}
	got, err := optionsA.LoadRecentDirs(context.Background())
	if err != nil || !reflect.DeepEqual(got, []string{dir}) {
		t.Fatal(got, err)
	}
	got, err = optionsB.LoadRecentDirs(context.Background())
	if err != nil || len(got) != 0 {
		t.Fatal("scope leaked", got, err)
	}
	path, err := recentDirectoriesPath(a, root)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte("schema_version=999\n"), 0600); err != nil {
		t.Fatal(err)
	}
	rememberLaunchDirectory(context.Background(), a, root, dir, &warnings)
	if !bytes.Contains(warnings.Bytes(), []byte("终端已启动")) {
		t.Fatal("hint failure obscured process acknowledgement")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "schema_version=999\n" {
		t.Fatal("future cache overwritten")
	}
}
