//go:build windows

package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func TestWindowsPsmuxNativeTerminalScreenInputResizeAndReconnect(t *testing.T) {
	h := newNativeHarness(t)
	launch, report := h.launch("Tab native terminal")
	s := h.start(launch)
	fixtureEvent(t, h.ctx, report, "started", "")
	live, err := h.b.OpenTerminal(h.ctx, s, 80, 20)
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	waitTerminalScreen(t, h.ctx, live, 80, 20)
	const text = "ASD_TAB_NATIVE 中文 🧪 e\u0301 ; \\ end"
	if err := live.Paste(h.ctx, text); err != nil {
		t.Fatal(err)
	}
	// Paste does not imply Enter. Submit is a separate explicit key event.
	events, err := readNativeEvents(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Event == "input" {
			t.Fatalf("paste submitted before Enter: %+v", event)
		}
	}
	if err := live.SendKey(h.ctx, "Enter"); err != nil {
		t.Fatal(err)
	}
	fixtureEvent(t, h.ctx, report, "input", text)
	if err := live.Resize(h.ctx, 92, 24); err != nil {
		t.Fatal(err)
	}
	waitTerminalScreen(t, h.ctx, live, 92, 24)
	if err := live.Close(); err != nil {
		t.Fatal(err)
	}
	rows, err := h.fresh().List(h.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := findNativeSession(rows, s); err != nil {
		t.Fatal(err)
	}
	reopened, err := h.fresh().OpenTerminal(h.ctx, s, 92, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err := reopened.SendText(h.ctx, "ASD_REOPENED"); err != nil {
		t.Fatal(err)
	}
	if err := reopened.SendKey(h.ctx, "Enter"); err != nil {
		t.Fatal(err)
	}
	fixtureEvent(t, h.ctx, report, "input", "ASD_REOPENED")
}

func waitTerminalScreen(t *testing.T, parent context.Context, live *terminal.PSMux, width, height int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	for {
		screen, err := live.Next(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if screen.Width == width && screen.Height == height {
			return
		}
		if ctx.Err() != nil {
			t.Fatal(fmt.Sprintf("screen never reached %dx%d", width, height))
		}
	}
}

func TestWindowsPsmuxNativeTerminalOrdinaryShellKeepsPromptAndDraft(t *testing.T) {
	h := newNativeHarness(t)
	shell := os.Getenv("ASD_SHELL_TEST_BINARY")
	if shell == "" {
		t.Skip("ordinary shell native test requires ASD_SHELL_TEST_BINARY")
	}
	if !filepath.IsAbs(shell) {
		t.Fatal("shell path must be absolute")
	}
	s := h.start(session.Launch{Name: "Ordinary shell Tab", Executable: shell, Args: []string{"-NoLogo", "-NoProfile"}, Dir: h.dir})
	live, err := h.b.OpenTerminal(h.ctx, s, 80, 20)
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	waitTerminalScreen(t, h.ctx, live, 80, 20)
	if err := live.SendText(h.ctx, "'first' | Set-Content -LiteralPath 'observed-1.txt'"); err != nil {
		t.Fatal(err)
	}
	if err := live.SendKey(h.ctx, "Enter"); err != nil {
		t.Fatal(err)
	}
	assertNativeShellFile(t, h, "observed-1.txt", "first")
	// A second prompt and a new command remain part of the same root shell Run.
	if err := live.Paste(h.ctx, "'第二 🧪' | Set-Content -LiteralPath 'observed-2.txt'"); err != nil {
		t.Fatal(err)
	}
	if err := live.SendKey(h.ctx, "Enter"); err != nil {
		t.Fatal(err)
	}
	assertNativeShellFile(t, h, "observed-2.txt", "第二 🧪")
	const draft = "'draft-preserved' | Set-Content -LiteralPath 'observed-3.txt'"
	if err := live.SendText(h.ctx, draft); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(h.ctx, 10*time.Second)
	defer cancel()
	for {
		screen, err := live.Next(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var text strings.Builder
		for _, row := range screen.Rows {
			for _, run := range row.Runs {
				text.WriteString(run.Text)
			}
		}
		if strings.Contains(text.String(), "observed-3.txt") {
			break
		}
	}
	if _, err := os.Stat(filepath.Join(h.dir, "observed-3.txt")); !os.IsNotExist(err) {
		t.Fatal("draft submitted before Enter")
	}
	_ = live.Close()
	rows, err := h.fresh().List(h.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := findNativeSession(rows, s); err != nil {
		t.Fatal(err)
	}
	reopened, err := h.fresh().OpenTerminal(h.ctx, s, 80, 20)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err := reopened.SendKey(h.ctx, "Enter"); err != nil {
		t.Fatal(err)
	}
	assertNativeShellFile(t, h, "observed-3.txt", "draft-preserved")
}

func assertNativeShellFile(t *testing.T, h *nativeHarness, name, want string) {
	t.Helper()
	waitNative(t, h.ctx, "ordinary shell execution "+name, func(context.Context) error {
		text, err := os.ReadFile(filepath.Join(h.dir, name))
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(text)) != want {
			return fmt.Errorf("unexpected synthetic shell output %q", text)
		}
		return nil
	})
}
