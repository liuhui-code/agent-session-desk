package store_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func replyCredential(v session.Session) session.AgentCredential {
	return session.AgentCredential{SessionID: v.ID, RunID: v.RunID, Secret: strings.Repeat("9", 64)}
}

func TestReplyRejectsWrongSourceAndPreservesConflictingResults(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	b := sample()
	other := b
	other.ID = strings.Repeat("3", 32)
	other.RunID = strings.Repeat("4", 32)
	other.BackendID = "other"
	c := replyCredential(b)
	wrong := replyCredential(other)
	wrong.Secret = strings.Repeat("8", 64)
	for _, v := range []session.Session{b, other} {
		if err = db.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	if err = db.BindAgent(ctx, c, b); err != nil {
		t.Fatal(err)
	}
	if err = db.BindAgent(ctx, wrong, other); err != nil {
		t.Fatal(err)
	}
	msg := session.Message{ID: strings.Repeat("a", 32), Target: b, Body: "request", Fingerprint: strings.Repeat("f", 64)}
	if _, err = db.QueueMessage(ctx, msg); err != nil {
		t.Fatal(err)
	}
	r := session.ReplyRequest{ID: strings.Repeat("b", 32), ReplyTo: msg.ID, Body: "result"}
	bad := c
	bad.Secret = strings.Repeat("0", 64)
	for _, credential := range []session.AgentCredential{wrong, bad} {
		if _, err = db.RecordReply(ctx, credential, r); err == nil {
			t.Fatal("unbound or wrong Run source accepted")
		}
	}
	missing := r
	missing.ReplyTo = strings.Repeat("e", 32)
	if _, err = db.RecordReply(ctx, c, missing); err == nil {
		t.Fatal("missing reply_to guessed another request")
	}
	if _, err = db.RecordReply(ctx, c, r); err != nil {
		t.Fatal(err)
	}
	conflict := r
	conflict.Body = "changed"
	if _, err = db.RecordReply(ctx, c, conflict); err == nil {
		t.Fatal("same reply ID overwrote content")
	}
	conflict = r
	conflict.ID = strings.Repeat("c", 32)
	if _, err = db.RecordReply(ctx, c, conflict); err == nil {
		t.Fatal("second result replaced first")
	}
	rows, err := db.Replies(ctx, msg.ID)
	if err != nil || len(rows) != 1 || rows[0].Body != r.Body {
		t.Fatal("original result lost")
	}
	if err = db.MarkStopped(ctx, b); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Inbox(ctx, c); err == nil {
		t.Fatal("stopped credential remained active")
	}
	if _, err = db.RecordReply(ctx, c, r); err == nil {
		t.Fatal("stopped Run submitted even an idempotent reply")
	}
	rows, err = db.Replies(ctx, msg.ID)
	if err != nil || len(rows) != 1 {
		t.Fatal("offline result lost after Run stopped")
	}
}

func TestReplyConcurrentWritersDeduplicateOrRefuseReplacement(t *testing.T) {
	for _, sameID := range []bool{true, false} {
		t.Run(fmt.Sprint(sameID), func(t *testing.T) {
			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "metadata.sqlite3")
			var dbs []*store.Store
			for i := 0; i < 16; i++ {
				db, err := store.Open(ctx, path, "a")
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				dbs = append(dbs, db)
			}
			v := sample()
			c := replyCredential(v)
			if err := dbs[0].Save(ctx, v); err != nil {
				t.Fatal(err)
			}
			if err := dbs[0].BindAgent(ctx, c, v); err != nil {
				t.Fatal(err)
			}
			msg := session.Message{ID: strings.Repeat("a", 32), Target: v, Body: "request", Fingerprint: strings.Repeat("b", 64)}
			if _, err := dbs[0].QueueMessage(ctx, msg); err != nil {
				t.Fatal(err)
			}
			start := make(chan struct{})
			results := make(chan error, 16)
			var wg sync.WaitGroup
			for i, db := range dbs {
				wg.Add(1)
				go func(i int, db *store.Store) {
					defer wg.Done()
					<-start
					id := 1
					if !sameID {
						id = i + 1
					}
					_, err := db.RecordReply(ctx, c, session.ReplyRequest{ID: fmt.Sprintf("%032x", id), ReplyTo: msg.ID, Body: "result"})
					results <- err
				}(i, db)
			}
			close(start)
			wg.Wait()
			close(results)
			wins := 0
			for err := range results {
				if err == nil {
					wins++
				}
			}
			want := 1
			if sameID {
				want = 16
			}
			rows, err := dbs[0].Replies(ctx, msg.ID)
			if err != nil || len(rows) != 1 || wins != want {
				t.Fatalf("wins %d want %d rows %d err %v", wins, want, len(rows), err)
			}
		})
	}
}

func TestReplyCredentialScopeRotationAndPaneExitRejections(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := sample()
	c := replyCredential(v)
	if err = db.Save(ctx, v); err != nil {
		t.Fatal(err)
	}
	if err = db.BindAgent(ctx, c, v); err != nil {
		t.Fatal(err)
	}
	if err = db.BindAgent(ctx, c, v); err != nil {
		t.Fatal("same credential retry not idempotent")
	}
	changed := c
	changed.Secret = strings.Repeat("1", 64)
	if err = db.BindAgent(ctx, changed, v); err == nil {
		t.Fatal("implicit credential rotation allowed")
	}
	other, err := store.Open(ctx, path, "b")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if _, err = other.AgentIdentity(ctx, c); err == nil {
		t.Fatal("credential crossed scope")
	}
	if err = db.Save(ctx, exited()); err != nil {
		t.Fatal(err)
	}
	if _, err = db.AgentIdentity(ctx, c); err == nil {
		t.Fatal("observed dead pane remained authorised")
	}
}

func TestReplyBindingCapacityDoesNotRotateOrPrune(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var first session.AgentCredential
	var original session.Session
	for i := 0; i <= store.MaxAgentBindings; i++ {
		v := sample()
		v.ID = fmt.Sprintf("%032x", i*2+1)
		v.RunID = fmt.Sprintf("%032x", i*2+2)
		v.BackendID = fmt.Sprintf("sim_%d", i)
		if err = db.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
		c, e := session.NewAgentCredential(v.ID, v.RunID)
		if e != nil {
			t.Fatal(e)
		}
		err = db.BindAgent(ctx, c, v)
		if i < store.MaxAgentBindings && err != nil {
			t.Fatal(err)
		}
		if i == store.MaxAgentBindings && err == nil {
			t.Fatal("binding capacity exceeded")
		}
		if i == 0 {
			first, original = c, v
		}
	}
	if err = db.BindAgent(ctx, first, original); err != nil {
		t.Fatal("full capacity blocked original binding retry")
	}
	if _, err = db.AgentIdentity(ctx, first); err != nil {
		t.Fatal("original binding was pruned")
	}
}

func TestReplyBoundInboxRoundTripDoesNotAdvanceDelivery(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "metadata.sqlite3"), "a")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	target := sample()
	if err = db.Save(ctx, target); err != nil {
		t.Fatal(err)
	}
	c := replyCredential(target)
	if err = db.BindAgent(ctx, c, target); err != nil {
		t.Fatal(err)
	}
	msg := session.Message{ID: strings.Repeat("a", 32), Target: target, Body: "request 中文", Fingerprint: strings.Repeat("b", 64)}
	if _, err = db.QueueMessage(ctx, msg); err != nil {
		t.Fatal(err)
	}
	inbox, err := db.Inbox(ctx, c)
	if err != nil || len(inbox) != 1 || inbox[0].ID != msg.ID {
		t.Fatalf("bound inbox: %+v %v", inbox, err)
	}
	r := session.ReplyRequest{ID: strings.Repeat("c", 32), ReplyTo: msg.ID, Body: "result 中文🙂"}
	got, err := db.RecordReply(ctx, c, r)
	if err != nil || got.ReplyTo != msg.ID || got.Source.RunID != target.RunID || got.Body != r.Body || got.RecordedAt.IsZero() {
		t.Fatalf("reply: %+v %v", got, err)
	}
	again, err := db.RecordReply(ctx, c, r)
	if err != nil || again.ID != got.ID || !again.RecordedAt.Equal(got.RecordedAt) {
		t.Fatal("reply retry lost identity")
	}
	inbox, err = db.Inbox(ctx, c)
	if err != nil || len(inbox) != 0 {
		t.Fatal("answered request remains in pending inbox")
	}
	all, err := db.Replies(ctx, msg.ID)
	if err != nil || len(all) != 1 || all[0].ID != r.ID {
		t.Fatalf("reply query: %+v %v", all, err)
	}
	status, err := db.Message(ctx, msg.ID)
	if err != nil || status.State != "queued" {
		t.Fatal("reply forged terminal delivery or completion state")
	}
}
