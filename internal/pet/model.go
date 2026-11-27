// Package pet displays existing workbench attention items in a native Windows
// companion. The separate pet-demo command remains a visual diagnostic.
package pet

import "fmt"

type Kind string

const (
	Idle     Kind = "idle"
	Approval Kind = "approval"
	Question Kind = "question"
	Unknown  Kind = "unknown"
)

type Model struct {
	Kind  Kind
	Count int
	blink bool
	demo  bool
}

func New(kind Kind, count int) (Model, error) {
	if kind != Idle && kind != Approval && kind != Question && kind != Unknown {
		return Model{}, fmt.Errorf("demo state must be idle, approval, question or unknown")
	}
	if count < 1 || count > 99 {
		return Model{}, fmt.Errorf("demo count must be 1-99")
	}
	return Model{Kind: kind, Count: count, demo: true}, nil
}
func (m Model) Label() string {
	if m.demo {
		return "DEMO"
	}
	if m.Kind == Unknown {
		return "LINK"
	}
	if m.Count > 0 {
		return "WAIT"
	}
	return "FREE"
}
func (m Model) Cell() int {
	switch m.Kind {
	case Approval:
		return 2
	case Question:
		return 3
	case Idle:
		if m.blink {
			return 1
		}
	}
	return 0
}
func (m *Model) Tick() { m.blink = !m.blink }
func (m *Model) Next() {
	if !m.demo {
		return
	}
	m.blink = false
	switch m.Kind {
	case Idle:
		m.Kind = Approval
	case Approval:
		m.Kind = Question
	case Question:
		m.Kind = Unknown
	default:
		m.Kind = Idle
	}
}
