package terminal

import "context"

// Connection operates one validated terminal Run. Calls are not agent task
// receipts. Close only detaches the client; callers stop a Run separately.
type Connection interface {
	Next(context.Context) (Screen, error)
	SendText(context.Context, string) error
	SendKey(context.Context, string) error
	Paste(context.Context, string) error
	Resize(context.Context, int, int) error
	Refresh(context.Context) error
	Scroll(context.Context, int, int, bool) error
	Close() error
}
