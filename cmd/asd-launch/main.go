// The fixed portable launcher follows alpha.json after program updates.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/liuhui-code/agent-session-desk/internal/updater"
)

func main() { os.Exit(run()) }

func run() int {
	self, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ASD launcher: executable path unavailable")
		return 1
	}
	target, err := updater.Entrypoint(context.Background(), filepath.Dir(self))
	if err != nil {
		fmt.Fprintln(os.Stderr, "ASD launcher: installation invalid or update incomplete; run update status / repair")
		return 1
	}
	a, e1 := os.Stat(self)
	b, e2 := os.Stat(target)
	if e1 != nil || e2 != nil || os.SameFile(a, b) {
		fmt.Fprintln(os.Stderr, "ASD launcher: invalid recursive entrypoint")
		return 1
	}
	cmd := exec.Command(target, os.Args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = filepath.Dir(self)
	if err = cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "ASD launcher: could not start the selected program")
		return 1
	}
	return 0
}
