package manager_test

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
)

func faultMessageRequest() session.MessageRequest {
	v := observed()
	return session.MessageRequest{RequestID: strings.Repeat("a", 32), TargetSessionID: v.ID, TargetRunID: v.RunID, Body: "AB中🙂 e\u0301 literal $();\r\nsecond line\n"}
}

// This is a real killed worker/SQLite recovery test with a simulated transport.
// The sidecar is an independent input observer, never product acceptance data.
func TestMessageFaultProcessRecoveryNeverReplays(t *testing.T) {
	for _, point := range []string{"queued", "claimed", "partial", "written", "receipt"} {
		t.Run(point, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			root := t.TempDir()
			path := filepath.Join(root, "metadata.sqlite3")
			db, err := store.Open(ctx, path, "simulation")
			if err != nil {
				t.Fatal(err)
			}
			if err = db.Save(ctx, observed()); err != nil {
				t.Fatal(err)
			}
			db.Close()
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(ctx, exe, "-test.run=^TestMessageFaultWorkerProcess$")
			cmd.Env = append(os.Environ(), "ASD_MESSAGE_FAULT_POINT="+point, "ASD_MESSAGE_FAULT_ROOT="+root)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			defer stdin.Close()
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err = cmd.Start(); err != nil {
				t.Fatal(err)
			}
			line, readErr := bufio.NewReader(stdout).ReadString('\n')
			killErr := cmd.Process.Kill() // exact owned worker; never process-name cleanup
			waitErr := cmd.Wait()
			if readErr != nil || line != "FAULT_BARRIER\n" || killErr != nil || waitErr == nil {
				t.Fatalf("worker barrier/kill failed: read=%v kill=%v wait=%v stderr=%s", readErr, killErr, waitErr, &stderr)
			}
			db, err = store.Open(ctx, path, "simulation")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			want := "unknown"
			if point == "queued" {
				want = "queued"
			} else if point == "receipt" {
				want = "input_written"
			}
			q, err := db.Message(ctx, faultMessageRequest().RequestID)
			if err != nil || q.State != want {
				t.Fatalf("recovered state=%s want=%s error=%v", q.State, want, err)
			}
			inputPath := filepath.Join(root, "input.bin")
			before, err := os.ReadFile(inputPath)
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			wantInput := ""
			if point == "partial" {
				wantInput = faultMessageRequest().Body[:4] // deliberately split UTF-8
			} else if point == "written" || point == "receipt" {
				wantInput = faultMessageRequest().Body
			}
			if string(before) != wantInput {
				t.Fatal("independent input observation does not match fault point")
			}
			calls := 0
			f := &headlessMessageTerminal{terminal: terminal{items: []session.Session{observed()}}}
			f.paste = func(context.Context, session.Session, string) error { calls++; return nil }
			m := manager.New(f, db)
			again, err := m.QueueMessage(ctx, faultMessageRequest())
			if err != nil || again.ID != q.ID || again.State != want {
				t.Fatal("sender restart lost original idempotency receipt")
			}
			if point != "queued" {
				if _, err = m.DeliverQueued(ctx, q.ID); err == nil || calls != 0 {
					t.Fatal("restarted worker replayed uncertain or written input")
				}
			} else if calls != 0 {
				t.Fatal("mere request recovery delivered queued input")
			}
			after, _ := os.ReadFile(inputPath)
			if !bytes.Equal(before, after) {
				t.Fatal("input sidecar changed during recovery")
			}
		})
	}
}

func TestMessageFaultWorkerProcess(t *testing.T) {
	point := os.Getenv("ASD_MESSAGE_FAULT_POINT")
	if point == "" {
		return
	}
	root := os.Getenv("ASD_MESSAGE_FAULT_ROOT")
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(root, "metadata.sqlite3"), "simulation")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	barrier := func() {
		fmt.Println("FAULT_BARRIER")
		_, _ = bufio.NewReader(os.Stdin).ReadByte()
		t.Fatal("fault worker resumed instead of being terminated")
	}
	f := &headlessMessageTerminal{terminal: terminal{items: []session.Session{observed()}}}
	f.paste = func(_ context.Context, _ session.Session, body string) error {
		if point == "claimed" {
			barrier()
		}
		if point == "partial" {
			body = body[:4]
		}
		file, e := os.OpenFile(filepath.Join(root, "input.bin"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return e
		}
		_, e = file.WriteString(body)
		if e == nil {
			e = file.Sync()
		}
		file.Close()
		if e != nil {
			return e
		}
		if point == "partial" || point == "written" {
			barrier()
		}
		return nil
	}
	m := manager.New(f, db)
	q, err := m.QueueMessage(ctx, faultMessageRequest())
	if err != nil {
		t.Fatal(err)
	}
	if point == "queued" {
		barrier()
	}
	if _, err = m.DeliverQueued(ctx, q.ID); err != nil {
		t.Fatal(err)
	}
	barrier()
}

func TestMessageFaultSixteenConsumersWriteOnce(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	var databases []*store.Store
	for i := 0; i < 16; i++ {
		db, err := store.Open(ctx, path, "simulation")
		if err != nil {
			t.Fatal(err)
		}
		databases = append(databases, db)
		defer db.Close()
	}
	var writes, successes atomic.Int32
	f := &headlessMessageTerminal{terminal: terminal{items: []session.Session{observed()}}}
	f.paste = func(_ context.Context, _ session.Session, body string) error {
		writes.Add(1)
		q, err := databases[0].Message(ctx, faultMessageRequest().RequestID)
		if err != nil || q.State != "unknown" || body != faultMessageRequest().Body {
			return fmt.Errorf("write without durable claim or changed payload")
		}
		return nil
	}
	m := manager.New(f, databases[0])
	q, err := m.QueueMessage(ctx, faultMessageRequest())
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, db := range databases {
		wg.Add(1)
		go func(db *store.Store) {
			defer wg.Done()
			<-start
			if _, err := manager.New(f, db).DeliverQueued(ctx, q.ID); err == nil {
				successes.Add(1)
			}
		}(db)
	}
	close(start)
	wg.Wait()
	receipt, err := m.Message(ctx, q.ID)
	if err != nil || writes.Load() != 1 || successes.Load() != 1 || receipt.State != "input_written" {
		t.Fatalf("writes=%d successes=%d state=%s err=%v", writes.Load(), successes.Load(), receipt.State, err)
	}
}

func TestMessageFaultBusyDatabaseNeverWritesInput(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metadata.sqlite3")
	db, err := store.Open(ctx, path, "simulation")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	writes := 0
	f := &headlessMessageTerminal{terminal: terminal{items: []session.Session{observed()}}}
	f.paste = func(context.Context, session.Session, string) error { writes++; return nil }
	m := manager.New(f, db)
	q, err := m.QueueMessage(ctx, faultMessageRequest())
	if err != nil {
		t.Fatal(err)
	}
	locker, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer locker.Close()
	locker.SetMaxOpenConns(1)
	if _, err = locker.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	defer locker.ExecContext(ctx, "ROLLBACK")
	deadline, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	if _, err = m.DeliverQueued(deadline, q.ID); err == nil || writes != 0 {
		t.Fatal("input written without a durable claim under database contention")
	}
	if _, err = locker.ExecContext(ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	got, err := m.Message(ctx, q.ID)
	if err != nil || got.State != "queued" {
		t.Fatalf("busy preflight consumed request: %+v %v", got, err)
	}
}
