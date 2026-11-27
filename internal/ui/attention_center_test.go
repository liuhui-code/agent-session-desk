package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestAttentionCenterOpensWithoutViewingAndRestoresOriginalInputFocus(t *testing.T) {
	h, s := attentionUIFixture(t)
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	h.m.workspace.sidebarHidden = true
	before := h.m.workspace.active
	h.send(tea.KeyPressMsg{Code: 'a', Mod: tea.ModAlt})
	h.until(func() bool { return strings.Contains(ansi.Strip(h.m.View().Content), "Pending items") })
	h.send(tea.PasteMsg{Content: "y\n"})
	if s.visits != 0 || h.m.workspace.active != before {
		t.Fatal("opening/paste viewed or answered a request")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEsc})
	h.until(func() bool { return h.m.attention.center == nil && h.m.workspace.conn != nil })
	if !h.m.workspace.focused || !h.m.workspace.sidebarHidden {
		t.Fatal("center changed original focus/sidebar")
	}
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if len(stream.writes) != 0 {
		t.Fatal("modal input reached terminal", stream.writes)
	}
}

func TestAttentionCenterFailedDrainReleasesNavigationIntent(t *testing.T) {
	h, s := attentionUIFixture(t)
	s.mu.Lock()
	first := s.streams["first"]
	s.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'x', Text: "accepted"})
	h.send(tea.KeyPressMsg{Code: 'a', Mod: tea.ModAlt})
	h.send(workspaceInputDone{generation: h.m.workspace.generation, err: fmt.Errorf("write disconnected")})
	close(gate)
	if h.m.attention.centerIntent != nil || h.m.attention.center != nil || h.m.workspace.transition != workspaceIdle {
		t.Fatal("failed drain left an invisible input trap")
	}
	h.send(tea.KeyPressMsg{Code: 'a', Mod: tea.ModAlt})
	if h.m.attention.center == nil {
		t.Fatal("management navigation remained trapped")
	}
}

func TestSnoozedAttentionKeepsCenterRecordButSuppressesQuickNavigation(t *testing.T) {
	h, s := attentionUIFixture(t)
	until := time.Now().Add(10 * time.Minute)
	s.item.SnoozedUntil = &until
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	h.send(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	if h.m.attention.busy || h.m.workspace.active != 0 || s.visits != 0 || h.m.sessionNeedsAttention(s.item.Source) {
		t.Fatal("snoozed item still interrupted quick navigation")
	}
	h.send(tea.KeyPressMsg{Code: 'a', Mod: tea.ModAlt})
	h.until(func() bool { return h.m.attention.center != nil })
	text := ansi.Strip(h.m.View().Content)
	if !strings.Contains(text, "1 items") || !strings.Contains(text, "Snoozed") {
		t.Fatal("snooze removed unresolved request", text)
	}
}

func TestAttentionCenterDrainsAcceptedInputAndRejectsLaterPaste(t *testing.T) {
	h, s := attentionUIFixture(t)
	s.mu.Lock()
	first := s.streams["first"]
	s.mu.Unlock()
	gate := make(chan struct{})
	first.mu.Lock()
	first.gate = gate
	first.mu.Unlock()
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	h.send(tea.KeyPressMsg{Code: 'x', Text: "accepted draft"})
	h.send(tea.KeyPressMsg{Code: 'a', Mod: tea.ModAlt})
	h.send(tea.PasteMsg{Content: "y\n"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	close(gate)
	h.until(func() bool { return h.m.attention.center != nil })
	first.mu.Lock()
	writes := strings.Join(first.writes, ",")
	first.mu.Unlock()
	if writes != "text:accepted draft" {
		t.Fatal("late modal input reached original terminal", writes)
	}
}
