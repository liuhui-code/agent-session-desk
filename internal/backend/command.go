package backend

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func (b *Backend) command(args ...string) *exec.Cmd {
	all := append([]string{"-L", b.namespace, "-f", b.config}, args...)
	cmd := exec.Command(b.bin, all...)
	cmd.Env = commandEnv(os.Environ(), b.kind, b.dataDir, b.config)
	if b.bundle != "" {
		cmd.Env = append(cmd.Env, "PSMUX_CONPTY_DIR="+b.bundle, "PSMUX_CONPTY_PIPE_MOUSE_ONLY=1")
	}
	return cmd
}

type commandOutput struct {
	text      string
	truncated bool
}

func (b *Backend) run(ctx context.Context, limit int, args ...string) (commandOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	base := b.command(args...)
	cmd := exec.CommandContext(ctx, base.Path, base.Args[1:]...)
	cmd.Env = base.Env
	configureCommand(cmd)
	cmd.WaitDelay = time.Second
	stdout := &tailBuffer{limit: limit}
	stderr := &tailBuffer{limit: 4096}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := cmd.Run()
	out := commandOutput{text: string(stdout.data), truncated: stdout.truncated}
	if err != nil {
		if ctx.Err() != nil {
			return out, fmt.Errorf("%s backend command cancelled or timed out: %w", b.kind, ctx.Err())
		}
		return out, &commandError{kind: b.kind, err: err, stderr: string(stderr.data)}
	}
	return out, nil
}

type commandError struct {
	kind   string
	err    error
	stderr string
}

func (e *commandError) Error() string {
	return fmt.Sprintf("%s backend command failed: %v: %s", e.kind, e.err, strings.TrimSpace(e.stderr))
}
func (e *commandError) Unwrap() error { return e.err }
func noServer(err error) bool {
	var command *commandError
	var exit *exec.ExitError
	if !errors.As(err, &command) || !errors.As(err, &exit) || exit.ExitCode() != 1 {
		return false
	}
	s := strings.ToLower(command.stderr)
	return strings.Contains(s, "no server running") || (strings.Contains(s, "error connecting to") && strings.Contains(s, "no such file or directory"))
}

type tailBuffer struct {
	data      []byte
	limit     int
	truncated bool
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if len(b.data)+n > b.limit {
		b.truncated = true
	}
	if n >= b.limit {
		b.data = append(b.data[:0], p[n-b.limit:]...)
		return n, nil
	}
	if len(b.data)+n > b.limit {
		drop := len(b.data) + n - b.limit
		copy(b.data, b.data[drop:])
		b.data = b.data[:len(b.data)-drop]
	}
	b.data = append(b.data, p...)
	return n, nil
}
