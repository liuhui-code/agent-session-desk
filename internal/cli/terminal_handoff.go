package cli

import "os/exec"

// GUI hosts can live as long as their window. Hand off ownership instead of
// keeping the console launcher (or its context cancellation) alive beside them.
func startTerminalHandoff(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
