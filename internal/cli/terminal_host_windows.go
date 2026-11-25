//go:build windows

package cli

import (
	"context"
	"fmt"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func openTerminalHost(ctx context.Context, args []string, inside bool, input io.Reader, output io.Writer) (bool, error) {
	if inside || os.Getenv("WT_SESSION") != "" || input != os.Stdin || output != os.Stdout {
		return false, nil
	}
	executable, err := os.Executable()
	if err != nil {
		return true, err
	}
	host, err := localTerminalHost(filepath.Dir(executable))
	if err != nil {
		return true, err
	}
	binary := host.Path
	if binary == "" {
		binary, err = exec.LookPath("wt.exe")
		if err != nil {
			return true, fmt.Errorf("Windows Terminal is required; install it or open this program inside Windows Terminal")
		}
	}
	dir, err := os.Getwd()
	if err != nil {
		return true, err
	}
	hostArgs := terminalHostArgs(executable, dir, args)
	if host.Profile != "" {
		hostArgs = append([]string{"--window", "new", "new-tab", "--profile", host.Profile}, hostArgs[3:]...)
	}
	cmd := exec.CommandContext(ctx, binary, hostArgs...)
	if err = cmd.Run(); err != nil {
		return true, fmt.Errorf("could not open Windows Terminal: %w", err)
	}
	return true, nil
}

func applyPortableOptions(opts backend.Options, stateDir string) (backend.Options, string, error) {
	executable, err := os.Executable()
	if err != nil {
		return opts, stateDir, err
	}
	return portableOptions(executable, opts, stateDir)
}
