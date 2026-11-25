package cli

import "strings"

// WT treats semicolons as action separators even inside a command line. Escape
// that parser layer; exec.Command owns Windows argv quoting, without a shell.
func terminalHostArgs(executable, dir string, args []string) []string {
	out := []string{"--window", "new", "new-tab", "--startingDirectory", strings.ReplaceAll(dir, ";", `\;`), "--", strings.ReplaceAll(executable, ";", `\;`)}
	for _, arg := range args {
		out = append(out, strings.ReplaceAll(arg, ";", `\;`))
	}
	return append(out, "--inside-terminal")
}
