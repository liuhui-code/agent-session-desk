// Package cli provides the development slice's entry points. These commands do
// expose bounded explicit messaging, never durable agent conversation state.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const Version = "0.0.0-m0-dev"

const usage = `Agent Session Desk — M0 source migration slice

Usage:
  asd [global options]                         Open the session manager
  asd [global options] doctor                  Probe the terminal backend
  asd [global options] start [options] -- EXE [ARG...]
  asd [global options] history [--limit N]      Read recorded metadata as JSON
  asd [global options] history [--limit N]     Read recorded metadata as JSON
  asd [global options] read [--lines N] SESSION_ID  Capture source-tagged screen lines
  asd [global options] send --request-id ID --to SESSION_ID --run RUN_ID --body-file FILE
  asd [global options] deliver MESSAGE_ID       Explicit backend paste; no UI or appended Enter
  asd [global options] status MESSAGE_ID        Read a delivery receipt offline
  asd [global options] agent-bind --session ID --run ID --token-file FILE
  asd [global options] inbox --token-file FILE  Read bound Run requests without replies
  asd [global options] attention-event --token-file FILE --event-file FILE
  asd [global options] attention-list           Read pending native decision requests
  asd [global options] attention-status ID      Query one request without answering
  asd [global options] reply --token-file FILE --claim-file FILE --reply-id ID --reply-to ID --body-file FILE
  asd [global options] claim --token-file FILE --claim-file FILE --claim-id ID --message-id ID [--lease-seconds N]
  asd [global options] renew --token-file FILE --claim-file FILE
  asd [global options] claim-status MESSAGE_ID   Read processing ownership offline
  asd [global options] replies MESSAGE_ID       Read recorded replies offline
  asd [global options] record SESSION_ID       Save a bounded screen snapshot
  asd [global options] output RUN_ID           Read retained screen records as JSON
  asd [global options] list [--json]            List live backend records
  asd [global options] preview SESSION_ID      Read a bounded snapshot as JSON
  asd [global options] attach SESSION_ID       Enter a full-screen terminal
  asd [global options] stop SESSION_ID         End this Run's backend session
  asd [global options] metadata check [FILE]   Check metadata without a backend
  asd [global options] metadata backup FILE    Back up metadata to a new file
  asd [global options] metadata restore FILE   Restore into an absent database
  asd version
  asd pet-demo [--state idle|approval|question|unknown]  Native Windows visual demo

Global options (before the subcommand):
  --backend PATH       psmux on Windows; tmux for Unix development
  --terminal-bundle DIR  Explicit pinned Windows x64 development terminal bundle
  --runtime-dir DIR    Product-owned backend data/config directory
  --namespace NAME     Isolated product backend namespace
  --state-dir DIR      Persistent metadata directory
  --state-dir DIR      Persistent metadata directory
  --config PATH        TOML launch configuration (default UI template or start --template)
  --pet                Show existing blocking reminders in a Windows desktop pet
  --document-app APP   Windows workbench: default or system-notepad

Start options:
  --name NAME          Session label (defaults to executable name)
  --cwd DIR            Working directory (defaults to current directory)
  --env KEY=value      Per-launch override; repeat as needed
  --template NAME      Start a named template from --config instead of EXE

Windows Tab workspace: type in the terminal; F6 switches management focus, Ctrl+PgUp/PgDn switches Tabs.
Management focus: arrows/j/k select, Enter opens, n creates, x closes only the Tab, r retries, ? help, q quits.
Unix development / auxiliary mode: arrows select, Enter attaches; a explicitly attaches from Tab management.
Inside the terminal: cancel IME composition, switch to English input, then Ctrl+b and d detach back to the manager.
Press h for read-only metadata history; Enter shows details and Esc returns.
Quitting the manager keeps backend sessions; stop explicitly ends the selected Run.

This is a development slice. Recorded metadata does not establish liveness.
Automatic delivery, agent acceptance/completion, conversation storage and full Windows acceptance remain unfinished.
`

type envFlags []string

func (e *envFlags) String() string { return "" } // Do not echo environment values.
func (e *envFlags) Set(value string) error {
	*e = append(*e, value)
	return nil
}

// Run is separated from os.Exit so argument/terminal boundaries can be checked.
func Run(ctx context.Context, args []string, input io.Reader, output, errorOutput io.Writer) int {
	if err := run(ctx, args, input, output, errorOutput); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(errorOutput, "asd:", safeError(err.Error()))
		return 1
	}
	return 0
}

func run(ctx context.Context, args []string, input io.Reader, output, errorOutput io.Writer) error {
	flags := flag.NewFlagSet("asd", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	var options backend.Options
	var configPath, stateDir, documentApp string
	var insideTerminal, desktopPet bool
	flags.StringVar(&documentApp, "document-app", "", "external Markdown app: default or system-notepad")
	flags.BoolVar(&desktopPet, "pet", false, "native desktop blocking reminder")
	flags.BoolVar(&insideTerminal, "inside-terminal", false, "internal terminal handoff")
	flags.StringVar(&options.Binary, "backend", "", "terminal backend executable")
	flags.StringVar(&options.TerminalBundle, "terminal-bundle", "", "pinned Windows x64 development terminal bundle")
	flags.StringVar(&options.RuntimeDir, "runtime-dir", "", "product runtime directory")
	flags.StringVar(&options.Namespace, "namespace", "", "product namespace")
	flags.StringVar(&stateDir, "state-dir", "", "persistent metadata directory")
	flags.StringVar(&configPath, "config", "", "explicit TOML launch configuration")
	flags.Usage = func() { fmt.Fprint(output, usage) }
	if err := flags.Parse(args); err != nil {
		return err
	}
	commandArgs := flags.Args()
	command := ""
	if len(commandArgs) > 0 {
		command, commandArgs = commandArgs[0], commandArgs[1:]
	}
	if documentApp != "" && (command != "" || runtime.GOOS != "windows" || (documentApp != "default" && documentApp != "system-notepad")) {
		return fmt.Errorf("--document-app requires the Windows workbench and default or system-notepad")
	}
	if desktopPet && (command != "" || runtime.GOOS != "windows") {
		return fmt.Errorf("--pet requires the Windows workbench without a subcommand")
	}
	if command == "help" {
		fmt.Fprint(output, usage)
		return nil
	}
	if command == "version" {
		if len(commandArgs) > 0 {
			return fmt.Errorf("version accepts no arguments")
		}
		fmt.Fprintln(output, "Agent Session Desk", Version)
		return nil
	}
	if command == "pet-demo" {
		if configPath != "" || stateDir != "" || options.RuntimeDir != "" || options.Namespace != "" || options.Binary != "" || options.TerminalBundle != "" {
			return fmt.Errorf("pet-demo uses no backend/config/state options")
		}
		return petDemo(ctx, commandArgs, errorOutput)
	}
	if command == "read" || command == "send" || command == "status" || command == "deliver" {
		if configPath != "" {
			return fmt.Errorf("--config is not used by messaging commands")
		}
		return messagingCommand(ctx, command, commandArgs, options, stateDir, input, output, errorOutput)
	}
	if command == "agent-bind" || command == "inbox" || command == "reply" || command == "replies" {
		if configPath != "" {
			return fmt.Errorf("--config is not used by messaging commands")
		}
		return agentMessagingCommand(ctx, command, commandArgs, options, stateDir, input, output, errorOutput)
	}
	if command == "claim" || command == "renew" || command == "claim-status" {
		if configPath != "" {
			return fmt.Errorf("--config is not used by messaging commands")
		}
		return taskClaimingCommand(ctx, command, commandArgs, options, stateDir, output, errorOutput)
	}
	if command == "attention-event" || command == "attention-list" || command == "attention-status" {
		if configPath != "" {
			return fmt.Errorf("--config is not used by attention commands")
		}
		return attentionCommand(ctx, command, commandArgs, options, stateDir, input, output, errorOutput)
	}
	if command != "" && command != "doctor" && command != "start" && command != "list" && command != "preview" && command != "attach" && command != "stop" && command != "history" && command != "metadata" && command != "output" && command != "record" {
		return fmt.Errorf("unknown command %q; run asd help", command)
	}
	if configPath != "" && command != "start" && command != "" {
		return fmt.Errorf("--config is used by the UI default template or start --template")
	}
	if command == "" {
		var err error
		options, stateDir, err = applyPortableOptions(options, stateDir)
		if err != nil {
			return err
		}
		if handled, err := openTerminalHost(ctx, args, insideTerminal, input, output); handled {
			return err
		}
	}
	if command == "metadata" {
		return metadataCommand(ctx, commandArgs, options, stateDir, output)
	}
	// Validate command shape before touching the backend or its runtime files.
	var launch session.Launch
	var err error
	jsonList := false
	limit := 100
	switch command {
	case "output":
		if len(commandArgs) != 1 || !validID(commandArgs[0]) {
			return fmt.Errorf("output requires one full run_id from asd history")
		}
	case "history":
		limit, err = historyLimit(commandArgs, errorOutput)
		if err != nil {
			return err
		}
	case "start":
		launch, err = parseLaunchWithConfig(commandArgs, configPath, output, errorOutput)
		if err != nil {
			return err
		}
		launch, err = session.ResolveLaunch(launch)
		if err != nil {
			return err
		}
	case "list":
		if len(commandArgs) == 1 && commandArgs[0] == "--json" {
			jsonList = true
		} else if len(commandArgs) != 0 {
			return fmt.Errorf("usage: asd list [--json]")
		}
	case "preview", "record", "attach", "stop":
		if len(commandArgs) != 1 || !validID(commandArgs[0]) {
			return fmt.Errorf("%s requires one full session_id from asd list", command)
		}
	default:
		if len(commandArgs) != 0 {
			return fmt.Errorf("%s accepts no arguments", command)
		}
	}
	if command == "history" || command == "output" {
		metadata, e := openMetadata(ctx, options, stateDir)
		if e != nil {
			return e
		}
		defer metadata.Close()
		if command == "output" {
			rows, e := metadata.Output(ctx, commandArgs[0])
			if e != nil {
				return e
			}
			return writeJSON(output, rows)
		}
		rows, e := metadata.History(ctx, limit)
		if e != nil {
			return e
		}
		return writeJSON(output, rows)
	}
	terminal, err := backend.New(options)
	if err != nil {
		return err
	}
	info, err := terminal.Probe(ctx)
	if err != nil {
		return err
	}
	if command == "doctor" {
		return writeJSON(output, struct {
			Name                string `json:"backend"`
			Version             string `json:"version"`
			Experimental        bool   `json:"experimental"`
			WindowsValidated    bool   `json:"windows_validated"`
			ProvenanceValidated bool   `json:"binary_provenance_validated"`
			Note                string `json:"note"`
		}{info.Name, info.Version, info.Experimental, false, false,
			"Version probe only; record official binary hash and run native Windows acceptance separately."})
	}
	metadata, err := openMetadata(ctx, options, stateDir)
	if err != nil {
		return err
	}
	defer metadata.Close()
	service := manager.New(terminal, metadata)
	if command == "" {
		return runWorkbench(ctx, service, options, stateDir, configPath, desktopPet, documentApp, input, output)
	}
	if command == "start" {
		created, err := service.Start(ctx, launch)
		if err != nil {
			return err
		}
		rememberLaunchDirectory(ctx, options, stateDir, launch.Dir, errorOutput)
		if err := rememberCLIStart(ctx, options, stateDir, launch); err != nil {
			fmt.Fprintln(errorOutput, err)
		}
		return writeJSON(output, created)
	}
	sessions, err := service.List(ctx)
	if err != nil {
		return err
	}
	if command == "list" {
		if sessions == nil {
			sessions = []session.Session{}
		}
		if jsonList {
			return writeJSON(output, sessions)
		}
		if len(sessions) == 0 {
			fmt.Fprintln(output, "No backend sessions. Start one with: asd start --name NAME -- EXE ARG...")
			return nil
		}
		fmt.Fprintln(output, "SESSION_ID                        TERMINAL    NAME")
		for _, s := range sessions {
			fmt.Fprintf(output, "%s  %-10s  %s\n", s.ID, s.TerminalState, strconv.QuoteToGraphic(s.Title))
		}
		fmt.Fprintln(output, "Terminal availability does not establish agent readiness or progress.")
		return nil
	}
	target, err := findSession(sessions, commandArgs[0])
	if err != nil {
		return err
	}
	switch command {
	case "record":
		record, e := service.RecordOutput(ctx, target)
		if e != nil {
			return e
		}
		return writeJSON(output, record)
	case "preview":
		preview, err := service.Preview(ctx, target, 60)
		if err != nil {
			return err
		}
		return writeJSON(output, struct {
			SessionID string          `json:"session_id"`
			RunID     string          `json:"run_id"`
			Kind      string          `json:"kind"`
			Snapshot  session.Preview `json:"snapshot"`
		}{target.ID, target.RunID, "screen_snapshot", preview})
	case "attach":
		cmd, err := service.Attach(ctx, target)
		if err != nil {
			return err
		}
		cmd.Stdin, cmd.Stdout, cmd.Stderr = input, output, errorOutput
		return cmd.Run()
	case "stop":
		if err := service.Stop(ctx, target); err != nil {
			return err
		}
		return writeJSON(output, struct {
			SessionID string `json:"session_id"`
			RunID     string `json:"run_id"`
			Result    string `json:"result"`
		}{target.ID, target.RunID, "backend_session_stopped"})
	}
	return nil
}

func findSession(sessions []session.Session, id string) (session.Session, error) {
	var found *session.Session
	for _, s := range sessions {
		if s.ID == id {
			if found != nil {
				return session.Session{}, fmt.Errorf("session %s has ambiguous backend records; no action taken", id)
			}
			copy := s
			found = &copy
		}
	}
	if found != nil {
		return *found, nil
	}
	return session.Session{}, fmt.Errorf("session %s is not present in this backend; use asd history for metadata; no implicit restart is available", id)
}

func validID(id string) bool {
	if len(id) != 32 {
		return false
	}
	for _, r := range id {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func safeError(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return ' '
		}
		return r
	}, value)
}
