package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestGroupRemovalDefaultNoticeAndTerminalLettersNeverMutateMetadata(t *testing.T) {
	s, _, _, _ := newGroupRemovalFixture(t)
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	for _, text := range []string{"d", "y", "n"} {
		h.send(tea.KeyPressMsg{Code: []rune(text)[0], Text: text})
	}
	h.until(func() bool { return h.m.workspace.pending == 0 })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: tea.KeyHome})
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	h.until(func() bool { return strings.Contains(h.m.View().Content, "Cannot delete the default group") })
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.send(tea.KeyPressMsg{Code: tea.KeyPgDown, Mod: tea.ModCtrl})
	if h.m.groupRemoval == nil || h.m.workspace.active != 0 {
		t.Fatal("notice accepted deletion or switched terminal")
	}
	h.send(tea.KeyPressMsg{Code: 'o', Text: "o"})
	h.until(func() bool { return h.m.groupRemoval == nil && h.m.workspace.conn != nil })
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.previews) != 0 || len(s.removals) != 0 {
		t.Fatal("protected default reached metadata mutation service")
	}
}

func TestGroupRemovalLoadingCancelAndStalePreviewNeverDeferConfirmation(t *testing.T) {
	s, _, _, _ := newGroupRemovalFixture(t)
	gate := make(chan struct{})
	t.Cleanup(func() { releaseGroupOrderGate(gate) })
	s.previewGate = gate
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '2', Text: "2"})
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	h.until(func() bool { return h.m.groupRemoval != nil && h.m.groupRemoval.loading })
	old := h.m.groupRemoval
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	h.send(tea.KeyPressMsg{Code: tea.KeyEsc})
	h.until(func() bool { return h.m.groupRemoval == nil && h.m.workspace.conn != nil })
	close(gate)
	s.groupMu.Lock()
	s.previewGate = nil
	s.groupMu.Unlock()
	openReviewedRemoval(h)
	current := h.m.groupRemoval
	h.send(groupRemovalPreview{dialog: old, plan: s.plan})
	if h.m.groupRemoval != current || h.m.groupRemoval.busy {
		t.Fatal("old preview changed a new confirmation or deferred y")
	}
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.removals) != 0 {
		t.Fatal("cancel/loading caused a removal")
	}
}

func TestGroupRemovalWaitsForAcceptedInputAndBusyConfirmationOnlyCommitsOnce(t *testing.T) {
	s, _, _, _ := newGroupRemovalFixture(t)
	inputGate, commitGate := make(chan struct{}), make(chan struct{})
	for _, gate := range []chan struct{}{inputGate, commitGate} {
		t.Cleanup(func() { releaseGroupOrderGate(gate) })
	}
	s.removeGate = commitGate
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	s.mu.Lock()
	stream := s.streams["first"]
	s.mu.Unlock()
	stream.mu.Lock()
	stream.gate = inputGate
	stream.mu.Unlock()
	h.send(tea.KeyPressMsg{Code: 'd', Text: "accepted-before-dialog"})
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	h.send(tea.KeyPressMsg{Code: '2', Text: "2"})
	h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})
	if h.m.groupRemoval != nil {
		t.Fatal("dialog cancelled accepted terminal input")
	}
	close(inputGate)
	h.until(func() bool { return h.m.groupRemoval != nil && !h.m.groupRemoval.loading })
	h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	dialog := h.m.groupRemoval
	for _, key := range []tea.KeyPressMsg{{Code: tea.KeyEnter}, {Code: 'y', Text: "y"}, {Code: tea.KeyEsc}, {Code: tea.KeyF6}, {Code: 'q', Text: "q"}} {
		h.send(key)
	}
	h.send(tea.PasteMsg{Content: "must-not-enter-terminal"})
	if h.m.groupRemoval != dialog || !dialog.busy || h.m.closed {
		t.Fatal("busy confirmation was abandoned")
	}
	close(commitGate)
	h.until(func() bool { return h.m.groupRemoval == nil && h.m.workspace.conn != nil })
	s.groupMu.Lock()
	defer s.groupMu.Unlock()
	if len(s.removals) != 1 {
		t.Fatal("confirmation was repeated")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	writes := strings.Join(stream.writes, ",")
	if strings.Count(writes, "text:accepted-before-dialog") != 1 || strings.Contains(writes, "must-not-enter-terminal") {
		t.Fatalf("accepted input lost/replayed or modal text leaked: %s", writes)
	}
}

func TestGroupRemovalFailuresStayVisibleAndNeverResubmit(t *testing.T) {
	for _, stage := range []string{"preview", "commit"} {
		t.Run(stage, func(t *testing.T) {
			s, root, _, _ := newGroupRemovalFixture(t)
			if stage == "preview" {
				s.previewErr = fmt.Errorf("synthetic preview failure")
			} else {
				s.removeErr = fmt.Errorf("synthetic concurrent metadata change")
			}
			h := newWorkspaceHarness(t, s, Options{})
			h.until(func() bool { return h.m.workspace.conn != nil })
			h.send(tea.KeyPressMsg{Code: tea.KeyF6})
			openReviewedRemoval(h)
			h.send(tea.KeyPressMsg{Code: tea.KeyLeft})
			h.send(tea.KeyPressMsg{Code: 'Y', Text: "Y"})
			h.until(func() bool { return h.m.groupRemoval != nil && !h.m.groupRemoval.busy && h.m.groupRemoval.err != nil })
			h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
			h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
			if h.m.groupRemoval == nil || h.m.groupByID(root.ID).ID == "" {
				t.Fatal("failure hidden or uncommitted tree changed")
			}
			h.send(tea.KeyPressMsg{Code: 'N', Text: "N"})
			h.until(func() bool { return h.m.groupRemoval == nil && h.m.workspace.conn != nil })
			s.groupMu.Lock()
			defer s.groupMu.Unlock()
			want := 1
			if stage == "preview" {
				want = 0
			}
			if len(s.removals) != want {
				t.Fatalf("failed %s was repeated", stage)
			}
		})
	}
}

func TestGroupRemovalAcknowledgementPreservesANewerObservedRunAndMembership(t *testing.T) {
	s, root, _, _ := newGroupRemovalFixture(t)
	gate := make(chan struct{})
	t.Cleanup(func() { releaseGroupOrderGate(gate) })
	s.removeGate = gate
	h := newWorkspaceHarness(t, s, Options{})
	h.until(func() bool { return h.m.workspace.conn != nil })
	h.send(tea.KeyPressMsg{Code: tea.KeyF6})
	openReviewedRemoval(h)
	h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	dialog := h.m.groupRemoval
	newer := sampleSessions()
	newer[0].RunID = "replacement-run"
	newer[0].BackendID = "replacement-backend"
	newer[0].Title = "newer name"
	newer[0].TitleRevision = 2
	newer[0].GroupRevision = 3
	h.send(sessionsMsg{generation: h.m.listGeneration, sessions: newer, groups: s.result.Groups})
	// Inspect application before running any newly scheduled backend command.
	h.m.Update(groupRemovalResult{dialog: dialog, view: s.result})
	if h.m.sessions[0] != newer[0] || h.m.workspace.tabs[0].session.RunID != sampleSessions()[0].RunID || h.m.groupByID(root.ID).ID != "" {
		t.Fatal("removal receipt replaced newer runtime/name/membership or recreated removed group")
	}
}
