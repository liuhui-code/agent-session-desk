package appearance

import (
	"context"
	"testing"
)

func TestTerminalThemeDeclarationRequiresCompleteBoundedIndex(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{{"15;0", "dark"}, {"0;7", "dark"}, {"0;8", "light"}, {"0;15", "light"}, {"0;1;15", "light"}, {"0;16", ""}, {"0;-1", ""}, {"0;15tail", ""}, {"0; 8", ""}, {"0;", ""}, {"15", ""}, {"0;999999999999999999999", ""}} {
		t.Run(tc.raw, func(t *testing.T) {
			v, ok := terminalTheme(tc.raw)
			if v != tc.want || ok != (tc.want != "") {
				t.Fatalf("declaration guessed: %q %v", v, ok)
			}
		})
	}
}

func TestObserverKeepsOnlyLatestPendingObservation(t *testing.T) {
	s := &stream{events: make(chan Observation, 1), done: make(chan struct{}), stop: func() {}}
	for range 100 {
		s.publish(Observation{Theme: "dark", Source: "terminal"})
	}
	s.publish(Observation{Theme: "light", Source: "terminal"})
	if len(s.events) != 1 {
		t.Fatal("unbounded observations")
	}
	v, e := s.Next(context.Background())
	if e != nil || v.Theme != "light" {
		t.Fatalf("latest observation lost: %+v %v", v, e)
	}
}
