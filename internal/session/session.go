// Package session defines the small, backend-independent M0 session model.
// Backend availability is deliberately separate from agent work or readiness.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type TerminalState string

const (
	TerminalAvailable TerminalState = "available"
	TerminalExited    TerminalState = "exited"
	TerminalUnknown   TerminalState = "unknown"
)

// Session identifies a logical session, one invocation, and its terminal.
// Command is display-only (the executable), never a reconstructed shell command.
// An available terminal does not establish that an agent is busy or ready.
type Session struct {
	ID            string        `json:"session_id"`
	RunID         string        `json:"run_id"`
	BackendID     string        `json:"backend_id"`
	Title         string        `json:"title"`
	TitleRevision int64         `json:"title_revision,omitempty"`
	Command       string        `json:"command"`
	TerminalState TerminalState `json:"terminal_state"`
	PaneExit      *PaneExit     `json:"pane_exit,omitempty"`
}

// PaneExit is an observation of the product-associated pane, not an agent
// lifecycle or task result. The observation time is not the process exit time.
// Both current backends leave ExitCode nil until a reliable contract is proved.
type PaneExit struct {
	PaneID     string    `json:"pane_id"`
	Source     string    `json:"source"`
	ObservedAt time.Time `json:"observed_at"`
	ExitCode   *int      `json:"exit_code"`
}

// Preview is a bounded screen snapshot, not a conversation or a complete log.
type Preview struct {
	Text      string `json:"text"`
	Truncated bool   `json:"truncated"`
}

// Launch keeps argument boundaries intact. Shell parsing only occurs when the
// caller explicitly chooses a shell as Executable and supplies its arguments.
type Launch struct {
	Name       string
	Executable string
	Args       []string
	Dir        string
	Env        []string
}

func NewID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate session identity: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}

// Validate checks the transport contract without interpreting user arguments.
// Values of environment variables are never included in validation errors.
func (l Launch) Validate() error {
	if strings.TrimSpace(l.Executable) == "" {
		return fmt.Errorf("executable is required")
	}
	for _, item := range []struct{ name, value string }{
		{"name", l.Name}, {"executable", l.Executable}, {"working directory", l.Dir},
	} {
		if strings.ContainsRune(item.value, '\x00') {
			return fmt.Errorf("%s contains a NUL byte", item.name)
		}
	}
	for i, arg := range l.Args {
		if strings.ContainsRune(arg, '\x00') {
			return fmt.Errorf("argument %d contains a NUL byte", i+1)
		}
	}
	for i, entry := range l.Env {
		key, _, ok := strings.Cut(entry, "=")
		if !ok || key == "" || strings.ContainsAny(key, "\x00\r\n") || strings.ContainsRune(entry, '\x00') {
			return fmt.Errorf("environment entry %d must be KEY=value without NUL bytes", i+1)
		}
	}
	return nil
}
