package session

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestResolveLaunchUsesTargetPATHAndPreservesBytes(t *testing.T) {
	dir := t.TempDir()
	binDir := filepath.Join(dir, "工具 path")
	if err := os.Mkdir(binDir, 0700); err != nil {
		t.Fatal(err)
	}
	name := "agent fixture"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(binDir, name)
	if err := os.WriteFile(path, []byte("fixture"), 0700); err != nil {
		t.Fatal(err)
	}
	args := []string{"", "a b", "中文", `a"b`, `tail\`, "$(literal);&|", "--flag=value"}
	launch := Launch{Executable: name, Dir: dir, Args: args, Env: []string{"PATH=工具 path", "ASD_TEST=one=two"}}
	oldPath := os.Getenv("PATH")
	resolved, err := ResolveLaunch(launch)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Executable != path || resolved.Dir != dir || !reflect.DeepEqual(resolved.Args, args) {
		t.Fatalf("launch changed path or argv: %#v", resolved)
	}
	if os.Getenv("PATH") != oldPath {
		t.Fatal("resolving launch changed process PATH")
	}
	resolved.Args[0] = "changed"
	if launch.Args[0] != "" {
		t.Fatal("resolved launch aliases caller's argument slice")
	}
}

func TestResolveLaunchRejectsBadDirectoryAndMissingExecutable(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "not-directory")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	for _, l := range []Launch{
		{Executable: "missing-agent", Dir: dir, Env: []string{"PATH="}},
		{Executable: "missing-agent", Dir: filepath.Join(dir, "absent")},
		{Executable: "missing-agent", Dir: file},
		{Executable: "  ", Dir: dir},
	} {
		if _, err := ResolveLaunch(l); err == nil {
			t.Fatalf("invalid launch accepted: %#v", l)
		}
	}
}

func TestResolveRelativeExecutableAgainstTargetDirectory(t *testing.T) {
	dir := t.TempDir()
	name := "relative-agent"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := os.WriteFile(filepath.Join(dir, name), nil, 0700); err != nil {
		t.Fatal(err)
	}
	l, err := ResolveLaunch(Launch{Executable: "." + string(os.PathSeparator) + name, Dir: dir})
	if err != nil || l.Executable != filepath.Join(dir, name) {
		t.Fatalf("relative command: %#v, %v", l, err)
	}
}

func TestLaunchRejectsNULAndInvalidEnvironmentWithoutEchoingSecrets(t *testing.T) {
	for _, l := range []Launch{
		{Executable: "exe\x00"}, {Executable: "exe", Dir: "dir\x00"},
		{Executable: "exe", Args: []string{"secret\x00value"}},
		{Executable: "exe", Env: []string{"=secret"}},
		{Executable: "exe", Env: []string{"secret"}},
		{Executable: "exe", Env: []string{"KEY=secret\x00value"}},
	} {
		err := l.Validate()
		if err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatalf("invalid transport accepted or secret echoed: %v", err)
		}
	}
	if err := (Launch{Executable: "exe", Args: []string{""}, Env: []string{"KEY=", "KEY=a=b"}}).Validate(); err != nil {
		t.Fatalf("valid empty/equal values rejected: %v", err)
	}
}

func TestWindowsLaunchEnvironmentAndScriptBoundary(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows filesystem/PATH semantics require Windows execution")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "agent.EXE"), nil, 0700); err != nil {
		t.Fatal(err)
	}
	l, err := ResolveLaunch(Launch{Executable: "agent", Dir: dir, Env: []string{"PATH=invalid", "Path=" + dir, "Pathext=.EXE"}})
	if err != nil || !strings.EqualFold(l.Executable, filepath.Join(dir, "agent.EXE")) {
		t.Fatalf("case-insensitive Windows environment: %#v, %v", l, err)
	}
	if len(l.Env) != 2 || l.Env[0] != "Path="+dir {
		t.Fatalf("backend could see different PATH than the resolver: %#v", l.Env)
	}
	for _, ext := range []string{".cmd", ".bat", ".ps1"} {
		path := filepath.Join(dir, "script"+ext)
		if err := os.WriteFile(path, nil, 0700); err != nil {
			t.Fatal(err)
		}
		if _, err := ResolveLaunch(Launch{Executable: path, Dir: dir}); err == nil || !strings.Contains(err.Error(), "explicit shell") {
			t.Fatalf("script silently accepted: %s, %v", ext, err)
		}
	}
}

func TestDuplicateEnvironmentOverrideUsesLastValue(t *testing.T) {
	got := normalizeOverrides([]string{"KEY=first", "OTHER=preserved", "KEY=last"})
	want := []string{"OTHER=preserved", "KEY=last"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("duplicate overrides: %#v, want %#v", got, want)
	}
}
