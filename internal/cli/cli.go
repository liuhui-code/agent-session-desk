// Package cli provides the development slice's entry points. These commands do
// not implement the M3 agent messaging API or durable conversation history.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

const Version = "0.0.0-m0-dev"

const usage = `Agent Session Desk — M0 source migration slice

Usage:
  asd [global options]                         Open the session manager
  asd [global options] doctor                  Probe the terminal backend
  asd [global options] start [options] -- EXE [ARG...]
  asd [global options] list [--json]            List live backend records
  asd [global options] preview SESSION_ID      Read a bounded snapshot as JSON
  asd [global options] attach SESSION_ID       Enter a full-screen terminal
  asd [global options] stop SESSION_ID         End this Run's backend session
  asd version

Global options (before the subcommand):
  --backend PATH       psmux on Windows; tmux for Unix development
  --runtime-dir DIR    Product-owned backend data/config directory
  --namespace NAME     Isolated product backend namespace

Start options:
  --name NAME          Session label (defaults to executable name)
  --cwd DIR            Working directory (defaults to current directory)
  --env KEY=value      Per-launch override; repeat as needed

Inside the manager: arrows/j/k select, Enter attaches, r refreshes, ? help, q quits.
Inside the terminal: Ctrl+b then d detaches back to the manager.
Quitting the manager keeps backend sessions; stop explicitly ends the selected Run.

This is a development slice. Native Windows acceptance, durable session storage,
cross-session messaging and attention notifications are not completed.
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
	flags.StringVar(&options.Binary, "backend", "", "terminal backend executable")
	flags.StringVar(&options.RuntimeDir, "runtime-dir", "", "product runtime directory")
	flags.StringVar(&options.Namespace, "namespace", "", "product namespace")
	flags.Usage = func() { fmt.Fprint(output, usage) }
	if err := flags.Parse(args); err != nil {
		return err
	}
	commandArgs := flags.Args()
	command := ""
	if len(commandArgs) > 0 {
		command, commandArgs = commandArgs[0], commandArgs[1:]
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
	if command != "" && command != "doctor" && command != "start" && command != "list" && command != "preview" && command != "attach" && command != "stop" {
		return fmt.Errorf("unknown command %q; run asd help", command)
	}
	// Validate command shape before touching the backend or its runtime files.
	var launch session.Launch
	var err error
	jsonList := false
	switch command {
	case "start":
		launch, err = parseLaunch(commandArgs, output, errorOutput)
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
	case "preview", "attach", "stop":
		if len(commandArgs) != 1 || !validID(commandArgs[0]) {
			return fmt.Errorf("%s requires one full session_id from asd list", command)
		}
	default:
		if len(commandArgs) != 0 {
			return fmt.Errorf("%s accepts no arguments", command)
		}
	}
	service, err := backend.New(options)
	if err != nil {
		return err
	}
	info, err := service.Probe(ctx)
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
	if command == "" {
		model := ui.New(ctx, service)
		defer model.Close()
		_, err := tea.NewProgram(model, tea.WithContext(ctx), tea.WithInput(input), tea.WithOutput(output)).Run()
		return err
	}
	if command == "start" {
		created, err := service.Start(ctx, launch)
		if err != nil {
			return err
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

func parseLaunch(args []string, output, errorOutput io.Writer) (session.Launch, error) {
	var l session.Launch
	var env envFlags
	flags := flag.NewFlagSet("start", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	flags.StringVar(&l.Name, "name", "", "session name")
	flags.StringVar(&l.Dir, "cwd", "", "working directory")
	flags.Var(&env, "env", "KEY=value (repeatable)")
	flags.Usage = func() { fmt.Fprint(output, usage) }
	if err := flags.Parse(args); err != nil {
		return l, err
	}
	if flags.NArg() == 0 {
		return l, fmt.Errorf("usage: asd start [--name NAME] [--cwd DIR] [--env KEY=value] -- EXE [ARG...]")
	}
	l.Executable, l.Args, l.Env = flags.Arg(0), flags.Args()[1:], env
	return l, l.Validate()
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
	return session.Session{}, fmt.Errorf("session %s is not present in this backend; no historical record or implicit restart is available in M0", id)
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
