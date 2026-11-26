package ui

import (
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestCompanionActivationReusesAttentionNavigationWithoutAnswering(t *testing.T) {
	h, s := attentionUIFixture(t)
	var published []session.AttentionItem
	h.m.attentionObserver = func(items []session.AttentionItem, known bool) {
		if known {
			published = items
		}
	}
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	if len(published) != 1 {
		t.Fatal("existing blocker not published")
	}
	h.send(AttentionActivate{ID: s.item.ID, Version: s.item.Version})
	h.until(func() bool { return h.m.workspace.active == 1 && h.m.workspace.conn != nil && !h.m.attention.busy })
	if len(published) != 1 || published[0].ViewedAt == nil || published[0].State != "pending" {
		t.Fatal("companion view resolved request")
	}
	s.mu.Lock()
	stream := s.streams[s.item.Source.ID]
	s.mu.Unlock()
	stream.mu.Lock()
	writes := len(stream.writes)
	stream.mu.Unlock()
	if writes != 0 {
		t.Fatal("companion approved or submitted input")
	}
}

func TestCompanionStaleActivationDoesNotOpenNewerPrompt(t *testing.T) {
	h, s := attentionUIFixture(t)
	h.send(sessionsMsg{sessions: sampleSessions(), attention: []session.AttentionItem{s.item}})
	h.send(AttentionActivate{ID: s.item.ID, Version: s.item.Version + 1})
	if h.m.attention.busy || h.m.workspace.active != 0 {
		t.Fatal("stale companion activation navigated")
	}
}
