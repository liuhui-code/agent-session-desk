//go:build windows

package backend

import (
	"os/exec"
	"syscall"
)

// Background metadata queries must not flash new console windows. The full
// screen attach command deliberately does not use this setting.
func configureCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
