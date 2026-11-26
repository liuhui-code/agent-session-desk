package backend

import (
	"context"
	"errors"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

func (b *Backend) SupportsMessageDelivery() bool { return b.kind == "psmux" }

// Paste writes to the original pane through psmux's control queue. It neither
// creates a UI client nor changes terminal size, focus or process lifetime.
func (b *Backend) Paste(ctx context.Context, target session.Session, text string) error {
	if !b.SupportsMessageDelivery() {
		return errors.New("headless delivery is currently verified only for psmux")
	}
	if err := session.ValidateMessageBody(text); err != nil {
		return err
	}
	record, err := b.validateTarget(ctx, target)
	if err != nil {
		return err
	}
	// Reuse the original pane/bundle/liveness checks; do not execute attach.
	if _, err = b.Attach(ctx, target); err != nil {
		return err
	}
	base := filepath.Join(b.dataDir, b.namespace+"__"+target.BackendID)
	portText, err := readTerminalEndpoint(base + ".port")
	if err != nil {
		return err
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil || port == 0 {
		return errors.New("invalid headless terminal endpoint")
	}
	key, err := readTerminalEndpoint(base + ".key")
	if err != nil {
		return err
	}
	pane, err := strconv.Atoi(strings.TrimPrefix(record.PaneID, "%"))
	if err != nil {
		return errors.New("invalid original pane identity")
	}
	conn, err := (&net.Dialer{Timeout: 3 * time.Second}).DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(int(port))))
	if err != nil {
		return errors.New("headless terminal endpoint unavailable")
	}
	return terminal.PastePSMux(ctx, conn, key, pane, text)
}
