//go:build windows

package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/instance"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
	"github.com/liuhui-code/agent-session-desk/internal/updater"
	"golang.org/x/sys/windows"
)

type portableAssetTransport func(*http.Request) (*http.Response, error)

func (f portableAssetTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func updateCopy(t *testing.T, src, dst string) {
	t.Helper()
	in, e := os.Open(src)
	if e != nil {
		t.Fatal(e)
	}
	defer in.Close()
	out, e := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if e != nil {
		t.Fatal(e)
	}
	_, e = io.Copy(out, in)
	closeErr := out.Close()
	if e != nil || closeErr != nil {
		t.Fatal(e, closeErr)
	}
}
func updateHash(t *testing.T, path string) string {
	t.Helper()
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		t.Fatal(e)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func TestNativeUpdateCrashProcess(t *testing.T) {
	root := os.Getenv("ASD_UPDATE_CRASH_ROOT")
	if root == "" {
		t.Skip("child only")
	}
	e := newUpdateEngine(root, true)
	backup := e.Backup
	e.Backup = func(ctx context.Context, src, dst string) error {
		if err := backup(ctx, src, dst); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, "crash-ready.json"), []byte(`{"backup_complete":true}`), 0600); err != nil {
			return err
		}
		for {
			time.Sleep(time.Second)
		}
	}
	if _, err := e.Apply(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWindowsNativePortableUpdateRollbackAndKilledTransaction(t *testing.T) {
	bundle, product, candidate, fixture := os.Getenv("ASD_UPDATE_BUNDLE"), os.Getenv("ASD_UPDATE_PRODUCT"), os.Getenv("ASD_UPDATE_CANDIDATE"), os.Getenv("ASD_UPDATE_FIXTURE")
	if bundle == "" || product == "" || candidate == "" || fixture == "" {
		t.Skip("explicit isolated native update environment required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	root := t.TempDir()
	id, _ := session.NewID()
	namespace := "asd-upd-" + id[:12]
	os.Mkdir(filepath.Join(root, "terminal-runtime"), 0700)
	for name := range backend.TerminalBundleFingerprints() {
		updateCopy(t, filepath.Join(bundle, name), filepath.Join(root, "terminal-runtime", name))
	}
	updateCopy(t, product, filepath.Join(root, "old.exe"))
	oldHash := updateHash(t, product)
	newHash := updateHash(t, candidate)
	if newHash == oldHash {
		t.Fatal("independent candidate bytes required")
	}
	alpha, _ := json.Marshal(map[string]any{"entrypoint": "old.exe", "namespace": namespace, "files": map[string]string{"old.exe": oldHash}})
	os.WriteFile(filepath.Join(root, "alpha.json"), alpha, 0600)
	call := func(args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, product, append([]string{"update"}, args...)...)
		return cmd.CombinedOutput()
	}
	must := func(target any, args ...string) {
		t.Helper()
		b, e := call(args...)
		if e != nil {
			t.Fatalf("update %s: %v %s", args[0], e, b)
		}
		if target != nil && json.Unmarshal(b, target) != nil {
			t.Fatal("invalid update receipt", string(b))
		}
	}
	contractPath := filepath.Join(root, "asd-update.json")
	tag := "v0.0.0-native-update.1"
	must(nil, "package", "--program", candidate, "--tag", tag, "--output", contractPath)
	contractBytes, _ := os.ReadFile(contractPath)
	var contract updater.Contract
	if json.Unmarshal(contractBytes, &contract) != nil {
		t.Fatal("contract")
	}
	makeAsset := func(name, hash string, size int64) map[string]any {
		return map[string]any{"name": name, "state": "uploaded", "size": size, "digest": "sha256:" + hash, "browser_download_url": "https://github.com/liuhui-code/agent-session-desk/releases/download/" + tag + "/" + name}
	}
	meta, _ := json.Marshal(map[string]any{"tag_name": tag, "assets": []any{makeAsset("asd-update.json", updateHash(t, contractPath), int64(len(contractBytes))), makeAsset(contract.Product, contract.SHA256, contract.Size)}})
	client := updater.Client{HTTP: &http.Client{Transport: portableAssetTransport(func(r *http.Request) (*http.Response, error) {
		var body io.ReadCloser = io.NopCloser(strings.NewReader(string(meta)))
		if strings.HasSuffix(r.URL.Path, "asd-update.json") {
			body = io.NopCloser(strings.NewReader(string(contractBytes)))
		} else if strings.HasSuffix(r.URL.Path, ".exe") {
			f, e := os.Open(candidate)
			if e != nil {
				return nil, e
			}
			body = f
		}
		return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header), Request: r}, nil
	})}}
	opts := backend.Options{TerminalBundle: filepath.Join(root, "terminal-runtime"), RuntimeDir: filepath.Join(root, "data", "runtime"), Namespace: namespace}
	b, err := backend.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	kind, location, _ := backend.Location(opts)
	metadata := filepath.Join(root, "data", "state", "metadata.sqlite3")
	db, err := store.Open(ctx, metadata, strings.ToLower(kind+":"+location))
	if err != nil {
		t.Fatal(err)
	}
	svc := manager.New(b, db)
	report := filepath.Join(root, "fixture.jsonl")
	original, err := svc.Start(ctx, session.Launch{Name: "Update original", Executable: fixture, Dir: root, Env: []string{"ASD_FIXTURE_REPORT_PATH=" + report}})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	var originalProcess windows.Handle
	defer func() {
		c, x := context.WithTimeout(context.Background(), 15*time.Second)
		defer x()
		if e := svc.Stop(c, original); e != nil {
			t.Error("owned Run cleanup", e)
		}
		if originalProcess != 0 {
			if result, e := windows.WaitForSingleObject(originalProcess, 10000); e != nil || result != windows.WAIT_OBJECT_0 {
				t.Error("fixture not ended", result, e)
			}
			windows.CloseHandle(originalProcess)
		}
		db.Close()
	}()
	var started struct {
		PID uint32 `json:"pid"`
	}
	until := time.Now().Add(15 * time.Second)
	for time.Now().Before(until) {
		if data, e := os.ReadFile(report); e == nil {
			line, _, _ := strings.Cut(string(data), "\n")
			if json.Unmarshal([]byte(line), &started) == nil && started.PID > 0 {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	if started.PID == 0 {
		t.Fatal("independent fixture receipt missing")
	}
	originalProcess, err = windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, started.PID)
	if err != nil {
		t.Fatal(err)
	}
	var created, ended, kernel, user windows.Filetime
	if err = windows.GetProcessTimes(originalProcess, &created, &ended, &kernel, &user); err != nil {
		t.Fatal(err)
	}
	var image [32768]uint16
	n := uint32(len(image))
	if e := windows.QueryFullProcessImageName(originalProcess, 0, &image[0], &n); e != nil || !strings.EqualFold(windows.UTF16ToString(image[:n]), fixture) {
		t.Fatal("original fixture image mismatch", e)
	}
	engine := newUpdateEngine(root, true)
	if err = engine.Download(ctx, tag, client); err != nil {
		t.Fatal("native streamed staging", err)
	}
	key, err := instance.ScopeKey(kind, location, metadata)
	if err != nil {
		t.Fatal(err)
	}
	live, err := instance.Listen(ctx, key, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out, e := call("apply", "--install-dir", root, "--confirm-closed"); e == nil || !strings.Contains(string(out), "close the original workbench") {
		live.Close()
		t.Fatal("live UI update not refused", e, string(out))
	}
	live.Close()
	var receipt updater.Receipt
	must(&receipt, "apply", "--install-dir", root, "--confirm-closed")
	if receipt.Previous != "old.exe" || updateHash(t, filepath.Join(root, receipt.Current)) != newHash {
		t.Fatal("wrong switched program", receipt)
	}
	if check, e := store.Check(ctx, filepath.Join(root, ".asd-update", receipt.Backup)); e != nil || check.SchemaVersion != 11 {
		t.Fatal("real metadata backup", check, e)
	}
	if _, err = svc.Rename(ctx, original, "Work after update"); err != nil {
		t.Fatal(err)
	}
	must(&receipt, "rollback", "--install-dir", root, "--confirm-closed")
	if receipt.Current != "old.exe" {
		t.Fatal("wrong rollback", receipt)
	}
	rows, err := svc.List(ctx)
	if err != nil || len(rows) != 1 || rows[0].ID != original.ID || rows[0].RunID != original.RunID || rows[0].BackendID != original.BackendID || rows[0].Title != "Work after update" {
		t.Fatal("rollback lost current work or original Run", err)
	}
	// Crash only this newly created updater helper, after a real fixed SQLite
	// snapshot, before program/pointer publication. Never kill the original Run.
	if err = engine.Download(ctx, tag, client); err != nil {
		t.Fatal(err)
	}
	self, _ := os.Executable()
	crash := exec.Command(self, "-test.run=^TestNativeUpdateCrashProcess$", "-test.timeout=60s")
	crash.Env = append(os.Environ(), "ASD_UPDATE_CRASH_ROOT="+root)
	log, err := os.Create(filepath.Join(root, "crash.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	crash.Stdout = log
	crash.Stderr = log
	if err = crash.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = crash.Process.Kill(); _ = crash.Wait() }()
	until = time.Now().Add(20 * time.Second)
	ready := false
	for time.Now().Before(until) {
		if data, e := os.ReadFile(filepath.Join(root, "crash-ready.json")); e == nil && strings.Contains(string(data), "true") {
			ready = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !ready {
		t.Fatal("updater never reached the observed backup boundary")
	}
	if err = crash.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if crash.Wait() == nil {
		t.Fatal("expected intentionally killed updater")
	}
	if _, e := call("apply", "--install-dir", root, "--confirm-closed"); e == nil {
		t.Fatal("killed transaction replayed")
	}
	must(&receipt, "repair", "--install-dir", root, "--confirm-closed")
	if receipt.Action != "aborted_preparation" || receipt.Current != "old.exe" {
		t.Fatal("incorrect crash repair", receipt)
	}
	must(nil, "discard", "--install-dir", root)
	if state, e := windows.WaitForSingleObject(originalProcess, 0); e != nil || state != uint32(windows.WAIT_TIMEOUT) {
		t.Fatal("original fixture no longer alive", state, e)
	}
	var same windows.Filetime
	if e := windows.GetProcessTimes(originalProcess, &same, &ended, &kernel, &user); e != nil || same != created {
		t.Fatal("original fixture identity changed", e)
	}
	t.Log(fmt.Sprintf("original fixture pid=%d creation=%d retained; actual CLI package/probes, streamed mock-provider staging, live UI refusal, v11 backup, switch, current-data rollback, one deliberately killed updater, explicit repair, no terminal input", started.PID, created.Nanoseconds()))
}
