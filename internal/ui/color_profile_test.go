package ui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"context"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/config"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestReportedANSI256UsesPinnedSourcePaletteWithoutChangingAuthorTokens(t *testing.T) {
	m := NewWithOptions(context.Background(), newGroupedFixture(), Options{Footer: "compact", SidebarMode: "flat"})
	defer m.Close()
	m.width, m.height = 160, 32
	m.sessions, m.cursor, m.workspace.focused = sampleSessions(), 0, false
	_, cmd := m.Update(tea.ColorProfileMsg{Profile: colorprofile.ANSI256})
	if cmd != nil {
		t.Fatal("color capability started a command")
	}
	line := m.renderManagementFooter()
	plain := ansi.Strip(line)
	i := strings.Index(plain, "Enter")
	if i < 0 {
		t.Fatal("missing Enter hint")
	}
	x := ansi.StringWidth(plain[:i])
	if fg, bg := testForegroundAt(line, x), testBackgroundAt(line, x); fg != "idx:0" || bg != "idx:7" {
		t.Fatalf("source palette not selected: fg=%s bg=%s", fg, bg)
	}
	if colorHex(m.tokens.colors.bg) != "000000" || colorHex(m.tokens.colors.accent) != "bbbbbb" {
		t.Fatal("capability replaced authored palette")
	}
}

func TestColorProfileEventsKeepActiveRunInputAndPreferences(t *testing.T) {
	s := directDragFixture()
	h := newWorkspaceHarness(t, s, Options{SaveUISettings: func(context.Context, config.UISettings) error { t.Error("capability saved preferences"); return nil }})
	h.until(func() bool { return h.m.workspace.conn != nil })
	conn, run, saved := h.m.workspace.conn, h.m.workspace.tabs[h.m.workspace.active].session, h.m.preferences.saved
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "draft"})
	h.until(func() bool { return h.m.workspace.pending == 0 })
	for _, p := range []colorprofile.Profile{colorprofile.ANSI256, colorprofile.ANSI, colorprofile.ASCII, colorprofile.TrueColor} {
		h.send(tea.ColorProfileMsg{Profile: p})
	}
	if h.m.workspace.conn != conn || !sameRun(run, h.m.workspace.tabs[h.m.workspace.active].session) || h.m.preferences.saved != saved || !h.m.workspace.focused {
		t.Fatal("capability changed terminal ownership or saved state")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	count := 0
	for _, v := range stream.writes {
		if v == "text:draft" {
			count++
		} else if strings.HasPrefix(v, "text:") || strings.HasPrefix(v, "key:") {
			t.Errorf("capability injected input: %s", v)
		}
	}
	if count != 1 {
		t.Fatal("accepted terminal input lost or repeated")
	}
}

func TestColorProfileRestylesModalWithoutResettingInputOrAcceptingUnknown(t *testing.T) {
	m := NewWithOptions(context.Background(), nil, Options{Theme: "light"})
	defer m.Close()
	input := func() textinput.Model {
		v := textinput.New()
		v.SetValue("synthetic 中文")
		v.SetCursor(3)
		v.Focus()
		return v
	}
	m.launch = &launchDialog{fields: []textinput.Model{input()}}
	m.rename = &sessionRename{input: input()}
	m.search = &sessionSearch{input: input()}
	m.groupDialog = &groupDialog{inputs: [2]textinput.Model{input(), input()}, rename: true}
	for _, p := range []colorprofile.Profile{colorprofile.ANSI, colorprofile.ASCII, colorprofile.ANSI256, colorprofile.TrueColor} {
		_, cmd := m.Update(tea.ColorProfileMsg{Profile: p})
		if cmd != nil {
			t.Fatal("modal capability started a command")
		}
		for _, v := range []textinput.Model{m.launch.fields[0], m.rename.input, m.search.input, m.groupDialog.inputs[0], m.groupDialog.inputs[1]} {
			if v.Value() != "synthetic 中文" || v.Position() != 3 || !v.Focused() {
				t.Fatal("capability reset modal input state")
			}
		}
	}
	for _, p := range []colorprofile.Profile{colorprofile.Unknown, colorprofile.Profile(255)} {
		m.Update(tea.ColorProfileMsg{Profile: p})
		if m.colorProfile != colorprofile.TrueColor {
			t.Fatal("unreliable capability replaced known value")
		}
	}
	m.Close()
	m.Update(tea.ColorProfileMsg{Profile: colorprofile.ANSI})
	if m.colorProfile != colorprofile.TrueColor {
		t.Fatal("closed UI accepted a capability event")
	}
}

func TestLimitedColorProgramsNeverRequestRGBDefaults(t *testing.T) {
	for _, profile := range []colorprofile.Profile{colorprofile.ANSI256, colorprofile.ANSI, colorprofile.ASCII, colorprofile.NoTTY} {
		t.Run(profile.String(), func(t *testing.T) {
			m := NewWithOptions(context.Background(), directDragFixture(), Options{Theme: "light"})
			defer m.Close()
			out := &clipboardOutput{wake: make(chan struct{}, 1)}
			p := tea.NewProgram(m, tea.WithInput(nil), tea.WithOutput(out), tea.WithWindowSize(120, 32), tea.WithEnvironment([]string{"TERM=xterm-256color"}), tea.WithoutSignalHandler(), tea.WithColorProfile(profile))
			done := make(chan error, 1)
			go func() { _, err := p.Run(); done <- err }()
			defer func() {
				p.Quit()
				select {
				case err := <-done:
					if err != nil {
						t.Error(err)
					}
				case <-time.After(3 * time.Second):
					p.Kill()
					<-done
					t.Error("program did not quit normally")
				}
			}()
			timer := time.NewTimer(3 * time.Second)
			defer timer.Stop()
			for {
				out.Lock()
				captured := out.buffer.String()
				out.Unlock()
				if strings.Contains(captured, "shell marker first") {
					if regexp.MustCompile("\\x1b\\](10|11);").MatchString(captured) {
						t.Fatal("limited-color program queried or set RGB defaults")
					}
					break
				}
				select {
				case <-out.wake:
				case <-timer.C:
					t.Fatal("program frame not observed")
				}
			}
		})
	}
}

func TestRGBDefaultColorsRequireKnownTrueColorAndClearOnLimitedProfiles(t *testing.T) {
	m := NewWithOptions(context.Background(), newGroupedFixture(), Options{Theme: "light"})
	defer m.Close()
	for _, p := range []colorprofile.Profile{colorprofile.Unknown, colorprofile.TrueColor, colorprofile.ANSI256, colorprofile.ANSI, colorprofile.ASCII, colorprofile.NoTTY, colorprofile.TrueColor} {
		_, cmd := m.Update(tea.ColorProfileMsg{Profile: p})
		if cmd != nil {
			t.Fatal("capability change emitted a command")
		}
		view := m.View()
		if p == colorprofile.TrueColor {
			if colorHex(view.BackgroundColor) != "d5d6db" || colorHex(view.ForegroundColor) != "343b58" {
				t.Fatal("known TrueColor did not restore configured default colors")
			}
		} else if view.BackgroundColor != nil || view.ForegroundColor != nil {
			t.Fatalf("%s requested RGB default colors", p)
		}
	}
}
