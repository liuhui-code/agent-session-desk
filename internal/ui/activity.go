package ui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"crypto/sha256"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"time"
)

const (
	activityQuiet = 3 * time.Second
	// Allow ordinary serial capture delays without erasing the last observation.
	activityFreshFor = 15 * time.Second
)

type activityEntry struct {
	hash          [32]byte
	seen, changed time.Time
}

func (a *activityEntry) observe(h [32]byte, now time.Time) {
	if !a.seen.IsZero() && a.hash != h {
		a.changed = now
	}
	a.hash = h
	a.seen = now
}
func (a activityEntry) moving(now time.Time) bool {
	return !a.changed.IsZero() && now.Sub(a.changed) < activityQuiet
}

type activityState struct {
	cancel             context.CancelFunc
	epoch              uint64
	layoutKey, muteKey string
	width, height      int
	mutedUntil         time.Time
	entries            map[string]activityEntry
	busy, ticking      bool
	generation         uint64
}
type activitySample struct {
	key  string
	hash [32]byte
	ok   bool
}
type activityReceipt struct {
	epoch      uint64
	generation uint64
	samples    []activitySample
}
type activityTick struct{}

func activityKey(s session.Session) string { return s.ID + ":" + s.RunID + ":" + s.BackendID }
func (m *Model) sampleActivity() tea.Cmd {
	if m.activity.busy || m.workspace == nil || m.closed || m.attaching || m.workspace.busy || m.workspace.transition != workspaceIdle {
		return nil
	}
	g := m.workspaceGeometry()
	width, height := g.left, g.bodyHeight
	if g.above > 0 {
		width, height = g.listWidth, g.listHeight
	}
	if width <= 0 {
		return nil
	}
	var targets []session.Session
	for _, r := range m.sidebarNavigationWindow(width, height).rows {
		if r.index >= 0 && r.index < len(m.sessions) {
			s := m.sessions[r.index]
			if s.TerminalState == session.TerminalAvailable {
				targets = append(targets, s)
			}
			if len(targets) == 8 {
				break
			}
		}
	}
	if len(targets) == 0 {
		return nil
	}
	m.activity.busy = true
	m.activity.generation++
	gen := m.activity.generation
	epoch := m.activity.epoch
	svc := m.service
	ctx, cancel := context.WithTimeout(m.ctx, 2*time.Second)
	m.activity.cancel = cancel
	return func() tea.Msg {
		defer cancel()
		samples := make([]activitySample, 0, len(targets))
		// Rotate the first pane so a shared deadline cannot starve later rows.
		for i := range targets {
			s := targets[(i+int((gen-1)%uint64(len(targets))))%len(targets)]
			p, err := svc.Preview(ctx, s, 100)
			v := activitySample{key: activityKey(s), ok: err == nil && len(p.Text) <= 65536}
			if v.ok {
				v.hash = sha256.Sum256([]byte(p.Text))
			}
			samples = append(samples, v)
		}
		return activityReceipt{generation: gen, epoch: epoch, samples: samples}
	}
}
func (m *Model) activityTick() tea.Cmd {
	if m.activity.ticking || m.closed {
		return nil
	}
	now := time.Now()
	moving := false
	for _, e := range m.activity.entries {
		moving = moving || e.moving(now)
	}
	if !moving {
		return nil
	}
	m.activity.ticking = true
	return tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg { return activityTick{} })
}
func (m *Model) activityEvent(msg tea.Msg) (tea.Cmd, bool) {
	switch x := msg.(type) {
	case activityReceipt:
		if x.generation != m.activity.generation {
			return nil, true
		}
		m.activity.busy = false
		if x.epoch != m.activity.epoch {
			return m.activityTick(), true
		}
		next := make(map[string]activityEntry, len(x.samples))
		now := time.Now()
		for _, v := range x.samples {
			e, known := m.activity.entries[v.key]
			if !v.ok {
				if known && now.Sub(e.seen) <= activityFreshFor {
					next[v.key] = e
				}
				continue
			}
			if v.key == m.activity.muteKey && (now.Before(m.activity.mutedUntil) || e.seen.Before(m.activity.mutedUntil.Add(-activityQuiet))) {
				e.seen = time.Time{}
				e.changed = time.Time{}
			}
			e.observe(v.hash, now)
			next[v.key] = e
		}
		m.activity.entries = next
		return m.activityTick(), true
	case activityTick:
		m.activity.ticking = false
		return m.activityTick(), true
	}
	return nil, false
}
func (m *Model) activityMark(s session.Session) string {
	e, ok := m.activity.entries[activityKey(s)]
	now := time.Now()
	if !ok || now.Sub(e.seen) > activityFreshFor || m.listErr != nil {
		return ""
	}
	if e.moving(now) {
		f := m.tokens.activityFrames
		return f[int(now.UnixNano()/int64(150*time.Millisecond))%len(f)]
	}
	return m.tokens.activityIdle
}

// A view resize can reflow captured lines without any process output. Keep only
// the active identity and one settle deadline, never a timer per Session.
func (m *Model) activityViewport(width, height int) {
	w := m.workspace
	if w == nil || w.active < 0 || w.active >= len(w.tabs) {
		return
	}
	key := activityKey(w.tabs[w.active].session)
	a := &m.activity
	if a.layoutKey == key && a.width == width && a.height == height {
		return
	}
	a.layoutKey = key
	a.width = width
	a.height = height
	a.epoch++
	a.muteKey = key
	a.mutedUntil = time.Now().Add(activityQuiet)
	if e, ok := a.entries[key]; ok {
		e.changed = time.Time{}
		a.entries[key] = e
	}
}

// Connection work has priority over optional passive snapshots. Old receipts
// may clear busy, but epoch checking prevents them from updating observations.
func (m *Model) pauseActivity() {
	if m.activity.cancel != nil {
		m.activity.cancel()
		m.activity.cancel = nil
	}
	m.activity.epoch++
}
