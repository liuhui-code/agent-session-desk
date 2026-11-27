package ui

import "testing"
import "context"
import "fmt"
import "time"

func TestActivityBaselineAndSilence(t *testing.T) {
	a := activityEntry{}
	now := time.Unix(100, 0)
	a.observe([32]byte{1}, now)
	if a.moving(now) {
		t.Fatal("baseline claimed output")
	}
	a.observe([32]byte{2}, now.Add(time.Second))
	if !a.moving(now.Add(time.Second)) {
		t.Fatal("change not active")
	}
	a.observe([32]byte{2}, now.Add(4*time.Second))
	if a.moving(now.Add(4 * time.Second)) {
		t.Fatal("no silence")
	}
}

func TestActivityUnknownAndRunReplacement(t *testing.T) {
	m := New(context.Background(), nil)
	defer m.Close()
	s := sampleSessions()[0]
	now := time.Now()
	m.activity.entries = map[string]activityEntry{activityKey(s): {seen: now}}
	if m.activityMark(s) == "" {
		t.Fatal("quiet absent")
	}
	s.RunID = "other"
	if m.activityMark(s) != "" {
		t.Fatal("inherited another Run")
	}
	s = sampleSessions()[0]
	m.listErr = fmt.Errorf("offline")
	if m.activityMark(s) != "" {
		t.Fatal("offline claimed idle")
	}
}
