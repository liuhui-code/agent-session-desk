package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestTerminalHostPreservesGlobalArgumentsAndEscapesWTSeparators(t *testing.T) {
	args := []string{"--state-dir", `C:\中文 path;a`, "--namespace", "asd-test"}
	token := "11111111111111111111111111111111"
	got := terminalHostArgs(`C:\a path\asd.exe`, `C:\cwd`, args, token)
	want := []string{"--window", "new", "new-tab", "--title", "ASD Workbench " + token, "--useApplicationTitle", "--startingDirectory", `C:\cwd`, "--", `C:\a path\asd.exe`, "--state-dir", `C:\中文 path\;a`, "--namespace", "asd-test", "--inside-terminal", "--window-token", token}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("host argv changed: %#v", got)
	}
}

func TestPortableOptionsUseManifestNamespaceAndKeepExplicitFlags(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "terminal-runtime"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "terminal-runtime", "psmux.exe"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "alpha.json"), append([]byte{239, 187, 191}, []byte(`{"namespace":"asd-personal"}`)...), 0600); err != nil {
		t.Fatal(err)
	}
	opts, dir, err := portableOptions(filepath.Join(root, "asd.exe"), backend.Options{}, "")
	if err != nil || opts.Namespace != "asd-personal" || dir != filepath.Join(root, "data", "state") {
		t.Fatalf("portable defaults: %#v %q %v", opts, dir, err)
	}
	explicit := backend.Options{Namespace: "asd-explicit", RuntimeDir: "custom"}
	opts, dir, err = portableOptions(filepath.Join(root, "asd.exe"), explicit, "state")
	if err != nil || opts.Namespace != explicit.Namespace || opts.RuntimeDir != "custom" || dir != "state" {
		t.Fatal("explicit paths changed")
	}
}

func TestPinnedLocalTerminalRefusesChangedHost(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "WindowsTerminal.exe")
	body := []byte("pinned terminal fixture")
	if err := os.WriteFile(file, body, 0600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(body)
	raw, _ := json.Marshal(map[string]any{"terminal": map[string]string{"path": file, "sha256": hex.EncodeToString(digest[:])}, "profileGuid": "{fixture-profile}"})
	if err := os.WriteFile(filepath.Join(root, "local-host.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	host, err := localTerminalHost(root)
	if err != nil || host.Path != file || host.Profile != "{fixture-profile}" {
		t.Fatalf("pinned host: %#v %v", host, err)
	}
	if err := os.WriteFile(file, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := localTerminalHost(root); err == nil {
		t.Fatal("changed host accepted")
	}
}
