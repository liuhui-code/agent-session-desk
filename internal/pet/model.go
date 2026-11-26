// Package pet is an opt-in native Windows visual demo, not an agent observer.
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
}

func New(kind Kind, count int) (Model, error) {
	if kind != Idle && kind != Approval && kind != Question && kind != Unknown {
		return Model{}, fmt.Errorf("demo state must be idle, approval, question or unknown")
	}
	if count < 1 || count > 99 {
		return Model{}, fmt.Errorf("demo count must be 1-99")
	}
	return Model{Kind: kind, Count: count}, nil
}
func (m Model) Label() string { return "DEMO" }
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
