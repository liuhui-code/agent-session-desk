package pet

import (
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"testing"
	"time"
)

func TestSnoozedItemDoesNotAlertOrNavigateButExpiredItemReturns(t *testing.T) {
	item := liveItem("a", "approval")
	until := time.Now().Add(time.Minute)
	item.SnoozedUntil = &until
	model := project([]session.AttentionItem{item}, true)
	if model.Kind != Idle || model.Count != 0 {
		t.Fatal("snoozed reminder still alerts", model)
	}
	var got Activation
	c := NewCompanion(func(a Activation) { got = a })
	c.Publish([]session.AttentionItem{item}, true)
	if !c.Activate() || got.ID != "" {
		t.Fatal("snoozed pet click opened an item")
	}
	past := time.Now().Add(-time.Second)
	item.SnoozedUntil = &past
	if model = project([]session.AttentionItem{item}, true); model.Kind != Approval || model.Count != 1 {
		t.Fatal("expired reminder did not return")
	}
}
