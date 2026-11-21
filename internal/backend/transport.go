package backend

import (
	"fmt"
	"strings"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// The multiplexer parses its own command language before handing argv to the
// child. Reject known nonliteral forms until a verified transport replaces this
// M0 limitation. Never silently split, quote, or execute a different command.
func validateTransport(kind string, launch session.Launch) error {
	for _, value := range launch.Env {
		key, _, _ := strings.Cut(value, "=")
		upper := strings.ToUpper(key)
		if strings.HasPrefix(upper, "PSMUX_") || upper == "TMUX" || upper == "TMUX_PANE" || upper == "TMUX_TMPDIR" {
			return fmt.Errorf("environment key %s is reserved for the isolated terminal backend", key)
		}
		if kind == "psmux" {
			// Stable 3.4.0 parse_env_assignment trims the whole assignment
			// and accepts only ASCII identifier names. Do not trim for it.
			if strings.TrimSpace(value) != value {
				return fmt.Errorf("environment key %s has whitespace that psmux 3.4.0 would trim; this M0 transport cannot preserve it", key)
			}
			for i, c := range key {
				letter := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'
				if !letter && !(i > 0 && c >= '0' && c <= '9') {
					return fmt.Errorf("environment key %s is outside psmux 3.4.0's ASCII identifier contract", key)
				}
			}
		}
	}
	type payload struct{ label, value string }
	items := []payload{{"executable", launch.Executable}, {"working directory", launch.Dir}}
	for i, value := range launch.Env {
		items = append(items, payload{fmt.Sprintf("environment entry %d", i+1), value})
	}
	for i, value := range launch.Args {
		items = append(items, payload{fmt.Sprintf("argument %d", i+1), value})
	}
	for _, item := range items {
		arg := item.value
		if kind == "psmux" {
			// Stable 3.4.0 src/cli.rs normalize_flag_equals rewrites these
			// even after --. split_cli_command_list also consumes ; and \;.
			if len(arg) >= 4 && arg[0] == '-' && ((arg[1] >= 'a' && arg[1] <= 'z') || (arg[1] >= 'A' && arg[1] <= 'Z')) && arg[2] == '=' {
				return fmt.Errorf("%s uses -x=value syntax that psmux 3.4.0 rewrites after --; this M0 transport cannot preserve it", item.label)
			}
			if arg == ";" || arg == `\;` {
				return fmt.Errorf("%s is a command separator in psmux 3.4.0; this M0 transport cannot preserve it literally", item.label)
			}
		} else if strings.HasSuffix(arg, ";") {
			// tmux cmd_parse_from_arguments recognizes trailing semicolons
			// before command flag parsing. Do not guess at double escaping.
			return fmt.Errorf("%s ends in a tmux command separator; this M0 transport cannot preserve it literally", item.label)
		}
	}
	return nil
}
