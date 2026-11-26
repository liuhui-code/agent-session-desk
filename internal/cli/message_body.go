package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func readMessageBody(file string, input io.Reader) (string, error) {
	reader := input
	if file != "-" {
		r, err := os.Open(file)
		if err != nil {
			return "", fmt.Errorf("cannot read message file")
		}
		defer r.Close()
		reader = r
	}
	if reader == nil {
		return "", fmt.Errorf("message stdin unavailable")
	}
	data, err := io.ReadAll(io.LimitReader(reader, session.MaxMessageBytes+4))
	if err != nil {
		return "", fmt.Errorf("cannot read message body")
	}
	body := string(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}))
	return body, session.ValidateMessageBody(body)
}
