// Portions adapted from agent-manager internal/tmux/platform_windows.go
// at cc88464dfb6de59933ad3ff7246d987472e14beb.
// Copyright 2026 Yoan Wai and the agent-manager contributors.
// Licensed under the Apache License, Version 2.0.
// See third_party/agent-manager/LICENSE and third_party/agent-manager/NOTICE.
// Modified for Agent Session Desk: strict stable-version validation; isolated
// environment and routing; direct executable/argv without a PowerShell wrapper.

package backend

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const psmuxBaseline = "3.4.0"

// Stable psmux 3.4.0's src/types.rs::build_version_string appends build
// provenance. Accept its documented forms, not arbitrary trailing words. This
// parses the reported release; the native runner separately pins binary hashes.
var psmuxVersionLine = regexp.MustCompile(`^psmux (\S+)(?: \((?:[0-9a-f]{7,40}(?: [0-9]{4}-[0-9]{2}-[0-9]{2})?(?:, dirty)?|unknown commit)\))?$`)

// psmux prints a tmux-compatibility version AND its own version. Do not accept
// the compatibility line as evidence of the psmux release being exercised.
func parsePsmuxVersion(out string) (string, error) {
	found := "unknown"
	for _, line := range strings.FieldsFunc(out, func(r rune) bool { return r == '\n' || r == '\r' }) {
		if match := psmuxVersionLine.FindStringSubmatch(strings.TrimSpace(line)); match != nil {
			found = match[1]
			break
		}
	}
	major, minor, ok := splitVersion(found)
	if ok && major == 3 && minor == 4 && found == psmuxBaseline {
		return found, nil
	}
	return found, fmt.Errorf("psmux version %q is outside this M0 contract; use stable psmux %s", found, psmuxBaseline)
}

func splitVersion(s string) (int, int, bool) {
	dot := strings.Index(s, ".")
	if dot <= 0 {
		return 0, 0, false
	}
	major, err := strconv.Atoi(s[:dot])
	if err != nil || major < 0 {
		return 0, 0, false
	}
	rest := s[dot+1:]
	dot = strings.Index(rest, ".")
	if dot < 0 {
		dot = len(rest)
	}
	minor, err := strconv.Atoi(rest[:dot])
	if err != nil || minor < 0 {
		return 0, 0, false
	}
	return major, minor, true
}

// A psmux namespace holds one server per session. Session-wide commands carry
// an explicit route; paneArgs uses one session-qualified pane target instead.
func sessionRoute(name string) []string {
	if name == "" {
		return nil
	}
	return []string{"-t", name}
}

// For the subset accepted by validateTransport, -- selects psmux's direct
// executable/argv path. Its global normalizers still rewrite some forms even
// after --; rejecting those forms is mandatory before calling this builder.
// This migration does not inherit upstream's persistent shell wrapper:
// its continued existence would not prove the requested agent was still alive.
func psmuxStartArgs(name string, launch session.Launch) []string {
	args := []string{"new-session", "-d", "-s", name, "-c", launch.Dir, "-x", "120", "-y", "32"}
	for _, value := range launch.Env {
		args = append(args, "-e", value)
	}
	args = append(args, "--", launch.Executable)
	return append(args, launch.Args...)
}

// Never change os.Environ globally. Drop inherited backend routing first: an
// invocation inside somebody else's pane must still address our private server.
func commandEnv(inherited []string, kind, dataDir, config string) []string {
	env := make([]string, 0, len(inherited)+5)
	for _, value := range inherited {
		key, _, _ := strings.Cut(value, "=")
		upper := strings.ToUpper(key)
		if upper == "TMUX" || upper == "TMUX_PANE" || upper == "TMUX_TMPDIR" || strings.HasPrefix(upper, "PSMUX_") {
			continue
		}
		env = append(env, value)
	}
	if kind == "psmux" {
		env = append(env, "PSMUX_NO_WARM=1", "PSMUX_ALLOW_NESTING=1", "PSMUX_DATA_DIR="+dataDir, "PSMUX_CONFIG_FILE="+config)
	} else {
		env = append(env, "TMUX_TMPDIR="+dataDir)
	}
	return env
}
