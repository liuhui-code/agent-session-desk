package backend

import (
	"context"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (b *Backend) Start(ctx context.Context, l session.Launch) (session.Session, error) {
	id, err := session.NewID()
	if err != nil {
		return session.Session{}, err
	}
	rid, err := session.NewID()
	if err != nil {
		return session.Session{}, err
	}
	return b.StartRun(ctx, l, id, rid)
}
