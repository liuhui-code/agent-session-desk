package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestLaunchDialogCreatesGenericTerminalTabWithLiteralArguments(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker first") })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "New terminal") })
	dir := filepath.Join(t.TempDir(), "项目 工作区")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	values := []string{"普通Terminal", `C:\工具 box\pwsh.exe`, dir, `["-NoProfile","--model=local","arg with spaces","中文"]`}
	for i, value := range values {
		if i > 0 {
			h.send(tea.KeyPressMsg{Code: tea.KeyTab})
		}
		h.send(tea.PasteMsg{Content: value})
	}
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "shell marker created-1") })
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.starts) != 1 {
		t.Fatalf("duplicate terminal launch: %d", len(service.starts))
	}
	l := service.starts[0]
	if l.Name != values[0] || l.Executable != values[1] || l.Dir != values[2] || !reflect.DeepEqual(l.Args, []string{"-NoProfile", "--model=local", "arg with spaces", "中文"}) {
		t.Fatalf("launch field boundaries changed: %#v", l)
	}
}

func TestLaunchDialogRejectsInvalidArgumentsAndCancelsWithoutStarting(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { return h.m.launch != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: "pwsh.exe"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: `null`})
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if h.m.launch == nil || h.m.launch.err == nil {
		t.Fatal("invalid arguments accepted")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.until(func() bool { return h.m.workspace.conn != nil })
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.starts) != 0 {
		t.Fatal("cancel/invalid input launched process")
	}
}

func TestLaunchDialogFitsEveryBoundedViewport(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }}}
	h := newWorkspaceHarness(t, service)
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	for _, size := range [][2]int{{1, 1}, {10, 4}, {43, 10}, {80, 24}, {120, 32}} {
		h.send(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		for i := 0; i < len(h.m.launch.fields)+1; i++ {
			view := h.m.View()
			rows := strings.Split(view.Content, "\n")
			if len(rows) > size[1] {
				t.Fatalf("dialog height %d > %d", len(rows), size[1])
			}
			for _, row := range rows {
				if ansi.StringWidth(row) > size[0] {
					t.Fatalf("dialog width overflow: %d > %d", ansi.StringWidth(row), size[0])
				}
			}
			if view.Cursor != nil && (view.Cursor.X < 0 || view.Cursor.Y < 0 || view.Cursor.X >= size[0] || view.Cursor.Y >= size[1]) {
				t.Fatalf("cursor outside viewport: %+v", view.Cursor)
			}
			h.send(tea.KeyPressMsg{Code: tea.KeyTab})
		}
	}
}

func TestLaunchDialogUncertainStartCannotCreateDuplicateRun(t *testing.T) {
	service := &workspaceFixture{startErr: fmt.Errorf("metadata commit failed"), startResult: session.Session{ID: "created-id", RunID: "created-run"}, fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }}}
	h := newWorkspaceHarness(t, service)
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: "pwsh.exe"})
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.until(func() bool { return h.m.launch != nil && h.m.launch.uncertain })
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if !strings.Contains(h.m.View().Content, "created-id") {
		t.Fatal("uncertain Session identity is hidden")
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.starts) != 1 {
		t.Fatalf("uncertain result replayed: %d", len(service.starts))
	}
}
