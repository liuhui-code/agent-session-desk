// Package diagnostics stores bounded, allowlisted observations, never terminal
// text, raw errors, commands, paths, environment or authentication material.
package diagnostics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

const MaxEvents = 128
const MaxExportBytes = 256 * 1024

type Operation uint8

const (
	Discovery Operation = iota + 1
	Connect
	Stream
	Input
	Resize
)

func (op Operation) name() string {
	switch op {
	case Discovery:
		return "discovery"
	case Connect:
		return "connect"
	case Stream:
		return "stream"
	case Input:
		return "input_result"
	case Resize:
		return "resize"
	}
	return ""
}

type Event struct {
	At        time.Time `json:"at"`
	Operation string    `json:"operation"`
	Target    string    `json:"target,omitempty"`
	ElapsedMS int64     `json:"observed_elapsed_ms"`
	Cause     string    `json:"cause"`
}
type Runtime struct {
	ProductVersion string `json:"product_version"`
	Revision       string `json:"revision"`
	Dirty          bool   `json:"dirty"`
	Go             string `json:"go"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
	Backend        string `json:"backend"`
	BackendVersion string `json:"backend_version"`
}

func CurrentRuntime(backend, version string) Runtime {
	r := Runtime{Go: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, Backend: backend, BackendVersion: version}
	if b, ok := debug.ReadBuildInfo(); ok {
		for _, s := range b.Settings {
			switch s.Key {
			case "vcs.revision":
				r.Revision = s.Value
			case "vcs.modified":
				r.Dirty = s.Value == "true"
			}
		}
	}
	return r
}
func Reference(identity string) string {
	if identity == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(identity))
	return "r-" + hex.EncodeToString(sum[:8])
}

// Journal is confined to the UI update goroutine; Snapshot returns owned data.
type Journal struct {
	events      [MaxEvents]Event
	next, count int
}

func (j *Journal) Record(op Operation, identity string, elapsed time.Duration, err error) {
	if op.name() == "" {
		return
	}
	j.events[j.next] = Event{time.Now().UTC(), op.name(), Reference(identity), max(0, elapsed.Milliseconds()), Cause(err)}
	j.next = (j.next + 1) % MaxEvents
	j.count = min(j.count+1, MaxEvents)
}
func Cause(err error) string {
	if err == nil {
		return "ok"
	}
	if errors.Is(err, ErrSlotsFull) {
		return "export_full"
	}
	if errors.Is(err, fs.ErrPermission) {
		return "permission_denied"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	if errors.Is(err, context.Canceled) {
		return "cancelled"
	}
	text := err.Error()
	text = strings.ToLower(text[:min(len(text), 4096)])

	for _, v := range []struct{ prefix, cause string }{
		{"terminal handshake", "handshake_failed"}, {"terminal authentication", "handshake_failed"},
		{"invalid product pane identity", "identity_unverified"}, {"terminal session/pane changed", "identity_unverified"}, {"original pane", "identity_unverified"}, {"original session", "identity_unverified"},
		{"target terminal is not interactive", "terminal_not_interactive"}, {"terminal connection closed", "closed"}, {"terminal disconnected", "disconnected"},
	} {
		if strings.HasPrefix(text, v.prefix) {
			return v.cause
		}
	}

	return "operation_failed"
}

type Report struct {
	Version   int       `json:"schema_version"`
	Generated time.Time `json:"generated_at"`
	Runtime   Runtime   `json:"runtime"`
	Target    string    `json:"target,omitempty"`
	Terminal  string    `json:"terminal"`
	Client    string    `json:"client"`
	Events    []Event   `json:"events"`
}

func (j *Journal) Snapshot(r Runtime, identity, terminal, client string) Report {
	events := make([]Event, 0, j.count)
	for i := 0; i < j.count; i++ {
		events = append(events, j.events[(j.next-j.count+i+MaxEvents)%MaxEvents])
	}
	return Report{1, time.Now().UTC(), r, Reference(identity), terminal, client, events}
}

var safeVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
var safeProductVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+([-+][a-zA-Z0-9.-]+)?$`)
var safeRevision = regexp.MustCompile(`^[a-f0-9]{40}$`)
var safeReference = regexp.MustCompile(`^r-[a-f0-9]{16}$`)
var safeGo = regexp.MustCompile(`^go[0-9]+\.[0-9]+(\.[0-9]+)?$`)

func (r Report) Encode() ([]byte, error) {
	// Values remain allowlisted even if a caller supplies a malformed runtime.
	if len(r.Runtime.ProductVersion) > 96 || !safeProductVersion.MatchString(r.Runtime.ProductVersion) {
		r.Runtime.ProductVersion = "unknown"
	}
	if !safeRevision.MatchString(r.Runtime.Revision) {
		r.Runtime.Revision = "unknown"
	}
	if !safeGo.MatchString(r.Runtime.Go) {
		r.Runtime.Go = "unknown"
	}
	if !safeVersion.MatchString(r.Runtime.BackendVersion) {
		r.Runtime.BackendVersion = "unknown"
	}
	if r.Runtime.Backend != "psmux" && r.Runtime.Backend != "tmux" {
		r.Runtime.Backend = "unknown"
	}
	if r.Runtime.OS != "windows" && r.Runtime.OS != "darwin" && r.Runtime.OS != "linux" {
		r.Runtime.OS = "unknown"
	}
	if r.Runtime.Arch != "amd64" && r.Runtime.Arch != "arm64" {
		r.Runtime.Arch = "unknown"
	}
	if r.Terminal != "available" && r.Terminal != "exited" {
		r.Terminal = "unknown"
	}
	switch r.Client {
	case "connected", "connecting", "disconnected", "not_open":
	default:
		r.Client = "unknown"
	}

	if !safeReference.MatchString(r.Target) {
		r.Target = ""
	}
	events := make([]Event, 0, min(len(r.Events), MaxEvents))
	for _, event := range r.Events[max(0, len(r.Events)-MaxEvents):] {
		switch event.Operation {
		case "discovery", "connect", "stream", "input_result", "resize":
		default:
			continue
		}
		if !safeReference.MatchString(event.Target) {
			event.Target = ""
		}
		switch event.Cause {
		case "ok", "timeout", "cancelled", "permission_denied", "export_full", "handshake_failed", "identity_unverified", "terminal_not_interactive", "closed", "disconnected", "operation_failed":
		default:
			event.Cause = "operation_failed"
		}
		event.ElapsedMS = max(0, min(event.ElapsedMS, int64((24*time.Hour)/time.Millisecond)))
		events = append(events, event)
	}
	r.Version = 1
	r.Events = events
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil || len(data) > MaxExportBytes {
		return nil, errors.New("diagnostic report exceeds limit or cannot be encoded")
	}
	return data, nil
}
