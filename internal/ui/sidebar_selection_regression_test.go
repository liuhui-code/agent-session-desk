package ui

import (
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
	"time"
)

func TestGroupSelectionKeepsTitleColumn(t *testing.T) {
	h, _ := attentionUIFixture(t)
	m := h.m
	m.sidebarMode = "grouped"
	m.workspace.focused = false
	m.selectedGroupID = "default"
	m.cursor = 0
	a := ansi.Strip(m.renderSidebarGroupAt("default", 1, 40))
	m.cursor = -1
	b := ansi.Strip(m.renderSidebarGroupAt("default", 1, 40))
	if strings.Index(a, "My Sessions") != strings.Index(b, "My Sessions") {
		t.Fatalf("selection shifted: %q / %q", a, b)
	}
}
func TestActivityIndicatorBlueAcrossSelection(t *testing.T) {
	h, _ := attentionUIFixture(t)
	s := h.m.sessions[0]
	h.m.tokens.activityFrames = []string{" ◐"}
	h.m.activity.entries = map[string]activityEntry{activityKey(s): {seen: time.Now(), changed: time.Now()}}
	r, g, b, _ := h.m.tokens.colors.activityBlue.RGBA()
	want := fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8)
	for _, focused := range []bool{true, false} {
		h.m.workspace.focused = focused
		row := h.m.workspaceSidebarCard(0, 40, 2, 1)[0]
		plain := ansi.Strip(row)
		mark := h.m.activityMark(s)
		x := ansi.StringWidth(plain[:strings.LastIndex(plain, mark)]) + 1
		if testForegroundAt(row, x) != want {
			t.Fatal("activity color changed", testForegroundAt(row, x), want)
		}
	}
}
func TestLayoutChangeDoesNotBecomeOutput(t *testing.T) {
	h, _ := attentionUIFixture(t)
	m := h.m
	s := m.sessions[0]
	key := activityKey(s)
	m.activity.entries = map[string]activityEntry{key: {hash: [32]byte{1}, seen: time.Now()}}
	old := m.activity.epoch
	m.activityViewport(40, 20)
	m.activity.generation = 10
	m.activity.busy = true
	h.send(activityReceipt{generation: 10, epoch: old, samples: []activitySample{{key: key, hash: [32]byte{2}, ok: true}}})
	if m.activity.busy || m.activity.entries[key].moving(time.Now()) {
		t.Fatal("stale sample after layout")
	}
	h.send(activityReceipt{generation: 10, epoch: m.activity.epoch, samples: []activitySample{{key: key, hash: [32]byte{3}, ok: true}}})
	if m.activity.entries[key].moving(time.Now()) {
		t.Fatal("reflow became activity")
	}
	m.activity.mutedUntil = time.Now().Add(-time.Second)
	h.send(activityReceipt{generation: 10, epoch: m.activity.epoch, samples: []activitySample{{key: key, hash: [32]byte{4}, ok: true}}})
	if !m.activity.entries[key].moving(time.Now()) {
		t.Fatal("real output suppressed after settle")
	}
}

func TestDelayedLayoutSampleStartsNewBaseline(t *testing.T) {
	h, _ := attentionUIFixture(t)
	m := h.m
	key := activityKey(m.sessions[0])
	m.activity.entries = map[string]activityEntry{key: {hash: [32]byte{1}, seen: time.Now().Add(-10 * time.Second)}}
	m.activityViewport(40, 20)
	m.activity.mutedUntil = time.Now().Add(-time.Second)
	h.send(activityReceipt{epoch: m.activity.epoch, samples: []activitySample{{key: key, hash: [32]byte{2}, ok: true}}})
	if m.activity.entries[key].moving(time.Now()) {
		t.Fatal("delayed reflow became output")
	}
}
