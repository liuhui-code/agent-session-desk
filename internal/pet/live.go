package pet

import (
	"sync"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type Activation struct {
	ID      string
	Version int64
}
type liveView struct {
	items []session.AttentionItem
	known bool
}

// Companion consumes the workbench's existing snapshots. It never queries a
// backend, resolves a request or sends terminal input. Native wakes coalesce.
type Companion struct {
	mu       sync.Mutex
	view     liveView
	wake     func()
	queued   bool
	activate func(Activation)
	last     string
}

func NewCompanion(activate func(Activation)) *Companion { return &Companion{activate: activate} }

func (c *Companion) Publish(items []session.AttentionItem, known bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.view = liveView{append([]session.AttentionItem(nil), items[:min(len(items), session.MaxAttentionItems)]...), known}
	if !c.queued && c.wake != nil {
		c.queued = true
		c.wake() // internal nonblocking PostMessage; serialized with detach
	}
}

func (c *Companion) snapshot() liveView {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queued = false
	return liveView{append([]session.AttentionItem(nil), c.view.items...), c.view.known}
}

func (c *Companion) setWake(wake func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.wake = wake
	c.queued = false
	if wake != nil {
		wake()
	}
}

func (c *Companion) Activate() bool {
	c.mu.Lock()
	if !c.view.known || c.activate == nil {
		c.mu.Unlock()
		return false
	}
	var pending []session.AttentionItem
	for _, item := range c.view.items {
		if item.State == "pending" && item.TerminalAvailable {
			pending = append(pending, item)
		}
	}
	if len(pending) == 0 {
		c.mu.Unlock()
		return false
	}
	index := 0
	for i, item := range pending {
		if item.ID == c.last {
			index = (i + 1) % len(pending)
		}
	}
	for n := range pending {
		if i := (index + n) % len(pending); pending[i].ViewedAt == nil {
			index = i
			break
		}
	}
	item := pending[index]
	c.last = item.ID
	callback := c.activate
	c.mu.Unlock()
	callback(Activation{item.ID, item.Version})
	return true
}

func project(items []session.AttentionItem, known bool) Model {
	m := Model{Kind: Idle}
	for _, item := range items {
		if item.State != "pending" {
			continue
		}
		m.Count++
		if !item.TerminalAvailable {
			known = false
		}
		if item.Kind == "approval" {
			m.Kind = Approval
		} else if item.Kind == "question" && m.Kind != Approval {
			m.Kind = Question
		}
	}
	if !known {
		m.Kind = Unknown
	}
	return m
}
