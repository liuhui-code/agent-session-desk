package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/liuhui-code/agent-session-desk/internal/config"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func parseLaunch(args []string, output, errorOutput io.Writer) (session.Launch, error) {
	return parseLaunchWithConfig(args, "", output, errorOutput)
}

func parseLaunchWithConfig(args []string, path string, output, errorOutput io.Writer) (session.Launch, error) {
	var l session.Launch
	var env envFlags
	var id string
	flags := flag.NewFlagSet("start", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	flags.StringVar(&l.Name, "name", "", "session name")
	flags.StringVar(&l.Dir, "cwd", "", "working directory")
	flags.StringVar(&id, "template", "", "named launch template")
	flags.Var(&env, "env", "KEY=value (repeatable)")
	flags.Usage = func() { fmt.Fprint(output, usage) }
	if err := flags.Parse(args); err != nil {
		return l, err
	}
	selected := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "template" {
			selected = true
		}
	})
	if selected {
		if flags.NArg() != 0 {
			return l, fmt.Errorf("--template cannot be combined with an explicit executable")
		}
		if path == "" {
			return l, fmt.Errorf("--template requires --config PATH")
		}
		t, err := config.LoadLaunch(path, id)
		if err != nil {
			return t, err
		}
		flags.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "name":
				t.Name = l.Name
			case "cwd":
				t.Dir = l.Dir
			}
		})
		t.Env = append(t.Env, env...)
		return t, t.Validate()
	}
	if path != "" {
		return l, fmt.Errorf("--config requires start --template NAME")
	}
	if flags.NArg() == 0 {
		return l, fmt.Errorf("usage: asd start [--name NAME] [--cwd DIR] [--env KEY=value] -- EXE [ARG...]")
	}
	l.Executable, l.Args, l.Env = flags.Arg(0), flags.Args()[1:], env
	return l, l.Validate()
}
