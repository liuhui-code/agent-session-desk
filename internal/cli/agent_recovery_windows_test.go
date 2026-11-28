//go:build windows

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
	"golang.org/x/sys/windows"
)

type recoveryFixtureReceipt struct {
	PID                    uint32
	ConversationID, Marker string
}
type recoveryFixtureHistory struct{ ConversationID, Marker string }

// Independent native process with persisted synthetic history; no vendor Agent.
func TestRecoveryNativeFixtureProcess(t *testing.T) {
	file := os.Getenv("ASD_RESUME_FIXTURE_HISTORY")
	if file == "" {
		t.Skip("child only")
	}
	var h recoveryFixtureHistory
	contextPath := os.Getenv("ASD_RECOVERY_CONTEXT_FILE")
	if contextPath == "" {
		h = recoveryFixtureHistory{os.Getenv("ASD_RESUME_FIXTURE_ID"), os.Getenv("ASD_RESUME_FIXTURE_MARKER")}
		b, _ := json.Marshal(h)
		if err := os.WriteFile(file, b, 0600); err != nil {
			t.Fatal(err)
		}
	} else {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, &h); err != nil {
			t.Fatal(err)
		}
		b, err = os.ReadFile(contextPath)
		if err != nil {
			t.Fatal(err)
		}
		var c recoveryContext
		if err = json.Unmarshal(b, &c); err != nil {
			t.Fatal(err)
		}
		if c.ConversationID != h.ConversationID || h.Marker == "" {
			t.Fatal("wrong native history")
		}
		r := recovery.Result{Protocol: "asd.agent-recovery-result", SchemaVersion: 1, SessionID: c.SessionID, RunID: c.RunID, AttemptID: c.AttemptID, ConversationID: h.ConversationID, Status: "restored"}
		b, _ = json.Marshal(r)
		if ok, err := publishPrivateDocument(c.ResultFile, b); err != nil || !ok {
			t.Fatal("result publication", err)
		}
	}
	b, _ := json.Marshal(recoveryFixtureReceipt{uint32(os.Getpid()), h.ConversationID, h.Marker})
	if ok, err := publishPrivateDocument(os.Getenv("ASD_RESUME_FIXTURE_REPORT"), b); err != nil || !ok {
		t.Fatal("receipt publication", err)
	}
	for {
		time.Sleep(time.Second)
	}
}

func waitRecoveryJSON(t *testing.T, path string, target any) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if b, e := os.ReadFile(path); e == nil && json.Unmarshal(b, target) == nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("timed out waiting for synthetic JSON", filepath.Base(path))
}

func recoveryProcess(t *testing.T, pid uint32, image string) windows.Handle {
	t.Helper()
	h, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		t.Fatal(err)
	}
	var buf [32768]uint16
	n := uint32(len(buf))
	if err = windows.QueryFullProcessImageName(h, 0, &buf[0], &n); err != nil || !strings.EqualFold(windows.UTF16ToString(buf[:n]), image) {
		windows.CloseHandle(h)
		t.Fatal("fixture image identity", err)
	}
	var created, ended, kernel, user windows.Filetime
	if err = windows.GetProcessTimes(h, &created, &ended, &kernel, &user); err != nil {
		windows.CloseHandle(h)
		t.Fatal(err)
	}
	t.Logf("fixture pid=%d creation=%d", pid, created.Nanoseconds())
	return h
}

func TestWindowsNativeRecoveryCLI(t *testing.T)      { testWindowsNativeRecovery(t, false) }
func TestWindowsNativeRecoveryUIBridge(t *testing.T) { testWindowsNativeRecovery(t, true) }
func testWindowsNativeRecovery(t *testing.T, useUI bool) {
	bundle, product := os.Getenv("ASD_RECOVERY_NATIVE_BUNDLE"), os.Getenv("ASD_RECOVERY_PRODUCT")
	if bundle == "" || product == "" {
		t.Skip("explicit isolated native recovery environment required")
	}
	root := t.TempDir()
	id, _ := session.NewID()
	self, _ := os.Executable()
	base := []string{"--terminal-bundle", bundle, "--runtime-dir", filepath.Join(root, "runtime"), "--namespace", "asd-rec-" + id[:12], "--state-dir", filepath.Join(root, "state")}
	call := func(args ...string) ([]byte, error) {
		cmd := exec.Command(product, append(append([]string{}, base...), args...)...)
		return cmd.CombinedOutput()
	}
	must := func(target any, args ...string) {
		t.Helper()
		b, e := call(args...)
		if e != nil {
			t.Fatalf("%s: %v %s", args[0], e, b)
		}
		if target != nil {
			if e = json.Unmarshal(b, target); e != nil {
				t.Fatal("CLI JSON", e, string(b))
			}
		}
	}
	history := filepath.Join(root, "native-history.json")
	report1 := filepath.Join(root, "original.json")
	report2 := filepath.Join(root, "loaded.json")
	nativeID := "native-" + id
	marker := "persisted-" + id
	var original session.Session
	must(&original, "start", "--name", "Recovery fixture", "--cwd", root, "--env", "ASD_RESUME_FIXTURE_HISTORY="+history, "--env", "ASD_RESUME_FIXTURE_ID="+nativeID, "--env", "ASD_RESUME_FIXTURE_MARKER="+marker, "--env", "ASD_RESUME_FIXTURE_REPORT="+report1, "--", self, "-test.run=^TestRecoveryNativeFixtureProcess$")
	cleanupNeeded := true
	// Cleanup targets only this newly allocated logical Session in its namespace.
	defer func() {
		if !cleanupNeeded {
			return
		}
		if b, e := call("stop", original.ID); e != nil {
			t.Logf("cleanup: %v %s", e, b)
		}
	}()
	var first recoveryFixtureReceipt
	waitRecoveryJSON(t, report1, &first)
	old := recoveryProcess(t, first.PID, self)
	defer windows.CloseHandle(old)
	token := filepath.Join(root, "source-credential.json")
	must(nil, "agent-bind", "--session", original.ID, "--run", original.RunID, "--token-file", token)
	d := recovery.Declaration{Protocol: "asd.agent-recovery", SchemaVersion: 1, SessionID: original.ID, RunID: original.RunID, AgentID: "synthetic-agent", ConversationID: nativeID, Resume: recovery.Resume{Executable: self, Args: []string{"-test.run=^TestRecoveryNativeFixtureProcess$"}, Cwd: root, Env: map[string]string{"ASD_RESUME_FIXTURE_HISTORY": history, "ASD_RESUME_FIXTURE_REPORT": report2}}}
	declaration := filepath.Join(root, "recovery.json")
	b, _ := json.Marshal(d)
	if e := os.WriteFile(declaration, b, 0600); e != nil {
		t.Fatal(e)
	}
	var profile recovery.Profile
	must(&profile, "recovery-register", "--token-file", token, "--file", declaration, "--revision", "0")
	must(&profile, "recovery-profile", "--session", original.ID)
	if b, e := call("recover", "--session", original.ID, "--previous-run", original.RunID, "--revision", "1", "--confirm-ended", "--context-dir", filepath.Join(root, "refused")); e == nil {
		t.Fatalf("live terminal duplicated: %s", b)
	}
	must(nil, "stop", original.ID)
	if state, e := windows.WaitForSingleObject(old, 10000); e != nil || state != windows.WAIT_OBJECT_0 {
		t.Fatal("original fixture still alive", state, e)
	}
	dir := filepath.Join(root, "attempt")
	var attempt recovery.Attempt
	if !useUI {
		must(&attempt, "recover", "--session", original.ID, "--previous-run", original.RunID, "--revision", "1", "--confirm-ended", "--context-dir", dir)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		opts := backend.Options{TerminalBundle: bundle, RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-rec-" + id[:12]}
		b, err := backend.New(opts)
		if err != nil {
			t.Fatal(err)
		}
		path, scope, err := metadataPath(opts, filepath.Join(root, "state"))
		if err != nil {
			t.Fatal(err)
		}
		db, err := store.Open(ctx, path, scope)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		controls := withRecoveryControls(ui.Options{}, manager.New(b, db), opts, filepath.Join(root, "state"))
		review, err := controls.ReviewRecovery(ctx, original)
		if err != nil || review.Profile == nil {
			t.Fatal("UI review", err)
		}
		stale := review
		oldProfile := *review.Profile
		oldProfile.Revision++
		stale.Profile = &oldProfile
		rejected, err := controls.RestoreRecovery(ctx, stale)
		if err == nil || rejected.Attempted {
			t.Fatal("stale UI review launched")
		}
		outcome, err := controls.RestoreRecovery(ctx, review)
		if err != nil {
			t.Fatal("UI restore", err)
		}
		attempt = outcome.Attempt
		if !outcome.Attempted || outcome.Session.ID != original.ID || outcome.Session.RunID != attempt.RunID || outcome.Session.RunID == original.RunID {
			t.Fatal("UI launch identity")
		}
		dir = filepath.Join(root, "state", "recovery-context-01")
	}
	if attempt.State != "started" || attempt.RunID == original.RunID || attempt.SessionID != original.ID {
		t.Fatalf("attempt %#v", attempt)
	}
	var second recoveryFixtureReceipt
	waitRecoveryJSON(t, report2, &second)
	current := recoveryProcess(t, second.PID, self)
	defer windows.CloseHandle(current)
	if second.ConversationID != nativeID || second.Marker != marker || second.PID == first.PID {
		t.Fatalf("native history was not recovered: %#v", second)
	}
	result := filepath.Join(dir, "result.json")
	var resultDoc recovery.Result
	waitRecoveryJSON(t, result, &resultDoc)
	credential := filepath.Join(dir, "agent-credential.json")
	var beforeResult recovery.Attempt
	must(&beforeResult, "recovery-status", "--session", original.ID)
	if beforeResult.State != "started" {
		t.Fatal("process launch inferred restored")
	}
	var contextDoc recoveryContext
	waitRecoveryJSON(t, filepath.Join(dir, "context.json"), &contextDoc)
	must(nil, "recovery-result", "--token-file", contextDoc.CredentialFile, "--file", contextDoc.ResultFile)
	var accepted recovery.Attempt
	must(&accepted, "recovery-status", "--session", original.ID)
	if accepted.State != "restored" || accepted.ID != attempt.ID {
		t.Fatal("authenticated recovery result not accepted")
	}
	must(&attempt, "recovery-result", "--token-file", credential, "--file", result)
	if attempt.State != "restored" {
		t.Fatal("unconfirmed restore", attempt.State)
	}
	must(nil, "recovery-result", "--token-file", credential, "--file", result)
	if b, e := call("recovery-result", "--token-file", token, "--file", result); e == nil {
		t.Fatalf("old credential accepted: %s", b)
	}
	must(&attempt, "recovery-status", "--session", original.ID)
	if b, e := call("recover", "--session", original.ID, "--previous-run", original.RunID, "--revision", "1", "--confirm-ended", "--context-dir", filepath.Join(root, "duplicate")); e == nil {
		t.Fatalf("stale restore repeated: %s", b)
	}
	var live []session.Session
	must(&live, "list", "--json")
	if len(live) != 1 || live[0].ID != original.ID || live[0].RunID != attempt.RunID {
		t.Fatalf("wrong discovered target %#v", live)
	}
	var records []session.HistoryRecord
	must(&records, "history")
	if len(records) != 2 || records[0].Session.ID != records[1].Session.ID {
		t.Fatalf("history identities %#v", records)
	}
	// Verify metadata backup includes accepted profile and result, without context secrets.
	backup := filepath.Join(root, "backup.sqlite3")
	must(nil, "metadata", "backup", backup)
	data, e := os.ReadFile(backup)
	if e != nil {
		t.Fatal(e)
	}
	var secret credentialDocument
	waitRecoveryJSON(t, credential, &secret)
	if bytes.Contains(data, []byte(secret.Token)) {
		t.Fatal("credential secret in metadata backup")
	}
	must(nil, "stop", original.ID)
	if state, e := windows.WaitForSingleObject(current, 10000); e != nil || state != windows.WAIT_OBJECT_0 {
		t.Fatal("restored fixture still alive", state, e)
	}
	must(&live, "list", "--json")
	if len(live) != 0 {
		t.Fatalf("owned namespace retained backend targets: %#v", live)
	}
	t.Logf("session=%s old_run=%s new_run=%s namespace=asd-rec-%s", original.ID, original.RunID, attempt.RunID, id[:12])
	cleanupNeeded = false
	t.Log(fmt.Sprintf("same Session, two exact native processes, persisted marker, new credential/result, backup; attempt=%s", attempt.ID))
}
