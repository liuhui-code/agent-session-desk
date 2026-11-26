package terminal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// The peer is deliberately a protocol simulator, not a psmux compatibility test.
func TestHeadlessPasteFaultContract(t *testing.T) {
	for _, mode := range []string{"fragmented", "auth-refused", "wrong-nonce", "duplicate", "oversized", "truncated", "lost-ack", "cancel-auth", "cancel-write", "cancel-ack", "timeout-auth", "timeout-ack", "ack-without-eof"} {
		t.Run(mode, func(t *testing.T) {
			limit := 2 * time.Second
			if strings.HasPrefix(mode, "timeout-") {
				limit = 200 * time.Millisecond
			}
			ctx, cancel := context.WithTimeout(context.Background(), limit)
			defer cancel()
			client, server := net.Pipe()
			done := make(chan error, 1)
			go func() {
				defer server.Close()
				r := bufio.NewReader(server)
				auth, err := r.ReadString('\n')
				if err != nil || auth != "AUTH private-key\n" {
					done <- fmt.Errorf("authentication not received")
					return
				}
				if mode == "cancel-auth" || mode == "timeout-auth" {
					if mode == "cancel-auth" {
						cancel()
					}
					_, err = r.ReadByte()
					done <- requirePeerClosed(err)
					return
				}
				if mode == "auth-refused" {
					_, _ = io.WriteString(server, "DENIED\n")
					_, err = r.ReadByte()
					done <- requirePeerClosed(err)
					return
				}
				_, _ = io.WriteString(server, "OK\n")
				if mode == "cancel-write" {
					// Consume only a prefix of the 64 KiB payload's control command.
					prefix := make([]byte, 16)
					_, _ = io.ReadFull(server, prefix)
					cancel()
					_, err = server.Read(prefix)
					// A pending Write may transfer another prefix before close; its
					// overall result must still be unconfirmed, checked by the caller.
					done <- nil
					return
				}
				command, err := r.ReadString('\n')
				parts := strings.Split(strings.TrimSpace(command), "display-message -p ")
				if err != nil || len(parts) != 2 || !strings.HasPrefix(command, "send-paste ") || !strings.Contains(command, " -t %7 ; ") {
					done <- fmt.Errorf("expected exactly one targeted paste/barrier")
					return
				}
				ack := parts[1]
				switch mode {
				case "fragmented":
					for _, ch := range "\n" + ack + "\n\n" {
						if _, err = fmt.Fprintf(server, "%c", ch); err != nil {
							break
						}
					}
				case "wrong-nonce":
					_, _ = fmt.Fprintln(server, "asd_delivery_foreign")
				case "duplicate":
					_, _ = fmt.Fprintf(server, "%s\n%s\n", ack, ack)
				case "oversized":
					_, _ = fmt.Fprintln(server, strings.Repeat("x", 8192))
				case "truncated":
					_, _ = fmt.Fprint(server, ack[:len(ack)-1])
				case "cancel-ack", "ack-without-eof":
					if mode == "ack-without-eof" {
						_, _ = fmt.Fprintln(server, ack)
					}
					cancel()
					_, err = r.ReadByte()
					done <- requirePeerClosed(err)
					return
				case "timeout-ack":
					_, err = r.ReadByte()
					done <- requirePeerClosed(err)
					return
				}
				done <- nil
			}()
			err := PastePSMux(ctx, client, "private-key", 7, strings.Repeat("x", MaxInputBytes))
			if (err == nil) != (mode == "fragmented") {
				t.Fatalf("mode %s: success=%v", mode, err == nil)
			}
			if err != nil && strings.Contains(err.Error(), "private-key") {
				t.Fatal("credential in transport error")
			}
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("simulator did not join after client close")
			}
		})
	}
}

func requirePeerClosed(err error) error {
	if err == nil {
		return fmt.Errorf("input emitted after refusal/cancellation")
	}
	return nil
}
