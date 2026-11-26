package manager_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/cli"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
	terminalTransport "github.com/liuhui-code/agent-session-desk/internal/terminal"
)

// Opt-in functional slice only. No desktop/IME/soak or real-agent claim.
func TestWindowsNativeMessagePasteUsesOriginalRunWithoutEnter(t *testing.T) {
	bundle, shell := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_MESSAGING_SHELL")
	if bundle == "" || shell == "" {
		t.Skip("explicit pinned bundle and shell required")
	}
	root := t.TempDir()
	id, e := session.NewID()
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	opts := backend.Options{TerminalBundle: bundle, RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-msg-" + id[:12]}
	b, e := backend.New(opts)
	if e != nil {
		t.Fatal(e)
	}
	kind, location, e := backend.Location(opts)
	if e != nil {
		t.Fatal(e)
	}
	db, e := store.Open(ctx, filepath.Join(root, "state", "metadata.sqlite3"), strings.ToLower(kind+":"+location))
	if e != nil {
		t.Fatal(e)
	}
	m := manager.New(b, db)
	globals := []string{"--terminal-bundle", bundle, "--runtime-dir", opts.RuntimeDir, "--namespace", opts.Namespace, "--state-dir", filepath.Join(root, "state")}
	call := func(args ...string) []byte {
		var out, stderr bytes.Buffer
		if code := cli.Run(ctx, append(append([]string(nil), globals...), args...), nil, &out, &stderr); code != 0 {
			t.Fatalf("native CLI failed: %s", &stderr)
		}
		return out.Bytes()
	}

	var originals []session.Session
	defer func() {
		for _, v := range originals {
			c, stop := context.WithTimeout(context.Background(), 10*time.Second)
			if e := m.Stop(c, v); e != nil {
				t.Errorf("original cleanup failed: %v", e)
			}
			stop()
		}
		db.Close()
	}()
	marker := "SOURCE_" + id + "_\u4e2d\u6587"
	start := func(name, output string) session.Session {
		v, e := m.Start(ctx, session.Launch{Name: name, Executable: shell, Args: []string{"-NoLogo", "-NoProfile", "-NoExit", "-Command", "Write-Output '" + output + "'"}, Dir: root})
		if e != nil {
			t.Fatal(e)
		}
		originals = append(originals, v)
		return v
	}
	a, c := start("Message source", marker), start("Message target", "TARGET_READY_"+id)
	var record session.OutputRecord
	for i := 0; i < 40; i++ {
		e = json.Unmarshal(call("read", "--lines", "60", a.ID), &record)
		if e == nil && strings.Contains(record.Text, marker) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if e != nil {
		t.Fatal(e)
	}
	sourceLine := 0
	for i, line := range strings.Split(record.Text, "\n") {
		if strings.TrimSpace(line) == marker {
			sourceLine = i + 1
			break
		}
	}
	if sourceLine == 0 {
		t.Fatal("independent source marker not captured")
	}
	conn, e := m.OpenTerminal(ctx, c, 300, 24)
	if e != nil {
		t.Fatal(e)
	}
	defer conn.Close()
	screen := func() terminalTransport.Screen {
		for i := 0; i < 40; i++ {
			if e := conn.Refresh(ctx); e != nil {
				t.Fatal(e)
			}
			s, e := conn.Next(ctx)
			if e != nil {
				t.Fatal(e)
			}
			if s.Cursor.Visible && strings.Contains(s.Render(300, 24), "PS ") {
				return s
			}
		}
		t.Fatal("target prompt not observed")
		return terminalTransport.Screen{}
	}
	before := screen()
	fileID, e := session.NewID()
	if e != nil {
		t.Fatal(e)
	}
	bodyPath := filepath.Join(root, "message.txt")
	body := "literal $() ` ;\n\u4e2d\u6587"
	if e = os.WriteFile(bodyPath, append([]byte{0xef, 0xbb, 0xbf}, []byte(body)...), 0600); e != nil {
		t.Fatal(e)
	}
	var fileMessage session.Message
	if e = json.Unmarshal(call("send", "--request-id", fileID, "--to", c.ID, "--run", c.RunID, "--body-file", bodyPath), &fileMessage); e != nil || fileMessage.Body != body || fileMessage.State != "queued" {
		t.Fatalf("UTF-8 BOM file altered: %+v %v", fileMessage, e)
	}

	var q session.Message
	if e = json.Unmarshal(call("send", "--request-id", id, "--to", c.ID, "--run", c.RunID, "--source-run", a.RunID, "--record", record.ID, "--start-line", fmt.Sprint(sourceLine), "--end-line", fmt.Sprint(sourceLine)), &q); e != nil {
		t.Fatal(e)
	}
	if q.State != "queued" {
		t.Fatal("CLI send wrote input instead of only creating intent")
	}

	result, e := m.DeliverMessage(ctx, q.ID, c, conn.Paste)
	if e != nil || result.State != "input_written" {
		t.Fatalf("native input receipt: %+v %v", result, e)
	}
	var after terminalTransport.Screen
	for i := 0; i < 40; i++ {
		after = screen()
		if strings.Contains(after.Render(300, 24), marker) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(after.Render(300, 24), marker) || after.Cursor.Y != before.Cursor.Y {
		t.Fatal("paste not retained as a draft on the original prompt; Enter may have been added")
	}
	var receipt session.Message
	if e = json.Unmarshal(call("status", q.ID), &receipt); e != nil || receipt.State != "input_written" {
		t.Fatalf("native CLI status: %+v %v", receipt, e)
	}
	writes := 0
	if _, e = m.DeliverMessage(ctx, q.ID, c, func(context.Context, string) error { writes++; return nil }); e == nil || writes != 0 {
		t.Fatal("duplicate native delivery was attempted")
	}
	rows, e := m.List(ctx)
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range originals {
		same := false
		for _, current := range rows {
			if current.ID == v.ID && current.RunID == v.RunID && current.BackendID == v.BackendID {
				same = true
			}
		}
		if !same {
			t.Fatal("original native Run identity changed")
		}
	}
	t.Log("two original native shell Runs retained; snapshot provenance, draft paste, no appended Enter, no duplicate callback; owned cleanup follows")
}
