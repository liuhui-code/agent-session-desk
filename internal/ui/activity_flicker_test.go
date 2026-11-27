package ui

import (
	"context"
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"strings"
	"testing"
	"time"
)

func TestIdleSurvivesOrdinarySamplingDelay(t *testing.T) {
	h, _ := attentionUIFixture(t)
	s := h.m.sessions[0]
	h.m.activity.entries = map[string]activityEntry{activityKey(s): {seen: time.Now().Add(-4 * time.Second)}}
	if got := h.m.activityMark(s); got != h.m.tokens.activityIdle {
		t.Fatalf("quiet dot vanished between polls: %q", got)
	}
}
func TestIdleSurvivesOnePreviewFailure(t *testing.T) {
	h, _ := attentionUIFixture(t)
	s := h.m.sessions[0]
	h.m.activity.entries = map[string]activityEntry{activityKey(s): {seen: time.Now()}}
	h.send(activityReceipt{generation: h.m.activity.generation, epoch: h.m.activity.epoch, samples: []activitySample{{key: activityKey(s), ok: false}}})
	if got := h.m.activityMark(s); got != h.m.tokens.activityIdle {
		t.Fatalf("single preview failure removed quiet dot: %q", got)
	}
}

func TestActivitySamplingRotatesFirstTarget(t *testing.T) {
	var first []string
	calls := 0
	h, _ := attentionUIFixture(t)
	m := h.m
	m.activity.busy = false
	m.service = fakeService{preview: func(_ context.Context, s session.Session, _ int) (session.Preview, error) {
		calls++
		if calls == 1 {
			first = append(first, s.ID)
			return session.Preview{Text: "quiet"}, nil
		}
		return session.Preview{}, context.DeadlineExceeded
	}}
	m.width, m.height = 120, 32
	for i := range m.sessions {
		m.sessions[i].TerminalState = session.TerminalAvailable
	}
	m.sidebarMode = "flat"
	for range 2 {
		calls = 0
		cmd := m.sampleActivity()
		if cmd == nil {
			t.Fatal("no sample")
		}
		m.activityEvent(cmd())
	}
	if len(first) != 2 || first[0] == first[1] {
		t.Fatalf("same pane always first within shared deadline: %v", first)
	}
}

func TestActivityFailureStillExpiresAndPrunes(t *testing.T) {
	h, _ := attentionUIFixture(t)
	m := h.m
	s := m.sessions[0]
	key := activityKey(s)
	m.activity.entries = map[string]activityEntry{key: {seen: time.Now().Add(-time.Minute)}, "removed": {seen: time.Now()}}
	h.send(activityReceipt{generation: m.activity.generation, epoch: m.activity.epoch, samples: []activitySample{{key: key}}})
	if len(m.activity.entries) != 0 || m.activityMark(s) != "" {
		t.Fatal("stale or removed observation retained")
	}
}

func TestIdleAndSpinnerKeepBlueOnSelectedSurface(t *testing.T) {
	h, _ := attentionUIFixture(t)
	m := h.m
	s := m.sessions[0]
	m.cursor = 0
	m.workspace.focused = false
	symbols := append([]string{m.tokens.activityIdle}, m.tokens.activityFrames...)
	r, g, b, _ := m.tokens.colors.activityBlue.RGBA()
	want := fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8)
	for i, symbol := range symbols {
		e := activityEntry{seen: time.Now()}
		if i > 0 {
			e.changed = time.Now()
			m.tokens.activityFrames = []string{symbol}
		}
		m.activity.entries = map[string]activityEntry{activityKey(s): e}
		row := m.workspaceSidebarCard(0, 40, 2, 1)[0]
		plain := ansi.Strip(row)
		index := strings.LastIndex(plain, symbol)
		if index < 0 {
			t.Fatalf("missing indicator %q", symbol)
		}
		start := ansi.StringWidth(plain[:index])
		bg := testBackgroundAt(row, 0)
		for x := start; x < start+ansi.StringWidth(symbol); x++ {
			if testForegroundAt(row, x) != want || testBackgroundAt(row, x) != bg {
				t.Fatalf("indicator lost color/surface %q at %d", symbol, x)
			}
		}
	}
}
