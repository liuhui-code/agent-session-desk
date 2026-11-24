package ui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/liuhui-code/agent-session-desk/internal/appearance"
)

type controlledAppearance struct {
	values       chan appearance.Observation
	closeGate    chan struct{}
	closeStarted chan struct{}
	closeErr     error
}

func newControlledAppearance() *controlledAppearance {
	return &controlledAppearance{values: make(chan appearance.Observation, 4), closeGate: make(chan struct{}), closeStarted: make(chan struct{})}
}
func (c *controlledAppearance) Next(ctx context.Context) (appearance.Observation, error) {
	select {
	case v := <-c.values:
		return v, nil
	case <-ctx.Done():
		return appearance.Observation{}, ctx.Err()
	}
}
func (c *controlledAppearance) Close() error { close(c.closeStarted); <-c.closeGate; return c.closeErr }
func themeEvent(t *testing.T, m *Model, msg tea.Msg) tea.Cmd {
	t.Helper()
	c, handled := m.systemThemeEvent(msg)
	if !handled {
		t.Fatal("theme receipt was not handled")
	}
	return c
}
func themeModel(t *testing.T, open func(context.Context) (appearance.Connection, error)) *Model {
	t.Helper()
	m := NewWithOptions(context.Background(), nil, Options{Theme: "system", OpenSystemTheme: open})
	m.Update(tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	t.Cleanup(m.Close)
	return m
}
func TestSystemThemeWaitsForCloseAndRejectsLateAndDuplicateReceipts(t *testing.T) {
	a, b := newControlledAppearance(), newControlledAppearance()
	opens := 0
	m := themeModel(t, func(context.Context) (appearance.Connection, error) {
		opens++
		if opens == 1 {
			return a, nil
		}
		return b, nil
	})
	opened := m.syncSystemTheme()()
	read := themeEvent(t, m, opened)
	if themeEvent(t, m, opened) != nil {
		t.Fatal("duplicate open started second reader")
	}
	a.values <- appearance.Observation{Theme: "light", Source: "windows"}
	light := read()
	read = themeEvent(t, m, light)
	if themeEvent(t, m, light) != nil {
		t.Fatal("duplicate read started second reader")
	}
	if colorHex(m.View().BackgroundColor) != "d5d6db" {
		t.Fatal("observation not applied")
	}
	m.applyUITheme("dark")
	closing := m.syncSystemTheme()
	if colorHex(m.View().BackgroundColor) != "000000" {
		t.Fatal("explicit dark not applied")
	}
	// Even a queued old Light result must not win over explicit Dark.
	late := systemThemeRead{m.systemTheme.op, m.systemTheme.op.sequence, appearance.Observation{Theme: "light", Source: "windows"}, nil}
	if themeEvent(t, m, late) != nil || colorHex(m.View().BackgroundColor) != "000000" {
		t.Fatal("late light overrode dark or repeated close")
	}
	_ = read // Its canceled context is also covered by the opening/Next contract tests.
	m.applyUITheme("system")
	if m.syncSystemTheme() != nil || opens != 1 {
		t.Fatal("reopened before close")
	}
	closed := make(chan tea.Msg, 1)
	go func() { closed <- closing() }()
	select {
	case <-a.closeStarted:
	case <-time.After(time.Second):
		t.Fatal("close did not start")
	}
	if m.syncSystemTheme() != nil || opens != 1 {
		t.Fatal("overlapped pending close")
	}
	close(a.closeGate)
	var closedMsg tea.Msg
	select {
	case closedMsg = <-closed:
	case <-time.After(time.Second):
		t.Fatal("close did not finish")
	}
	reopen := themeEvent(t, m, closedMsg)
	if reopen == nil {
		t.Fatal("System choice was lost while closing")
	}
	openedB := reopen()
	if opens != 2 {
		t.Fatal("did not open exactly one replacement")
	}
	_ = themeEvent(t, m, openedB)
	if themeEvent(t, m, closedMsg) != nil {
		t.Fatal("duplicate old close touched replacement")
	}
	if themeEvent(t, m, light) != nil {
		t.Fatal("old read touched replacement")
	}
	m.applyUITheme("light")
	close(b.closeGate)
	_ = themeEvent(t, m, m.syncSystemTheme()())
}

func TestSystemThemeCanceledOpeningCannotPoisonNewSystemChoice(t *testing.T) {
	opens := 0
	m := themeModel(t, func(ctx context.Context) (appearance.Connection, error) { opens++; return nil, ctx.Err() })
	opening := m.syncSystemTheme()
	m.applyUITheme("light")
	m.applyUITheme("system")
	if m.syncSystemTheme() != nil {
		t.Fatal("opened concurrently")
	}
	reopen := themeEvent(t, m, opening())
	if reopen == nil || m.systemTheme.failed {
		t.Fatal("canceled old open poisoned latest choice")
	}
	_ = reopen()
	if opens != 2 {
		t.Fatal("replacement not started exactly once")
	}
}

func TestSystemThemeUnknownRetainsLastGoodAndSanitizesCaption(t *testing.T) {
	a := newControlledAppearance()
	m := themeModel(t, func(context.Context) (appearance.Connection, error) { return a, nil })
	read := themeEvent(t, m, m.syncSystemTheme()())
	a.values <- appearance.Observation{Theme: "light", Source: "windows"}
	read = themeEvent(t, m, read())
	a.values <- appearance.Observation{Theme: "dark", Source: "PRIVATE-SYNTHETIC", Err: errors.New("PRIVATE-SYNTHETIC")}
	read = themeEvent(t, m, read())
	m.densitySettings = &densitySettings{themeChoice: 2}
	caption := m.systemThemeCaption()
	if colorHex(m.View().BackgroundColor) != "d5d6db" || !strings.Contains(caption, "未知") || !strings.Contains(caption, "保留 Light") || strings.Contains(caption, "PRIVATE") {
		t.Fatal("unknown source changed palette or leaked source")
	}
	// A terminal reader failure closes once and does not busy-loop reopen.
	op := m.systemTheme.op
	closing := themeEvent(t, m, systemThemeRead{op, op.sequence, appearance.Observation{}, errors.New("PRIVATE-SYNTHETIC")})
	close(a.closeGate)
	if themeEvent(t, m, closing()) != nil || !m.systemTheme.failed || m.syncSystemTheme() != nil {
		t.Fatal("fatal failure reopened automatically")
	}
	if m.retrySystemTheme() == nil {
		t.Fatal("explicit retry unavailable")
	}
	_ = read
}

func TestSystemThemeCloseFailurePreventsReplacement(t *testing.T) {
	a := newControlledAppearance()
	a.closeErr = errors.New("PRIVATE-SYNTHETIC")
	m := themeModel(t, func(context.Context) (appearance.Connection, error) { return a, nil })
	_ = themeEvent(t, m, m.syncSystemTheme()())
	m.applyUITheme("light")
	closing := m.syncSystemTheme()
	m.applyUITheme("system")
	close(a.closeGate)
	if themeEvent(t, m, closing()) != nil || m.retrySystemTheme() != nil {
		t.Fatal("replaced a source whose cleanup failed")
	}
	m.densitySettings = &densitySettings{themeChoice: 2}
	if !strings.Contains(m.systemThemeCaption(), "重启") {
		t.Fatal("cleanup failure hidden")
	}
}

func TestSystemThemeRestylesOpenInputsWithoutLosingEditingState(t *testing.T) {
	input := func() textinput.Model {
		v := textinput.New()
		v.SetValue("synthetic 输入")
		v.SetCursor(3)
		v.Focus()
		return v
	}
	a := newControlledAppearance()
	m := themeModel(t, func(context.Context) (appearance.Connection, error) { return a, nil })
	read := themeEvent(t, m, m.syncSystemTheme()())
	m.launch = &launchDialog{fields: []textinput.Model{input()}}
	m.rename = &sessionRename{input: input()}
	m.search = &sessionSearch{input: input()}
	m.groupDialog = &groupDialog{rename: true, inputs: [2]textinput.Model{input(), input()}}
	a.values <- appearance.Observation{Theme: "light", Source: "windows"}
	_, _ = m.Update(read()) // Observation precedes all modal input handlers.
	for _, v := range []textinput.Model{m.launch.fields[0], m.rename.input, m.search.input, m.groupDialog.inputs[0], m.groupDialog.inputs[1]} {
		if v.Value() != "synthetic 输入" || v.Position() != 3 || !v.Focused() {
			t.Fatal("theme reset editing context")
		}
	}
}
