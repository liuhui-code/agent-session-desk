package manager_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

type recoveryGuardTerminal struct{ terminal }

func (f *recoveryGuardTerminal) StartRun(context.Context, session.Launch, string, string) (session.Session, error) {
	f.starts++
	return session.Session{}, errors.New("unexpected start")
}

func TestRecoveryRefusesLiveUnknownOrStaleSourceBeforeLaunch(t *testing.T) {
	ctx := context.Background()
	db, e := store.Open(ctx, filepath.Join(t.TempDir(), "db"), "a")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	v := observed()
	if e = db.Save(ctx, v); e != nil {
		t.Fatal(e)
	}
	c, _ := session.NewAgentCredential(v.ID, v.RunID)
	if e = db.BindAgent(ctx, c, v); e != nil {
		t.Fatal(e)
	}
	d := recovery.Declaration{Protocol: "asd.agent-recovery", SchemaVersion: 1, SessionID: v.ID, RunID: v.RunID, AgentID: "synthetic-agent", ConversationID: "conversation-example-42", Resume: recovery.Resume{Executable: `C:\Tools\MyAgent\agent.exe`, Args: []string{"resume", "conversation-example-42"}, Cwd: `C:\Work\project`}}

	if _, e = db.RegisterRecovery(ctx, c, d, 0); e != nil {
		t.Fatal(e)
	}
	nc, _ := session.NewAgentCredential(v.ID, "33333333333333333333333333333333")
	request := manager.RecoveryRequest{SessionID: v.ID, PreviousRunID: v.RunID, Revision: 1, ConfirmEnded: true, AttemptID: "44444444444444444444444444444444", Credential: nc, ContextFile: filepath.Join(t.TempDir(), "context.json")}
	for _, state := range []string{"live", "retained-exit", "unknown", "disconnected", "stale", "unconfirmed"} {
		t.Run(state, func(t *testing.T) {
			f := &recoveryGuardTerminal{terminal: terminal{items: []session.Session{v}}}
			r := request
			switch state {
			case "retained-exit":
				f.items[0].TerminalState = session.TerminalExited
			case "unknown":
				f.items[0].TerminalState = session.TerminalUnknown
			case "disconnected":
				f.err = errors.New("offline")
				f.items = nil
			case "stale":
				r.Revision = 2
			case "unconfirmed":
				r.ConfirmEnded = false
			}
			if _, e := manager.New(f, db).Recover(ctx, r); e == nil {
				t.Fatal("unsafe recovery allowed")
			}
			if f.starts != 0 {
				t.Fatal("unsafe recovery launched terminal")
			}
		})
	}
}
