package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func TestQuickLaunchUsesDefaultThenLastAndPreservesArguments(t *testing.T) {
	dir := t.TempDir()
	defaults := filepath.Join(dir, "launch.toml")
	last := filepath.Join(dir, "last.toml")
	first := session.Launch{Executable: "shell.exe", Args: []string{"-NoExit"}, Dir: dir}
	if err := EnsureDefaultLaunch(defaults, first); err != nil {
		t.Fatal(err)
	}
	got, err := LoadQuickLaunch(defaults, last)
	if err != nil || !reflect.DeepEqual(got.Args, first.Args) {
		t.Fatalf("default: %#v %v", got, err)
	}
	next := session.Launch{Name: "Review", Executable: "agent.exe", Args: []string{"a b", "", "中文"}, Dir: dir, Env: []string{"MODE=a=b"}}
	if err := SaveLastLaunch(last, next); err != nil {
		t.Fatal(err)
	}
	got, err = LoadQuickLaunch(defaults, last)
	if err != nil || !reflect.DeepEqual(got, next) {
		t.Fatalf("last changed: %#v %v", got, err)
	}
	before, _ := os.ReadFile(defaults)
	_ = EnsureDefaultLaunch(defaults, next)
	after, _ := os.ReadFile(defaults)
	if string(before) != string(after) {
		t.Fatal("authored default overwritten")
	}
	if err := os.WriteFile(last, []byte("invalid=1"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadQuickLaunch(defaults, last); err == nil {
		t.Fatal("corrupt last silently replaced with default")
	}
	if err := SaveLastLaunch(last, next); err == nil {
		t.Fatal("corrupt cache overwritten")
	}
}
