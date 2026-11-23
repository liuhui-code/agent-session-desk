package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func releaseGroupOrderGate(gate chan struct{}) {
	select {
	case <-gate:
	default:
		close(gate)
	}
}

func TestGroupOrderQuitWaitsForMetadataAcceptedInputAndPreferences(t *testing.T) {
	s, _, _, _ := newGroupOrderFixture(t)
	metadataGate, inputGate, preferenceGate := make(chan struct{}), make(chan struct{}), make(chan struct{})
	for _, gate := range []chan struct{}{metadataGate, inputGate, preferenceGate} {
		t.Cleanup(func() { releaseGroupOrderGate(gate) })
	}
	s.gate = metadataGate
	h := newWorkspaceHarness(t, s, Options{SaveUISettings: func(ctx context.Context, _ config.UISettings) error {
		select {
		case <-preferenceGate:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return h.m.workspace.conn != nil })
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	stream.mu.Lock()
	stream.gate = inputGate
	stream.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "accepted-draft"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	if h.m.groupOrder == nil || !h.m.preferences.busy || h.m.workspace.pending == 0 {
		t.Fatal("test did not hold all three accepted effects")
	}
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.PasteMsg{Content: "must-not-leak"})
	if h.m.closed || h.m.workspace.focused {
		t.Fatal("quit abandoned commit or reassigned terminal focus")
	}
	close(metadataGate)
	h.until(func() bool { return h.m.groupOrder == nil })
	if h.m.closed {
		t.Fatal("metadata acknowledgement abandoned pending terminal input/preferences")
	}
	close(inputGate)
	h.until(func() bool { return h.m.workspace.pending == 0 })
	if h.m.closed {
		t.Fatal("input acknowledgement abandoned preference writer")
	}
	close(preferenceGate)
	h.until(func() bool { return h.m.closed })
	stream.mu.Lock()
	defer stream.mu.Unlock()
	writes := strings.Join(stream.writes, ",")
	if strings.Count(writes, "text:accepted-draft") != 1 || strings.Contains(writes, "must-not-leak") {
		t.Fatalf("quit lost/replayed input or leaked late paste: %s", writes)
	}
}

func TestGroupOrderQuitIsCancelledWhenPreferencesFailBeforeOrderAcknowledgement(t *testing.T) {
	s, _, _, _ := newGroupOrderFixture(t)
	metadataGate, preferenceGate := make(chan struct{}), make(chan struct{})
	for _, gate := range []chan struct{}{metadataGate, preferenceGate} {
		t.Cleanup(func() { releaseGroupOrderGate(gate) })
	}
	s.gate = metadataGate
	h := newWorkspaceHarness(t, s, Options{SaveUISettings: func(ctx context.Context, _ config.UISettings) error {
		select {
		case <-preferenceGate:
			return fmt.Errorf("synthetic preference failure")
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	close(preferenceGate)
	h.until(func() bool { return !h.m.preferences.busy && h.m.sidebarPreferenceError != nil })
	close(metadataGate)
	h.until(func() bool { return h.m.groupOrder == nil })
	if h.m.closed || !strings.Contains(h.m.View().Content, "侧栏偏好未保存") {
		t.Fatal("order acknowledgement hid preference failure and quit")
	}
}

func TestGroupOrderQuitIsCancelledWhenTerminalFailsBeforeOrderAcknowledgement(t *testing.T) {
	s, _, _, _ := newGroupOrderFixture(t)
	gate := make(chan struct{})
	t.Cleanup(func() { releaseGroupOrderGate(gate) })
	s.gate = gate
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnd})
	h.send(tea.KeyPressMsg{Code: 'K', Text: "K"})
	h.send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	h.send(workspaceFrame{generation: h.m.workspace.generation, err: fmt.Errorf("synthetic terminal disconnect")})
	close(gate)
	h.until(func() bool { return h.m.groupOrder == nil })
	if h.m.closed || h.m.workspace.err == nil {
		t.Fatal("order acknowledgement hid terminal failure and quit")
	}
}
