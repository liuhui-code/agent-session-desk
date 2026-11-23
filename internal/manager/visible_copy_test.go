package manager_test

import (
	"context"
	"errors"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"testing"
)

type visibleTerminal struct {
	terminal
	captured   []session.Session
	captureErr error
}

func (f *visibleTerminal) CaptureVisible(_ context.Context, s session.Session) (session.Preview, error) {
	f.captured = append(f.captured, s)
	return session.Preview{Text: "fresh-visible"}, f.captureErr
}
func TestVisibleCopyDelegatesWithoutMetadataOrOtherTerminalEffects(t *testing.T) {
	f := &visibleTerminal{}
	m := manager.New(f, nil)
	target := observed()
	p, e := m.CaptureVisible(context.Background(), target)
	if e != nil || p.Text != "fresh-visible" || len(f.captured) != 1 || f.captured[0] != target || f.starts != 0 || f.stops != 0 {
		t.Fatal("capture changed target/lifetime or used metadata")
	}
	f.captureErr = errors.New("fresh capture failed")
	if _, e = m.CaptureVisible(context.Background(), target); e == nil {
		t.Fatal("capture error masked")
	}
	if _, e = manager.New(&terminal{}, nil).CaptureVisible(context.Background(), target); e == nil {
		t.Fatal("unsupported capture silently fell back")
	}
}
