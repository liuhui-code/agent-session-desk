// Package backend implements the experimental M0 multiplexer command contract.
// Backend user options are live discovery hints, not the later SQLite history.
package backend

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const (
	metadataOption   = "@asd_m0_record"
	maxMetadataBytes = 4096
	maxControlBytes  = 128 * 1024
	MaxPreviewBytes  = 64 * 1024
	maxPreviewLines  = 2000
	commandTimeout   = 35 * time.Second
)

var (
	namePattern      = regexp.MustCompile(`^asd_([0-9a-f]{32})_([0-9a-f]{32})$`)
	panePattern      = regexp.MustCompile(`^%[0-9]+$`)
	namespacePattern = regexp.MustCompile(`^asd-[a-zA-Z0-9_-]{1,48}$`)
)

type Options struct {
	Binary         string
	RuntimeDir     string
	Namespace      string
	TerminalBundle string
}

type Info struct {
	Name         string
	Version      string
	Experimental bool
}

type Backend struct {
	bin, kind, dataDir, config, namespace string
	bundle                                string
	discovery                             discoveryHints
}

// New prepares only product-owned paths. It neither starts a server nor changes
// the user's global multiplexer configuration. All operations remain experimental.
func New(opts Options) (*Backend, error) {
	kind, dir, err := Location(opts)
	if err != nil {
		return nil, err
	}
	bundle, err := prepareTerminalBundle(opts, kind)
	if err != nil {
		return nil, err
	}
	if bundle != "" {
		opts.Binary = filepath.Join(bundle, "psmux.exe")
	}
	if opts.Binary == "" {
		opts.Binary = kind
	}
	bin, err := exec.LookPath(opts.Binary)
	if err != nil {
		return nil, fmt.Errorf("find %s backend: %w", kind, err)
	}
	bin, err = filepath.Abs(bin)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("prepare backend directory: %w", err)
	}
	config := filepath.Join(dir, kind+"-m0.conf")
	body := "# Generated M0 contract; applies only to the product namespace.\nset -g status off\nset -g history-limit 2000\nset-window-option -g remain-on-exit on\n"
	if kind == "psmux" {
		body += "set -g warm off\n"
	}
	if err := ensureConfig(config, body); err != nil {
		return nil, err
	}
	return &Backend{bin: bin, kind: kind, dataDir: dir, config: config, namespace: filepath.Base(dir), bundle: bundle}, nil
}

func ensureConfig(path, body string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		current, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read product backend config: %w", readErr)
		}
		if string(current) != body {
			return errors.New("product backend config differs from this M0 contract; choose a fresh --runtime-dir")
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("create product backend config: %w", err)
	}
	_, writeErr := file.WriteString(body)
	closeErr := file.Close()
	return errors.Join(writeErr, closeErr)
}

func (b *Backend) Probe(ctx context.Context) (Info, error) {
	out, err := b.run(ctx, maxControlBytes, "-V")
	info := Info{Name: b.kind, Experimental: true}
	if err != nil {
		return info, err
	}
	if b.kind == "psmux" {
		info.Version, err = parsePsmuxVersion(out.text)
		return info, err
	}
	fields := strings.Fields(out.text)
	if len(fields) != 2 || fields[0] != "tmux" {
		return info, errors.New("backend did not report a tmux version")
	}
	info.Version = fields[1]
	if info.Version != "3.8" {
		return info, fmt.Errorf("tmux %s is outside this M0 development contract; use tmux 3.8", info.Version)
	}
	return info, nil
}

type liveRecord struct {
	Schema    int    `json:"schema"`
	SessionID string `json:"session_id"`
	RunID     string `json:"run_id"`
	Title     string `json:"title"`
	Command   string `json:"command"`
	PaneID    string `json:"pane_id"`
}

func encodeRecord(r liveRecord) (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	value := base64.RawStdEncoding.EncodeToString(data)
	if len(value) > maxMetadataBytes {
		return "", errors.New("M0 live discovery metadata exceeds 4096 bytes; shorten the name or executable path")
	}
	return value, nil
}

func decodeRecord(name, value string) (liveRecord, error) {
	var record liveRecord
	value = strings.TrimSpace(value)
	ids := namePattern.FindStringSubmatch(name)
	if ids == nil || len(value) > maxMetadataBytes {
		return record, errors.New("invalid M0 backend identity or metadata size")
	}
	data, err := base64.RawStdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return record, errors.New("invalid M0 backend metadata encoding")
	}
	if err = json.Unmarshal(data, &record); err != nil {
		return record, errors.New("invalid M0 backend metadata")
	}
	if record.Schema != 1 || record.SessionID != ids[1] || record.RunID != ids[2] || !panePattern.MatchString(record.PaneID) {
		return record, errors.New("M0 backend Session/Run/pane identity mismatch")
	}
	return record, nil
}

func (b *Backend) metadataArgs(name string, write string) []string {
	args := []string{}
	if b.kind == "psmux" {
		args = append(args, sessionRoute(name)...)
	}
	if write != "" {
		args = append(args, "set-option")
	} else {
		args = append(args, "show-options")
	}
	if b.kind == "psmux" {
		args = append(args, "-g")
	} else {
		args = append(args, "-t", name)
	}
	if write == "" {
		args = append(args, "-q", "-v")
	}
	args = append(args, metadataOption)
	if write != "" {
		args = append(args, write)
	}
	return args
}

func (b *Backend) readRecord(ctx context.Context, name string) (liveRecord, error) {
	if !namePattern.MatchString(name) {
		return liveRecord{}, errors.New("invalid product backend target")
	}
	out, err := b.run(ctx, maxMetadataBytes+2, b.metadataArgs(name, "")...)
	if err != nil {
		return liveRecord{}, err
	}
	if out.truncated {
		return liveRecord{}, errors.New("backend metadata exceeded the M0 limit")
	}
	return decodeRecord(name, out.text)
}

func (b *Backend) validateTarget(ctx context.Context, s session.Session) (liveRecord, error) {
	r, err := b.readRecord(ctx, s.BackendID)
	if err != nil {
		return r, err
	}
	if r.SessionID != s.ID || r.RunID != s.RunID {
		return r, errors.New("target Session/Run changed; refresh before retrying")
	}
	return r, nil
}

func (b *Backend) StartRun(ctx context.Context, launch session.Launch, id, runID string) (session.Session, error) {
	if !session.ValidID(id) || !session.ValidID(runID) {
		return session.Session{}, errors.New("full reserved Session/Run identity required")
	}
	var result session.Session
	if err := launch.Validate(); err != nil {
		return result, err
	}
	if err := validateTransport(b.kind, launch); err != nil {
		return result, err
	}
	var err error
	launch, err = session.ResolveLaunch(launch)
	if err != nil {
		return result, err
	}
	if err := validateTransport(b.kind, launch); err != nil {
		return result, err
	}
	if _, err = b.Probe(ctx); err != nil {
		return result, err
	}
	name := "asd_" + id + "_" + runID
	result = session.Session{ID: id, RunID: runID, BackendID: name, Title: launch.Name, Command: launch.Executable, TerminalState: session.TerminalUnknown}
	if result.Title == "" {
		result.Title = filepath.Base(launch.Executable)
	}
	// Validate metadata size before launching anything. The actual pane ID is
	// discovered after creation rather than assuming a platform's base index.
	r := liveRecord{Schema: 1, SessionID: id, RunID: runID, Title: result.Title, Command: result.Command, PaneID: "%0"}
	if _, err = encodeRecord(r); err != nil {
		return result, err
	}
	args := psmuxStartArgs(name, launch)
	if b.kind == "tmux" {
		// tmux preserves argv only for a multi-argument shell-command. With one
		// program it invokes the default shell. /usr/bin/env -- keeps even a
		// zero-argument launch on the direct exec path without adding an agent arg.
		envPath, lookupErr := exec.LookPath("env")
		if lookupErr != nil {
			return result, lookupErr
		}
		args = []string{"new-session", "-d", "-s", name, "-c", launch.Dir, "-x", "120", "-y", "32"}
		for _, value := range launch.Env {
			args = append(args, "-e", value)
		}
		args = append(args, "--", envPath, "--", launch.Executable)
		args = append(args, launch.Args...)
	}
	_, err = b.run(ctx, maxControlBytes, args...)
	if err != nil {
		return result, b.failedStart(name, err)
	}
	if err := b.verifyTerminalBundle(ctx, name); err != nil {
		return result, b.failedStart(name, err)
	}
	out, err := b.run(ctx, maxControlBytes, "list-panes", "-t", name, "-F", "#{pane_id}")
	if err != nil {
		return result, b.failedStart(name, err)
	}
	panes := strings.Fields(out.text)
	if len(panes) != 1 || !panePattern.MatchString(panes[0]) {
		return result, b.failedStart(name, errors.New("new backend session did not expose exactly one initial pane"))
	}
	r.PaneID = panes[0]
	encoded, err := encodeRecord(r)
	if err != nil {
		return result, b.failedStart(name, err)
	}
	if _, err = b.run(ctx, maxControlBytes, b.metadataArgs(name, encoded)...); err != nil {
		return result, b.failedStart(name, err)
	}
	result.TerminalState, result.PaneExit = b.observePane(ctx, name, r.PaneID)
	return result, nil
}

func (b *Backend) failedStart(name string, cause error) error {
	// A creation timeout may already have created a server. Clean up only this
	// newly generated exact target; never kill an entire namespace or process name.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := b.run(ctx, maxControlBytes, "kill-session", "-t", name)
	if err != nil {
		return fmt.Errorf("launch failed (%w); cleanup of %s was not confirmed: %v", cause, name, err)
	}
	return fmt.Errorf("launch failed; newly created backend session removed: %w", cause)
}

func (b *Backend) paneArgs(name, pane, command string, args ...string) []string {
	target := pane
	if b.kind == "psmux" {
		// Stable 3.4.0 validates only the first -t target. Two targets (a
		// global session route and a bare pane) can skip pane validation and
		// capture another active pane if the original is gone. Use exactly
		// one qualified target; the dot marks %N as a pane, not a window.
		target = name + ":." + pane
	}
	return append([]string{command, "-t", target}, args...)
}

func (b *Backend) paneState(ctx context.Context, name, pane string) session.TerminalState {
	out, err := b.run(ctx, maxControlBytes, b.paneArgs(name, pane, "display-message", "-p", "#{pane_dead}")...)
	if err != nil {
		return session.TerminalUnknown
	}
	switch strings.TrimSpace(out.text) {
	case "0":
		return session.TerminalAvailable
	case "1":
		return session.TerminalExited
	default:
		return session.TerminalUnknown
	}
}

func (b *Backend) Preview(ctx context.Context, s session.Session, lines int) (session.Preview, error) {
	r, err := b.validateTarget(ctx, s)
	if err != nil {
		return session.Preview{}, err
	}
	if lines < 1 {
		lines = 1
	}
	if lines > maxPreviewLines {
		lines = maxPreviewLines
	}
	// Snapshot only. No resize or terminal input is allowed in this path.
	out, err := b.run(ctx, MaxPreviewBytes, b.paneArgs(s.BackendID, r.PaneID, "capture-pane", "-p", "-S", "-"+strconv.Itoa(lines))...)
	if err != nil {
		return session.Preview{}, err
	}
	// psmux 3.4.0's capture-by-id path falls back to the active pane if the
	// original disappears between the CLI's validation and the capture. IDs
	// are monotonic within that server; discard the text if the original can
	// no longer be verified. This is not an atomic backend snapshot guarantee.
	if b.kind == "psmux" && b.paneState(ctx, s.BackendID, r.PaneID) == session.TerminalUnknown {
		return session.Preview{}, errors.New("original pane could not be verified after capture; snapshot discarded")
	}
	text := out.text
	for len(text) > 0 && !utf8.RuneStart(text[0]) {
		text = text[1:]
	}
	text = strings.ToValidUTF8(text, "�")
	if len(text) > MaxPreviewBytes {
		text = text[len(text)-MaxPreviewBytes:]
		for len(text) > 0 && !utf8.RuneStart(text[0]) {
			text = text[1:]
		}
		out.truncated = true
	}
	return session.Preview{Text: text, Truncated: out.truncated}, nil
}

func (b *Backend) Attach(ctx context.Context, s session.Session) (*exec.Cmd, error) {
	if _, err := b.attachRecord(ctx, s); err != nil {
		return nil, err
	}
	// The caller owns handoff. Never detach another client or bind its lifetime
	// to the short metadata probe context.
	return b.command("attach-session", "-t", s.BackendID), nil
}

func (b *Backend) attachRecord(ctx context.Context, s session.Session) (liveRecord, error) {
	r, err := b.validateTarget(ctx, s)
	if err != nil {
		return r, err
	}
	if err = b.verifyTerminalBundle(ctx, s.BackendID); err != nil {
		return r, err
	}
	// Query the active pane and its liveness together. A different pane is a
	// rejection, never a reason to select it or to reinterpret this Run.
	out, err := b.run(ctx, maxControlBytes, "display-message", "-p", "-t", s.BackendID, "#{pane_id}|#{pane_dead}")
	if err != nil {
		return r, err
	}
	parts := strings.Split(strings.TrimSpace(out.text), "|")
	if out.truncated || len(parts) != 2 || parts[0] != r.PaneID {
		return r, errors.New("could not verify the original active pane; no focus change")
	}
	if parts[1] == "1" {
		return r, errors.New("the original pane process has exited; its retained screen is available as a preview")
	}
	if parts[1] != "0" {
		return r, errors.New("could not verify the original pane; refresh before attaching")
	}
	return r, nil
}

func (b *Backend) Stop(ctx context.Context, s session.Session) error {
	if _, err := b.validateTarget(ctx, s); err != nil {
		return err
	}
	_, err := b.run(ctx, maxControlBytes, "kill-session", "-t", s.BackendID)
	if err == nil {
		b.discovery.mu.Lock()
		delete(b.discovery.records, s.BackendID)
		b.discovery.mu.Unlock()
	}
	return err
}
