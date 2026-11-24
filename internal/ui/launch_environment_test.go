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
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestLaunchDialogPassesLiteralEnvironmentOverridesWithoutChangingManager(t *testing.T) {
	t.Setenv("ASD_UI_ENV_TEST", "outer-sentinel")
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return sampleSessions(), nil }}}
	h := newWorkspaceHarness(t, service)
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.until(func() bool { return h.m.launch != nil })
	dir := filepath.Join(t.TempDir(), "项目 空间")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	fields := []string{"普通 CLI", `C:\工具 box\agent.exe`, dir, `["", "literal $ASD_UI_ENV_TEST"]`, `["ASD_UI_ENV_TEST=child 中文", "ASD_EMPTY=", "ASD_EQUAL=a=b", "ASD_LITERAL=$HOME", "ASD_UI_ENV_TEST=last value"]`}
	for i, value := range fields {
		if i > 0 {
			h.send(tea.KeyPressMsg{Code: tea.KeyTab})
		}
		h.send(tea.PasteMsg{Content: value})
	}
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.until(func() bool {
		return h.m.launch == nil && strings.Contains(h.m.View().Content, "shell marker created-1")
	})
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.starts) != 1 {
		t.Fatalf("launch count %d", len(service.starts))
	}
	got := service.starts[0]
	want := []string{"ASD_UI_ENV_TEST=child 中文", "ASD_EMPTY=", "ASD_EQUAL=a=b", "ASD_LITERAL=$HOME", "ASD_UI_ENV_TEST=last value"}
	if !reflect.DeepEqual(got.Env, want) || !reflect.DeepEqual(got.Args, []string{"", "literal $ASD_UI_ENV_TEST"}) || got.Dir != fields[2] {
		t.Fatalf("launch fields changed: %+v", got)
	}
	if strings.Contains(h.m.View().Content, "last value") || strings.Contains(h.m.View().Content, "child 中文") {
		t.Fatal("default view expanded environment values")
	}
	if os.Getenv("ASD_UI_ENV_TEST") != "outer-sentinel" {
		t.Fatal("manager environment changed")
	}
}

func TestLaunchEnvironmentFailureDoesNotExposeValuesOrRetryUncertainRun(t *testing.T) {
	const value = "synthetic-env-value-do-not-echo"
	service := &workspaceFixture{startErr: fmt.Errorf("backend stderr echoed %s", value), startResult: session.Session{ID: "env-created-session", RunID: "env-created-run"}, fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }}}
	h := newWorkspaceHarness(t, service)
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	h.send(tea.PasteMsg{Content: "synthetic-cli"})
	for range 3 {
		h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	h.send(tea.PasteMsg{Content: `["ASD_UI_VALUE=` + value + `"]`})
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.until(func() bool { return h.m.launch != nil && h.m.launch.uncertain })
	if got := h.m.launch.err.Error(); strings.Contains(got, value) || !strings.Contains(got, "env-created-session") || !strings.Contains(got, "不会自动重试") {
		t.Fatalf("unsafe or unidentified failure summary: %q", got)
	}
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.starts) != 1 {
		t.Fatal("uncertain environment launch was retried")
	}
}

func TestLaunchEnvironmentValidationAndCancelNeverStartOrRememberDraft(t *testing.T) {
	for _, raw := range []string{`null`, `{"ASD_KEY":"synthetic-private"}`, `[1]`, `[true]`, `[null]`, `["=synthetic-private"]`, `["synthetic-private"]`, `["ASD\nBAD=synthetic-private"]`, `["ASD_KEY=synthetic-private\u0000"]`, `["synthetic-private"`} {
		t.Run(fmt.Sprint(len(raw), "/", raw[:min(8, len(raw))]), func(t *testing.T) {
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }}}
			h := newWorkspaceHarness(t, service)
			h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
			h.send(tea.KeyPressMsg{Code: tea.KeyTab})
			h.send(tea.PasteMsg{Content: "synthetic-cli"})
			for range 3 {
				h.send(tea.KeyPressMsg{Code: tea.KeyTab})
			}
			h.send(tea.PasteMsg{Content: raw})
			h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			if h.m.launch == nil || h.m.launch.err == nil || strings.Contains(h.m.launch.err.Error(), "synthetic-private") {
				t.Fatal("invalid environment accepted or echoed")
			}
			service.mu.Lock()
			n := len(service.starts)
			service.mu.Unlock()
			if n != 0 {
				t.Fatal("validation started a process")
			}
			h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
			h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
			if h.m.launch == nil || h.m.launch.fields[4].Value() != "" {
				t.Fatal("cancelled environment draft was remembered")
			}
		})
	}
}
func TestBlankLaunchEnvironmentKeepsInheritedEnvironmentAndCanReachCreateButton(t *testing.T) {
	for _, raw := range []string{"", "  ", "[]"} {
		t.Run(fmt.Sprintf("%q", raw), func(t *testing.T) {
			service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }}}
			h := newWorkspaceHarness(t, service)
			h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
			h.send(tea.KeyPressMsg{Code: tea.KeyTab})
			h.send(tea.PasteMsg{Content: "synthetic-cli"})
			for range 3 {
				h.send(tea.KeyPressMsg{Code: tea.KeyTab})
			}
			h.send(tea.PasteMsg{Content: raw})
			h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
			if h.m.launch.focus != len(h.m.launch.fields) {
				t.Fatal("environment Enter did not reach creation button")
			}
			h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
			h.until(func() bool { return h.m.launch == nil })
			service.mu.Lock()
			defer service.mu.Unlock()
			if len(service.starts) != 1 || len(service.starts[0].Env) != 0 {
				t.Fatal("blank environment cleared or fabricated overrides")
			}
		})
	}
}

func TestEnvironmentFailureKeepsSafeFieldLocationWithoutBackendDetails(t *testing.T) {
	for _, c := range []struct{ source, field string }{{"executable not found or not executable: synthetic-private", "程序"}, {"working directory: synthetic-private is unavailable", "工作目录"}, {"environment key ASD_PRIVATE has invalid synthetic-private value", "环境覆盖"}} {
		t.Run(c.field, func(t *testing.T) {
			service := &workspaceFixture{startErr: fmt.Errorf("%s", c.source), fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }}}
			h := newWorkspaceHarness(t, service)
			h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
			h.send(tea.KeyPressMsg{Code: tea.KeyTab})
			h.send(tea.PasteMsg{Content: "synthetic-cli"})
			for range 3 {
				h.send(tea.KeyPressMsg{Code: tea.KeyTab})
			}
			h.send(tea.PasteMsg{Content: `["ASD_PRIVATE=synthetic-private"]`})
			h.send(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			h.until(func() bool { return h.m.launch != nil && h.m.launch.err != nil })
			summary := h.m.launch.err.Error()
			if !strings.HasPrefix(summary, c.field) || strings.Contains(summary, "synthetic-private") {
				t.Fatalf("field location lost or private value echoed: %s", summary)
			}
		})
	}
}

func TestEnvironmentPasteBoundAndMultilineRejectionKeepExistingDraft(t *testing.T) {
	service := &workspaceFixture{fakeService: fakeService{list: func(context.Context) ([]session.Session, error) { return nil, nil }}}
	h := newWorkspaceHarness(t, service)
	h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
	for range 4 {
		h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	const draft = `["ASD_MARKER=retained"]`
	h.send(tea.PasteMsg{Content: draft})
	for _, bad := range []string{strings.Repeat("x", launchFieldLimit+1), "ASD_MARKER=first\nsecond"} {
		h.send(tea.PasteMsg{Content: bad})
		if h.m.launch.fields[4].Value() != draft || h.m.launch.err == nil {
			t.Fatal("rejected environment paste changed accepted draft")
		}
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(service.starts) != 0 {
		t.Fatal("cancelled paste started a process")
	}
}
