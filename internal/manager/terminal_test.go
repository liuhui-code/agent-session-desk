package manager_test

import (
	"context"
	"errors"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	term "github.com/liuhui-code/agent-session-desk/internal/terminal"
)

type workspaceTerminal struct {
	*terminal
	target        session.Session
	width, height int
}

var connectResult = errors.New("synthetic connection outcome")

func (f *workspaceTerminal) OpenTerminal(_ context.Context, s session.Session, w, h int) (term.Connection, error) {
	f.target = s
	f.width = w
	f.height = h
	return nil, connectResult
}

func TestOpenTerminalForwardsStableRunWithoutStartingOrStopping(t *testing.T) {
	backend := &workspaceTerminal{terminal: &terminal{}}
	m := manager.New(backend, nil)
	_, err := m.OpenTerminal(context.Background(), observed(), 80, 20)
	if !errors.Is(err, connectResult) {
		t.Fatalf("connection result lost: %v", err)
	}
	if backend.target != observed() || backend.width != 80 || backend.height != 20 {
		t.Fatal("terminal target or geometry changed")
	}
	if backend.starts != 0 || backend.stops != 0 {
		t.Fatal("opening a view changed Run lifetime")
	}
}
