//go:build windows

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
	"golang.org/x/sys/windows"
)

func TestWindowsTerminalBundleInputHistoryAndReconnect(t *testing.T) {
	dir, shell := os.Getenv("ASD_TERMINAL_BUNDLE_TEST_DIR"), os.Getenv("ASD_SHELL_TEST_BINARY")
	if dir == "" || shell == "" {
		t.Skip("requires a pinned bundle and explicit PowerShell 7 executable")
	}
	h := newNativeHarness(t)
	t.Cleanup(func() {
		if destination := os.Getenv("ASD_NATIVE_EVIDENCE_DIR"); destination != "" {
			destination = filepath.Join(destination, t.Name())
			if err := os.MkdirAll(destination, 0700); err != nil {
				t.Error(err)
				return
			}
			for _, name := range []string{"bundle-flow.ps1", "root.json", "normal.txt", "alt.txt", "size.txt", "reopen.txt", "history-captures.jsonl"} {
				data, err := os.ReadFile(filepath.Join(h.dir, name))
				if os.IsNotExist(err) {
					continue
				}
				if err == nil {
					err = os.WriteFile(filepath.Join(destination, name), data, 0600)
				}
				if err != nil {
					t.Error(err)
				}
			}
		}
	})
	h.opts.Binary, h.opts.TerminalBundle = "", dir
	var err error
	h.b, err = New(h.opts)
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(h.dir, "bundle-flow.ps1")
	if err := os.WriteFile(script, []byte(bundleFlowScript), 0600); err != nil {
		t.Fatal(err)
	}
	s := h.start(session.Launch{Name: "Bundle input and history", Executable: shell, Args: []string{"-NoLogo", "-NoProfile", "-NoExit", "-File", script}, Dir: h.dir})
	assertNativeShellFile(t, h, "seeded.txt", "seeded")
	assertBundleHistory(t, h, s)
	live, err := h.b.OpenTerminal(h.ctx, s, 94, 25)
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	bundleWaitText(t, h.ctx, live, "BUNDLE_INPUT_READY")
	bundleSubmit(t, h, live, "ALT输入α")
	bundleExactFile(t, h, "normal.txt", "ALT输入α")
	assertBundleHistory(t, h, s)
	if err := os.WriteFile(filepath.Join(h.dir, "alt-gate.txt"), []byte("go"), 0600); err != nil {
		t.Fatal(err)
	}
	bundleWaitText(t, h.ctx, live, "BUNDLE_ALT_READY")
	if err := live.Resize(h.ctx, 61, 18); err != nil {
		t.Fatal(err)
	}
	waitTerminalScreen(t, h.ctx, live, 61, 18)
	bundleSubmit(t, h, live, "ALT输入α")
	bundleExactFile(t, h, "alt.txt", "ALT输入α")
	bundleWaitText(t, h.ctx, live, "BUNDLE_ALT_DONE")
	bundleExactFile(t, h, "size.txt", "61,18")
	assertBundleHistory(t, h, s)
	if err := live.Close(); err != nil {
		t.Fatal(err)
	}
	fresh := h.fresh()
	rows, err := fresh.List(h.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := findNativeSession(rows, s); err != nil {
		t.Fatal(err)
	}
	reopened, err := fresh.OpenTerminal(h.ctx, s, 94, 25)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	// The fixture has returned to the ordinary PowerShell prompt in the same Run.
	bundleSubmit(t, h, reopened, "'reopened' | Set-Content -LiteralPath 'reopen.txt'")
	assertNativeShellFile(t, h, "reopen.txt", "reopened")
	assertBundleHistory(t, h, s)
	identity, err := os.ReadFile(filepath.Join(h.dir, "root.json"))
	if err != nil {
		t.Fatal(err)
	}
	var root struct {
		PID     uint32
		Created uint64
		Path    string
	}
	if err := json.Unmarshal(identity, &root); err != nil {
		t.Fatal(err)
	}
	handle, err := bundleProcessHandle(root.PID, root.Created, root.Path)
	if err != nil {
		t.Fatal("original root process did not survive reconnect:", err)
	}
	windows.CloseHandle(handle)
	t.Logf("bundle flow retained original Session=%s Run=%s rootPID=%d creation=%d, exact Unicode, 40 history lines, alternate resize and prompt reconnect", s.ID, s.RunID, root.PID, root.Created)
}

func bundleSubmit(t *testing.T, h *nativeHarness, live terminal.Connection, text string) {
	t.Helper()
	if err := live.SendText(h.ctx, text); err != nil {
		t.Fatal(err)
	}
	if err := live.SendKey(h.ctx, "Enter"); err != nil {
		t.Fatal(err)
	}
}

func bundleExactFile(t *testing.T, h *nativeHarness, name, want string) {
	t.Helper()
	waitNative(t, h.ctx, name, func(context.Context) error {
		b, err := os.ReadFile(filepath.Join(h.dir, name))
		if err != nil {
			return err
		}
		if string(b) != want {
			return fmt.Errorf("independent input differs: got %q want %q", b, want)
		}
		return nil
	})
}

func bundleWaitText(t *testing.T, parent context.Context, live terminal.Connection, text string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	for {
		s, err := live.Next(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var body strings.Builder
		for _, row := range s.Rows {
			for _, r := range row.Runs {
				body.WriteString(r.Text)
			}
			body.WriteByte('\n')
		}
		if strings.Contains(body.String(), text) {
			return
		}
	}
}

func assertBundleHistory(t *testing.T, h *nativeHarness, s session.Session) {
	t.Helper()
	p, err := h.b.Preview(h.ctx, s, 2000)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(filepath.Join(h.dir, "history-captures.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = json.NewEncoder(file).Encode(p)
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("capture evidence: %v %v", err, closeErr)
	}
	var got, want []int
	for _, m := range regexp.MustCompile(`(?m)^history line ([0-9]+)\r?$`).FindAllStringSubmatch(p.Text, -1) {
		n, _ := strconv.Atoi(m[1])
		got = append(got, n)
	}
	for n := 1; n <= 40; n++ {
		want = append(want, n)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("history missing/duplicated: %v", got)
	}
}

const bundleFlowScript = `
$ErrorActionPreference='Stop'
$p=Get-Process -Id $PID
@{PID=$PID;Created=$p.StartTime.ToUniversalTime().ToFileTimeUtc();Path=$p.Path}|ConvertTo-Json|Set-Content root.json -Encoding utf8NoBOM
Clear-Host
1..40|ForEach-Object{'history line '+$_}
[IO.File]::WriteAllText((Join-Path $PWD 'seeded.txt'),'seeded')
'BUNDLE_INPUT_READY'
$line=[Console]::ReadLine()
[IO.File]::WriteAllText((Join-Path $PWD 'normal.txt'),$line,[Text.UTF8Encoding]::new($false))
while(!(Test-Path alt-gate.txt)){Start-Sleep -Milliseconds 20}
[Console]::Write("` + "`e" + `[?1049h` + "`e" + `[2J` + "`e" + `[HBUNDLE_ALT_READY")
try{
 $line=[Console]::ReadLine()
 [IO.File]::WriteAllText((Join-Path $PWD 'alt.txt'),$line,[Text.UTF8Encoding]::new($false))
 [IO.File]::WriteAllText((Join-Path $PWD 'size.txt'),([string][Console]::WindowWidth+','+[string][Console]::WindowHeight))
}finally{[Console]::Write("` + "`e" + `[?1049l")}
'BUNDLE_ALT_DONE'
`
