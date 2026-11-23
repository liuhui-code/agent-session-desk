package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
	"time"
)

func TestTeaColorCapabilityChangesEmitDefaultResetsAndKnownTrueColorRequests(t *testing.T) {
	m := NewWithOptions(context.Background(), directDragFixture(), Options{Theme: "light"})
	defer m.Close()
	out := &clipboardOutput{wake: make(chan struct{}, 1)}
	p := tea.NewProgram(m, tea.WithInput(nil), tea.WithOutput(out), tea.WithWindowSize(120, 32), tea.WithEnvironment([]string{"TERM=xterm-256color"}), tea.WithoutSignalHandler(), tea.WithColorProfile(colorprofile.TrueColor))
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
	await := func(check func(string) bool) string {
		t.Helper()
		timer := time.NewTimer(3 * time.Second)
		defer timer.Stop()
		for {
			out.Lock()
			captured := out.buffer.String()
			out.Unlock()
			if check(captured) {
				return captured
			}
			select {
			case <-out.wake:
			case <-timer.C:
				t.Fatal("color transition output not observed")
			}
		}
	}
	bg, fg := ansi.SetBackgroundColor("#d5d6db"), ansi.SetForegroundColor("#343b58")
	await(func(s string) bool {
		return strings.Contains(s, "shell marker first") && strings.Contains(s, bg) && strings.Contains(s, fg)
	})
	p.Send(tea.ColorProfileMsg{Profile: colorprofile.ANSI256})
	captured := await(func(s string) bool {
		return strings.Contains(s, ansi.ResetBackgroundColor) && strings.Contains(s, ansi.ResetForegroundColor)
	})
	beforeBG, beforeFG := strings.Count(captured, bg), strings.Count(captured, fg)
	p.Send(tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	await(func(s string) bool { return strings.Count(s, bg) > beforeBG && strings.Count(s, fg) > beforeFG })
	// These are emitted framework requests/resets, not a real terminal's response
	// or evidence that a user's previous runtime colors were restored.
}
