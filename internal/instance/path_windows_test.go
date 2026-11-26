//go:build windows

package instance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsInstanceScopeSurvivesCreationCaseAndExtendedPathSpelling(t *testing.T) {
	root := t.TempDir()
	runtimePath := filepath.Join(root, "Runtime", "asd-test")
	metadataFile := filepath.Join(root, "State", "metadata.sqlite3")
	before, err := ScopeKey("psmux", runtimePath, metadataFile)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(runtimePath, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Dir(metadataFile), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(metadataFile, []byte("scope identity fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	after, err := ScopeKey("psmux", strings.ToUpper(runtimePath), `\\?\`+metadataFile)
	if err != nil || before != after {
		t.Fatalf("same data scope changed after first startup/case/extended spelling: %v", err)
	}
}
