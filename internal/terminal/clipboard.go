package terminal

import (
	"context"
	"encoding/base64"
	"strings"
	"unicode/utf8"
)

// MaxClipboardBytes is a defensive bound, not a resource budget or log limit.
const MaxClipboardBytes = 64 * 1024

// ClipboardSource reports validated copy payloads separately from coalesced
// screens. Only the latest unread payload is retained; this never reads the
// user's clipboard or sends its contents to a child process.
type ClipboardSource interface {
	NextClipboard(context.Context) (string, error)
}

func (c *PSMux) NextClipboard(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	select {
	case <-c.done:
		return "", c.failure()
	default:
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-c.done:
		return "", c.failure()
	case text := <-c.clipboard:
		select {
		case <-c.done:
			return "", c.failure()
		default:
			return text, nil
		}
	}
}

func (c *PSMux) receiveClipboard(encoded string) {
	if encoded == "" || len(encoded) > base64.StdEncoding.EncodedLen(MaxClipboardBytes) {
		return
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(data) == 0 || len(data) > MaxClipboardBytes || !utf8.Valid(data) || strings.IndexByte(string(data), 0) >= 0 {
		return
	}
	text := string(data)
	select {
	case c.clipboard <- text:
	default:
		select {
		case <-c.clipboard:
		default:
		}
		select {
		case c.clipboard <- text:
		default:
		}
	}
}
