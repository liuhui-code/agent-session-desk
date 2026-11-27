package store_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func recoveryDeclaration(t *testing.T, v session.Session) recovery.Declaration {
	t.Helper()
	d := recovery.Declaration{Protocol: "asd.agent-recovery", SchemaVersion: 1, SessionID: v.ID, RunID: v.RunID, AgentID: "synthetic-agent", ConversationID: "conversation-example-42", Resume: recovery.Resume{Executable: `C:\Tools\MyAgent\agent.exe`, Args: []string{"resume", "conversation-example-42"}, Cwd: `C:\Work\project`}}

	return d
}
func TestRecoveryProfileIsBoundRevisionedAndDurable(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, e := store.Open(ctx, path, "recovery")
	if e != nil {
		t.Fatal(e)
	}
	v := sample()
	if e = db.Save(ctx, v); e != nil {
		t.Fatal(e)
	}
	c, e := session.NewAgentCredential(v.ID, v.RunID)
	if e != nil {
		t.Fatal(e)
	}
	if e = db.BindAgent(ctx, c, v); e != nil {
		t.Fatal(e)
	}
	d := recoveryDeclaration(t, v)
	p, e := db.RegisterRecovery(ctx, c, d, 0)
	if e != nil || p.Revision != 1 {
		t.Fatalf("register: %#v %v", p, e)
	}
	// Exact retries are idempotent even after the first response is lost.
	retry, e := db.RegisterRecovery(ctx, c, d, 0)
	if e != nil || retry.Revision != 1 {
		t.Fatalf("retry: %#v %v", retry, e)
	}
	d.ConversationID = "another"
	d.Resume.Args[1] = "another"
	if _, e = db.RegisterRecovery(ctx, c, d, 0); e == nil {
		t.Fatal("stale replacement accepted")
	}
	if _, e = db.RegisterRecovery(ctx, c, d, 1); e != nil {
		t.Fatal(e)
	}
	if e = db.MarkStopped(ctx, v); e != nil {
		t.Fatal(e)
	}
	if _, e = db.RegisterRecovery(ctx, c, d, 2); e == nil {
		t.Fatal("old credential still writes")
	}
	db.Close()
	db, e = store.Open(ctx, path, "recovery")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	p, e = db.RecoveryProfile(ctx, v.ID)
	if e != nil || p.Revision != 2 || p.Declaration.ConversationID != "another" {
		t.Fatalf("persisted: %#v %v", p, e)
	}
	bytes, _ := json.Marshal(p)
	if string(bytes) == "" {
		t.Fatal("empty profile")
	}
	other, e := store.Open(ctx, path, "other")
	if e != nil {
		t.Fatal(e)
	}
	defer other.Close()
	if _, e = other.RecoveryProfile(ctx, v.ID); e == nil {
		t.Fatal("cross scope profile exposed")
	}
}

func TestRecoveryReservationFencesDuplicateStartsAndKeepsPerRunSnapshot(t *testing.T) {
	ctx := context.Background()
	db, e := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "r")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	v := sample()
	if e = db.Save(ctx, v); e != nil {
		t.Fatal(e)
	}
	c, _ := session.NewAgentCredential(v.ID, v.RunID)
	if e = db.BindAgent(ctx, c, v); e != nil {
		t.Fatal(e)
	}
	p, e := db.RegisterRecovery(ctx, c, recoveryDeclaration(t, v), 0)
	if e != nil {
		t.Fatal(e)
	}
	next := v
	next.RunID = "33333333333333333333333333333333"
	next.BackendID = "asd_" + next.ID + "_" + next.RunID
	next.Title = "recovery invocation"
	next.Command = p.Declaration.Resume.Executable
	nc, _ := session.NewAgentCredential(next.ID, next.RunID)
	r := recovery.Reservation{Attempt: recovery.Attempt{ID: "44444444444444444444444444444444", SessionID: v.ID, RunID: next.RunID, ConversationID: p.Declaration.ConversationID, Revision: p.Revision}, PreviousRunID: v.RunID, Target: next, Launch: p.Declaration.Launch(next.Title), Credential: nc}
	if _, e = db.ReserveRecovery(ctx, r); e != nil {
		t.Fatal(e)
	}
	if _, e = db.ReserveRecovery(ctx, r); e == nil {
		t.Fatal("duplicate launch reservation accepted")
	}
	if e = db.Save(ctx, next); e != nil {
		t.Fatal("restored discovery mismatch:", e)
	}
	if e = db.Save(ctx, v); e != nil {
		t.Fatal("old discovery mismatch:", e)
	}
	rows, e := db.History(ctx, 10)
	if e != nil || len(rows) != 2 || rows[0].Session.ID != v.ID || rows[0].Session.Command != next.Command || rows[1].Session.Command != v.Command {
		t.Fatalf("history: %#v %v", rows, e)
	}
	d := recoveryDeclaration(t, v)
	d.ConversationID = "late"
	if _, e = db.RegisterRecovery(ctx, c, d, p.Revision); e == nil {
		t.Fatal("late source overwrite")
	}
	result := recovery.Result{Protocol: "asd.agent-recovery-result", SchemaVersion: 1, SessionID: next.ID, RunID: next.RunID, AttemptID: r.Attempt.ID, ConversationID: p.Declaration.ConversationID, Status: "restored"}
	if _, e = db.RecordRecoveryResult(ctx, nc, result); e == nil {
		t.Fatal("success before launch receipt")
	}
	if e = db.RecoveryStarted(ctx, next); e != nil {
		t.Fatal(e)
	}
	if _, e = db.RecordRecoveryResult(ctx, c, result); e == nil {
		t.Fatal("old credential result")
	}
	if _, e = db.RecordRecoveryResult(ctx, nc, result); e != nil {
		t.Fatal(e)
	}
	if _, e = db.RecordRecoveryResult(ctx, nc, result); e != nil {
		t.Fatal("idempotent ack", e)
	}
	if e = db.RecoveryStarted(ctx, next); e != nil {
		t.Fatal("late start receipt", e)
	}
	if current, e := db.RecoveryAttempt(ctx, next.ID); e != nil || current.State != "restored" {
		t.Fatal("late start regressed final result", current, e)
	}
	result.Status = "failed"
	if _, e = db.RecordRecoveryResult(ctx, nc, result); e == nil {
		t.Fatal("conflicting final result")
	}
}

func TestRecoveryUnknownAttemptCannotBeConfirmedAway(t *testing.T) {
	ctx := context.Background()
	db, e := store.Open(ctx, filepath.Join(t.TempDir(), "db"), "r")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	v := sample()
	if e = db.Save(ctx, v); e != nil {
		t.Fatal(e)
	}
	c, _ := session.NewAgentCredential(v.ID, v.RunID)
	if e = db.BindAgent(ctx, c, v); e != nil {
		t.Fatal(e)
	}
	p, e := db.RegisterRecovery(ctx, c, recoveryDeclaration(t, v), 0)
	if e != nil {
		t.Fatal(e)
	}
	reserve := func(previous, rid, aid string) error {
		credential, _ := session.NewAgentCredential(v.ID, rid)
		l := p.Declaration.Launch(v.Title)
		target := session.Session{ID: v.ID, RunID: rid, BackendID: "asd_" + v.ID + "_" + rid, Title: l.Name, Command: l.Executable}
		_, err := db.ReserveRecovery(ctx, recovery.Reservation{Attempt: recovery.Attempt{ID: aid, SessionID: v.ID, RunID: rid, ConversationID: p.Declaration.ConversationID, Revision: p.Revision}, PreviousRunID: previous, Target: target, Launch: l, Credential: credential})
		return err
	}
	rid := "33333333333333333333333333333333"
	if e = reserve(v.RunID, rid, "44444444444444444444444444444444"); e != nil {
		t.Fatal(e)
	}
	if e = reserve(rid, "55555555555555555555555555555555", "66666666666666666666666666666666"); e == nil {
		t.Fatal("unknown in-flight launch bypassed by another ended confirmation")
	}
}

func TestRecoveryConcurrentReservationsChooseOneRun(t *testing.T) {
	ctx := context.Background()
	db, e := store.Open(ctx, filepath.Join(t.TempDir(), "db"), "a")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	v := sample()
	if e = db.Save(ctx, v); e != nil {
		t.Fatal(e)
	}
	c, _ := session.NewAgentCredential(v.ID, v.RunID)
	if e = db.BindAgent(ctx, c, v); e != nil {
		t.Fatal(e)
	}
	p, e := db.RegisterRecovery(ctx, c, recoveryDeclaration(t, v), 0)
	if e != nil {
		t.Fatal(e)
	}
	results := make(chan error, 12)
	begin := make(chan struct{})
	for i := 0; i < 12; i++ {
		go func() {
			rid, _ := session.NewID()
			aid, _ := session.NewID()
			credential, _ := session.NewAgentCredential(v.ID, rid)
			next := v
			next.RunID = rid
			next.BackendID = "asd_" + v.ID + "_" + rid
			next.Command = p.Declaration.Resume.Executable
			r := recovery.Reservation{Attempt: recovery.Attempt{ID: aid, SessionID: v.ID, RunID: rid, ConversationID: p.Declaration.ConversationID, Revision: p.Revision}, PreviousRunID: v.RunID, Target: next, Launch: p.Declaration.Launch(next.Title), Credential: credential}
			<-begin
			_, err := db.ReserveRecovery(ctx, r)
			results <- err
		}()
	}
	close(begin)
	success := 0
	for i := 0; i < 12; i++ {
		e := <-results
		if e == nil {
			success++
		} else if !strings.Contains(e.Error(), "latest Run changed") {
			t.Fatal("unexpected reservation failure", e)
		}
	}
	if success != 1 {
		t.Fatal("concurrent starts authorized", success)
	}
	rows, e := db.History(ctx, 100)
	if e != nil || len(rows) != 2 {
		t.Fatal("duplicate Run history", rows, e)
	}
}
