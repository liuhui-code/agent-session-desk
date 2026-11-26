package cli

import (
	"github.com/liuhui-code/agent-session-desk/internal/pet"
	"strings"
)

// WT treats semicolons as action separators even inside a command line. Escape
// that parser layer; exec.Command owns Windows argv quoting, without a shell.
func terminalHostArgs(executable, dir string, args []string, token string) []string {
	out := []string{"--window", "new", "new-tab", "--title", "", "--useApplicationTitle", "--startingDirectory", strings.ReplaceAll(dir, ";", `\;`), "--", strings.ReplaceAll(executable, ";", `\;`)}
	marker, _ := pet.WindowMarker(token)
	out[4] = marker
	if len(args) > 0 && args[len(args)-1] == "--" {
		args = args[:len(args)-1]
	}
	for _, arg := range args {
		out = append(out, strings.ReplaceAll(arg, ";", `\;`))
	}
	return append(out, "--inside-terminal", "--window-token", token)
}
