//go:build !windows

package backend

import "os/exec"

func configureCommand(cmd *exec.Cmd) {}
