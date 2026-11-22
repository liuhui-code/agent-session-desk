package backend

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

// OpenTerminal connects to an existing psmux-owned terminal. It never starts or
// replaces an agent, shell, server, or pseudoconsole. Unix keeps the explicit
// full-screen attach path until a separate native terminal adapter is verified.
func (b *Backend) SupportsTerminal() bool { return b.kind == "psmux" }

func (b *Backend) OpenTerminal(ctx context.Context, s session.Session, width, height int) (terminal.Connection, error) {
	if b.kind != "psmux" {
		return nil, errors.New("Tab terminal transport is currently verified only for psmux 3.4.0")
	}
	if width < 1 || width > terminal.MaxWidth || height < 1 || height > terminal.MaxHeight {
		return nil, errors.New("terminal viewport exceeds screen bounds")
	}
	if _, err := b.Probe(ctx); err != nil {
		return nil, err
	}
	r, err := b.validateTarget(ctx, s)
	if err != nil {
		return nil, err
	}
	// Reuse the existing attach checks: live original pane and no focus steal.
	if _, err = b.Attach(ctx, s); err != nil {
		return nil, err
	}
	base := filepath.Join(b.dataDir, b.namespace+"__"+s.BackendID)
	portText, err := readTerminalEndpoint(base + ".port")
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil || port == 0 {
		return nil, errors.New("invalid product terminal endpoint")
	}
	key, err := readTerminalEndpoint(base + ".key")
	if err != nil {
		return nil, err
	}
	pane, err := strconv.Atoi(strings.TrimPrefix(r.PaneID, "%"))
	if err != nil {
		return nil, errors.New("invalid product pane identity")
	}
	conn, err := (&net.Dialer{Timeout: 3 * time.Second}).DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(int(port))))
	if err != nil {
		return nil, errors.New("could not connect to product terminal endpoint")
	}
	// The first native frame must name the same session and active original pane
	// before any user input can be written. Endpoint keys and frames are not logged.
	return terminal.OpenPSMux(ctx, conn, terminal.PSMuxOptions{Key: key, Session: s.BackendID, Pane: pane, Width: width, Height: height})
}

func readTerminalEndpoint(path string) (string, error) {
	st, err := os.Lstat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() < 1 || st.Size() > 256 {
		return "", errors.New("product terminal endpoint file is unavailable or invalid")
	}
	f, err := os.Open(path)
	if err != nil {
		return "", errors.New("could not read product terminal endpoint")
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(st, opened) {
		return "", errors.New("product terminal endpoint changed while opening")
	}
	data, err := io.ReadAll(io.LimitReader(f, 257))
	if err != nil || len(data) > 256 {
		return "", errors.New("product terminal endpoint exceeded its byte limit")
	}
	value := strings.TrimSpace(string(data))
	if value == "" || strings.ContainsAny(value, "\r\n\x00") {
		return "", errors.New("invalid product terminal endpoint value")
	}
	return value, nil
}
