//go:build windows

package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWindowsPortableCLIUsesPackagedStateWithoutFlags(t *testing.T) {
	product, bundle := os.Getenv("ASD_PORTABLE_PRODUCT"), os.Getenv("ASD_UPDATE_BUNDLE")
	if product == "" || bundle == "" {
		t.Skip("explicit isolated native product and bundle required")
	}
	root := t.TempDir()
	updateCopy(t, product, filepath.Join(root, "asd.exe"))
	os.Mkdir(filepath.Join(root, "terminal-runtime"), 0700)
	for _, n := range []string{"psmux.exe", "conpty.dll", "OpenConsole.exe"} {
		updateCopy(t, filepath.Join(bundle, n), filepath.Join(root, "terminal-runtime", n))
	}
	os.WriteFile(filepath.Join(root, "alpha.json"), []byte(`{"namespace":"asd-portable-scope"}`), 0600)
	for _, name := range []string{"APPDATA", "LOCALAPPDATA", "XDG_STATE_HOME", "XDG_CONFIG_HOME", "XDG_DATA_HOME"} {
		t.Setenv(name, filepath.Join(root, "isolated-home"))
	}
	call := func(args ...string) ([]byte, error) {
		cmd := exec.Command(filepath.Join(root, "asd.exe"), args...)
		cmd.Dir = root
		return cmd.CombinedOutput()
	}
	b, err := call("start", "--name", "Portable scope fixture", "--cwd", root, "--",
		filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe"),
		"-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 90")
	if err != nil {
		t.Fatalf("start: %v %s", err, b)
	}
	var started struct {
		ID string `json:"session_id"`
	}
	if json.Unmarshal(b, &started) != nil || len(started.ID) != 32 {
		t.Fatalf("start receipt: %s", b)
	}
	t.Cleanup(func() {
		if b, err := call("stop", started.ID); err != nil {
			t.Errorf("owned Run cleanup: %v %s", err, b)
		}
	})
	if _, err = os.Stat(filepath.Join(root, "data", "state", "metadata.sqlite3")); err != nil {
		t.Fatal("CLI did not share packaged workbench state", err)
	}
	b, err = call("list", "--json")
	if err != nil {
		t.Fatalf("list: %v %s", err, b)
	}
	var rows []struct {
		ID string `json:"session_id"`
	}
	if json.Unmarshal(b, &rows) != nil || len(rows) != 1 || rows[0].ID != started.ID {
		t.Fatal("portable CLI lost original Run identity")
	}
}
