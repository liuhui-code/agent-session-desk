package ui

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestSwitchKeepsCompletedSamplesOfUnaffectedOriginals(t *testing.T) {
	testSwitchCompletedSample(t, false)
}

func TestSwitchRejectsCompletedSampleWhoseOwnViewportChanged(t *testing.T) {
	testSwitchCompletedSample(t, true)
}

func testSwitchCompletedSample(t *testing.T, resizeSampled bool) {
	t.Helper()
	h, _ := attentionUIFixture(t)
	m := h.m
	m.activity.busy = false
	m.activity.generation = 1
	active, other := m.sessions[0], m.sessions[1]
	m.activity.entries = map[string]activityEntry{activityKey(other): {hash: sha256.Sum256([]byte("quiet")), seen: time.Now().Add(-20 * time.Second)}}
	blocked := make(chan struct{})
	m.service = fakeService{preview: func(ctx context.Context, s session.Session, _ int) (session.Preview, error) {
		if s.ID == other.ID {
			return session.Preview{Text: "quiet"}, nil
		}
		close(blocked)
		<-ctx.Done()
		return session.Preview{}, ctx.Err()
	}}
	cmd := m.sampleActivity()
	if cmd == nil {
		t.Fatal("no batch")
	}
	receipt := make(chan activityReceipt, 1)
	go func() { receipt <- cmd().(activityReceipt) }()
	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("sample did not reach second target")
	}
	m.pauseActivity()
	if resizeSampled {
		m.workspace.active = 1
		active = other
	}
	m.activityViewport(40, 20)
	if activityKey(active) != m.activity.muteKey {
		t.Fatal("wrong resized original")
	}
	select {
	case x := <-receipt:
		m.activityEvent(x)
	case <-time.After(time.Second):
		t.Fatal("cancel did not finish batch")
	}
	if resizeSampled {
		if m.activityMark(other) != "" {
			t.Fatal("resized original reused a pre-layout sample")
		}
		return
	}
	if m.activityMark(other) != m.tokens.activityIdle {
		t.Fatal("switch discarded a completed fresh sample of an unaffected original")
	}
}
