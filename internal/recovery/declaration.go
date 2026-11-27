// Package recovery defines bounded native-conversation recovery declarations.
// It stores launch descriptions, never conversation bodies or model context.
package recovery

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const MaxDeclarationBytes = 16 * 1024
const MaxResultBytes = 4 * 1024

// Declaration.RunID is the source Run, not the next recovery invocation.
type Declaration struct {
	Protocol       string `json:"protocol"`
	SchemaVersion  int    `json:"schema_version"`
	SessionID      string `json:"session_id"`
	RunID          string `json:"run_id"`
	AgentID        string `json:"agent_id"`
	AgentVersion   string `json:"agent_version,omitempty"`
	ConversationID string `json:"conversation_id"`
	Resume         Resume `json:"resume"`
}

type Resume struct {
	Executable string            `json:"executable"`
	Args       []string          `json:"args"`
	Cwd        string            `json:"cwd"`
	Env        map[string]string `json:"env,omitempty"`
}

type Result struct {
	Protocol       string `json:"protocol"`
	SchemaVersion  int    `json:"schema_version"`
	SessionID      string `json:"session_id"`
	RunID          string `json:"run_id"`
	AttemptID      string `json:"attempt_id"`
	ConversationID string `json:"conversation_id"`
	Status         string `json:"status"`
	Message        string `json:"message,omitempty"`
}

var agentPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
var envPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,63}$`)
var drivePattern = regexp.MustCompile(`^[A-Za-z]:[\\/]`)

func text(value string, min, max int) bool {
	n := utf8.RuneCountInString(value)
	if !utf8.ValidString(value) || n < min || n > max {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func (d Declaration) Validate() error {
	if d.Protocol != "asd.agent-recovery" || d.SchemaVersion != 1 || !session.ValidID(d.SessionID) || !session.ValidID(d.RunID) || !text(d.AgentID, 1, 64) || !agentPattern.MatchString(d.AgentID) || !text(d.AgentVersion, 0, 128) || !text(d.ConversationID, 1, 512) {
		return fmt.Errorf("invalid recovery declaration identity/version")
	}
	r := d.Resume
	if !text(r.Executable, 3, 4096) || !drivePattern.MatchString(r.Executable) || !text(r.Cwd, 3, 4096) || !drivePattern.MatchString(r.Cwd) || r.Args == nil || len(r.Args) > 128 || len(r.Env) > 64 {
		return fmt.Errorf("invalid recovery launch shape or Windows drive path")
	}
	lower := strings.ToLower(r.Executable)
	for _, ext := range []string{".cmd", ".bat", ".ps1"} {
		if strings.HasSuffix(lower, ext) {
			return fmt.Errorf("recovery scripts require an explicit interpreter executable")
		}
	}
	for _, arg := range r.Args {
		if !text(arg, 0, 4096) {
			return fmt.Errorf("invalid recovery argument")
		}
	}
	for k, v := range r.Env {
		if !text(k, 1, 64) || !envPattern.MatchString(k) || !text(v, 0, 4096) {
			return fmt.Errorf("invalid recovery environment")
		}
	}
	b, err := json.Marshal(d)
	if err != nil || len(b) > MaxDeclarationBytes {
		return fmt.Errorf("recovery declaration exceeds byte limit")
	}
	return nil
}

func (d Declaration) Launch(name string) session.Launch {
	env := make([]string, 0, len(d.Resume.Env))
	for k, v := range d.Resume.Env {
		env = append(env, k+"="+v)
	}
	sort.Strings(env)
	return session.Launch{Name: name, Executable: d.Resume.Executable, Args: append([]string{}, d.Resume.Args...), Dir: d.Resume.Cwd, Env: env}
}

func (r Result) Validate() error {
	if r.Protocol != "asd.agent-recovery-result" || r.SchemaVersion != 1 || !session.ValidID(r.SessionID) || !session.ValidID(r.RunID) || !session.ValidID(r.AttemptID) || !text(r.ConversationID, 1, 512) || !text(r.Message, 0, 512) {
		return fmt.Errorf("invalid recovery result identity/version")
	}
	if r.Status != "restored" && r.Status != "not_found" && r.Status != "failed" {
		return fmt.Errorf("invalid recovery result status")
	}
	b, err := json.Marshal(r)
	if err != nil || len(b) > MaxResultBytes {
		return fmt.Errorf("recovery result exceeds byte limit")
	}
	return nil
}
