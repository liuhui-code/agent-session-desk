package pet

import (
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func liveItem(id, kind string) session.AttentionItem {
	return session.AttentionItem{AttentionEvent: session.AttentionEvent{ID: strings.Repeat(id, 32), Version: 1, Kind: kind, State: "pending", Title: "existing blocking prompt"}, TerminalAvailable: true}
}

func TestLivePetUsesExistingPendingAndViewedNeverClearsIt(t *testing.T) {
	a, b := liveItem("a", "approval"), liveItem("b", "question")
	model := project([]session.AttentionItem{a, b}, true)
	if model.Kind != Approval || model.Count != 2 || model.Label() != "WAIT" {
		t.Fatalf("projection %+v %q", model, model.Label())
	}
	now := time.Now()
	a.ViewedAt = &now
	b.ViewedAt = &now
	model = project([]session.AttentionItem{a, b}, true)
	if model.Count != 2 || model.Label() == "DEMO" {
		t.Fatal("viewed request became resolved or simulated")
	}
	a.State = "resolved"
	model = project([]session.AttentionItem{a, b}, true)
	if model.Kind != Question || model.Count != 1 {
		t.Fatal("source resolution not reflected")
	}
	model = project([]session.AttentionItem{a, b}, false)
	if model.Kind != Unknown || model.Count != 1 {
		t.Fatal("disconnect resolved pending or claimed freshness")
	}
	b.State = "invalidated"
	model = project([]session.AttentionItem{a, b}, true)
	if model.Kind != Idle || model.Count != 0 || model.Label() != "ASD" {
		t.Fatal("closed source retained an active marker")
	}
}

func TestLivePetClickOnlyRequestsExactNavigationAndDoesNotMutatePending(t *testing.T) {
	var requests []Activation
	c := NewCompanion(func(a Activation) { requests = append(requests, a) })
	a, b := liveItem("a", "approval"), liveItem("b", "question")
	now := time.Now()
	a.ViewedAt = &now
	items := []session.AttentionItem{a, b}
	c.Publish(items, true)
	items[1].State = "resolved"
	if !c.Activate() {
		t.Fatal("pending activation missing")
	}
	if len(requests) != 1 || requests[0].ID != b.ID || requests[0].Version != b.Version {
		t.Fatal("click lost request identity/version")
	}
	if view := c.snapshot(); project(view.items, view.known).Count != 2 {
		t.Fatal("navigation cleared pending or retained caller slice")
	}
	c.Publish([]session.AttentionItem{a, b}, false)
	if c.Activate() {
		t.Fatal("stale view authorized navigation")
	}
}
