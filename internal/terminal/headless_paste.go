package terminal

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
	"unicode/utf8"
)

// PastePSMux uses the existing authenticated one-shot control path. A server
// FIFO display-message barrier acknowledges processing, not agent acceptance.
// There is no persistent client attach, viewport change or appended Enter.
func PastePSMux(ctx context.Context, conn net.Conn, key string, pane int, text string) error {
	defer conn.Close()
	if key == "" || strings.ContainsAny(key, "\r\n\x00") || pane < 0 || !utf8.ValidString(text) || len(text) == 0 || len(text) > MaxInputBytes {
		return errors.New("invalid headless paste parameters")
	}
	deadline := time.Now().Add(5 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(conn, "AUTH %s\n", key); err != nil {
		return errors.New("headless terminal authentication failed")
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024), 4096)
	if !scanner.Scan() || scanner.Text() != "OK" {
		return errors.New("headless terminal authentication failed")
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return errors.New("could not create delivery barrier")
	}
	ack := "asd_delivery_" + hex.EncodeToString(nonce[:])
	command := fmt.Sprintf("send-paste %s -t %%%d ; display-message -p %s\n", base64.StdEncoding.EncodeToString([]byte(text)), pane, ack)
	if _, err := fmt.Fprint(conn, command); err != nil {
		return errors.New("headless paste write uncertain; do not replay")
	}
	acknowledged := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line != ack || acknowledged {
			return errors.New("headless paste response unconfirmed; do not replay")
		}
		acknowledged = true
	}
	if scanner.Err() != nil || !acknowledged {
		return errors.New("headless paste acknowledgement missing; do not replay")
	}
	return nil
}
