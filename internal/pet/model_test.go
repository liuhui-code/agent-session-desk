package pet

import "testing"

func TestDemoStatesCannotBecomeRealApproval(t *testing.T) {
	for _, kind := range []Kind{Idle, Approval, Question, Unknown} {
		m, err := New(kind, 2)
		if err != nil {
			t.Fatal(err)
		}
		if m.Label() != "DEMO" {
			t.Fatal("synthetic state lacks demo label")
		}
		if m.Kind != kind {
			t.Fatal("state changed")
		}
	}
	for _, kind := range []Kind{"pending", "allow", ""} {
		if _, err := New(kind, 1); err == nil {
			t.Fatal("unsupported state accepted")
		}
	}
	if _, err := New(Approval, 100); err == nil {
		t.Fatal("unbounded count")
	}
}

func TestDemoAnimationIsBoundedAndClickOnlyCyclesPreview(t *testing.T) {
	m, _ := New(Idle, 1)
	if m.Cell() != 0 {
		t.Fatal("idle cell")
	}
	m.Tick()
	if m.Cell() != 1 {
		t.Fatal("blink cell")
	}
	m.Next()
	if m.Kind != Approval || m.Cell() != 2 {
		t.Fatal("approval preview")
	}
	m.Next()
	if m.Kind != Question || m.Cell() != 3 {
		t.Fatal("question preview")
	}
	m.Next()
	if m.Kind != Unknown {
		t.Fatal("unknown preview")
	}
	m.Next()
	if m.Kind != Idle || m.Cell() != 0 {
		t.Fatal("cycle changed identity or exceeded states")
	}
}
