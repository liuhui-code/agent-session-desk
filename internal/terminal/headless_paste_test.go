package terminal

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestHeadlessPasteUsesNativePasteAndAckWithoutAttachResizeOrEnter(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	body := strings.Repeat("\u4e2d\u6587 ' $();\n", 1000)
	done := make(chan error, 1)
	go func() {
		defer server.Close()
		r := bufio.NewReader(server)
		auth, e := r.ReadString('\n')
		if e != nil {
			done <- e
			return
		}
		if auth != "AUTH secret\n" {
			done <- fmt.Errorf("authentication boundary changed")
			return
		}
		if _, e = fmt.Fprint(server, "OK\n"); e != nil {
			done <- e
			return
		}
		line, e := r.ReadString('\n')
		if e != nil {
			done <- e
			return
		}
		for _, forbidden := range []string{"client-size", "client-attach", "PERSISTENT", "send-keys", "Enter"} {
			if strings.Contains(line, forbidden) {
				done <- fmt.Errorf("unexpected UI or submit command")
				return
			}
		}
		if !strings.Contains(line, "send-paste "+base64.StdEncoding.EncodeToString([]byte(body))) || !strings.Contains(line, " -t %7") {
			done <- fmt.Errorf("literal bytes or exact pane changed")
			return
		}
		pieces := strings.Split(strings.TrimSpace(line), "display-message -p ")
		if len(pieces) != 2 {
			done <- fmt.Errorf("missing completion barrier")
			return
		}
		_, e = fmt.Fprintln(server, pieces[1])
		done <- e
	}()
	if e := PastePSMux(context.Background(), client, "secret", 7, body); e != nil {
		t.Fatal(e)
	}
	if e := <-done; e != nil {
		t.Fatal(e)
	}
}
func TestHeadlessPasteMissingAckRemainsUncertain(t *testing.T) {
	client, server := net.Pipe()
	go func() {
		defer server.Close()
		r := bufio.NewReader(server)
		_, _ = r.ReadString('\n')
		_, _ = fmt.Fprint(server, "OK\n")
		_, _ = r.ReadString('\n')
	}()
	if e := PastePSMux(context.Background(), client, "secret", 7, "task"); e == nil {
		t.Fatal("missing barrier ack reported success")
	}
}
