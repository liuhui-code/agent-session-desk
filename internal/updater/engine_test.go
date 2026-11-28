package updater

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func setup(t *testing.T) (Engine, Contract, []byte) {
	t.Helper()
	dir := t.TempDir()
	old := []byte("old")
	product := []byte("new")
	c := testContract()
	c.SHA256 = sum(product)
	c.Size = int64(len(product))
	os.MkdirAll(filepath.Join(dir, "terminal-runtime"), 0700)
	os.MkdirAll(filepath.Join(dir, "data", "state"), 0700)
	for name := range c.Compatibility.Runtime {
		data := []byte(name)
		c.Compatibility.Runtime[name] = sum(data)
		os.WriteFile(filepath.Join(dir, "terminal-runtime", name), data, 0600)
	}
	os.WriteFile(filepath.Join(dir, "old.exe"), old, 0600)
	os.WriteFile(filepath.Join(dir, "data", "state", "metadata.sqlite3"), []byte("private metadata"), 0600)
	a, _ := json.Marshal(map[string]any{"entrypoint": "old.exe", "namespace": "asd-update-test", "files": map[string]string{"old.exe": sum(old)}})
	os.WriteFile(filepath.Join(dir, "alpha.json"), a, 0600)
	e := Engine{Root: dir, Offline: func(_ context.Context, f func() error) error { return f() }, Probe: func(context.Context, string) (Probe, error) {
		return Probe{Protocol: 1, Platform: c.Platform, Revision: c.Revision, Compatibility: c.Compatibility}, nil
	}, Check: func(context.Context, string) (int, error) { return 11, nil }, Backup: func(_ context.Context, src, dst string) error {
		b, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, b, 0600)
	}}
	return e, c, product
}
func TestBackupFailureNeverChangesProgramOrConfiguration(t *testing.T) {
	e, c, p := setup(t)
	ctx := context.Background()
	if err := e.Stage(ctx, c, bytes.NewReader(p)); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(e.Root, "alpha.json"))
	e.Backup = func(context.Context, string, string) error { return errors.New("disk full") }
	if _, err := e.Apply(ctx); err == nil {
		t.Fatal("backup failure accepted")
	}
	after, _ := os.ReadFile(filepath.Join(e.Root, "alpha.json"))
	if string(after) != string(before) {
		t.Fatal("entrypoint changed before backup")
	}
}

func TestSwitchAndRollbackKeepCurrentMetadataAndOnePreviousProgram(t *testing.T) {
	e, c, p := setup(t)
	ctx := context.Background()
	if err := e.Stage(ctx, c, bytes.NewReader(p)); err != nil {
		t.Fatal(err)
	}
	r, err := e.Apply(ctx)
	if err != nil || r.Current == "old.exe" || r.Previous != "old.exe" {
		t.Fatal(r, err)
	}
	// Work done after update must survive a program rollback.
	file := filepath.Join(e.Root, "data", "state", "metadata.sqlite3")
	os.WriteFile(file, []byte("new private work"), 0600)
	r, err = e.Rollback(ctx)
	if err != nil || r.Current != "old.exe" {
		t.Fatal(r, err)
	}
	data, _ := os.ReadFile(file)
	if string(data) != "new private work" {
		t.Fatal("rollback restored stale metadata")
	}
}

func TestInterruptedPreparationRequiresExplicitRepairWithoutReplay(t *testing.T) {
	e, c, p := setup(t)
	ctx := context.Background()
	e.Backup = func(context.Context, string, string) error { return errors.New("power loss") }
	if err := e.Stage(ctx, c, bytes.NewReader(p)); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Apply(ctx); err == nil {
		t.Fatal("injected backup succeeded")
	}
	if s, err := e.Status(ctx); err != nil || !s.RepairRequired {
		t.Fatal(s, err)
	}
	if _, err := e.Apply(ctx); err == nil {
		t.Fatal("interrupted update replayed")
	}
	r, err := e.Repair(ctx)
	if err != nil || r.Current != "old.exe" {
		t.Fatal(r, err)
	}
	s, err := e.Status(ctx)
	if err != nil || s.RepairRequired || !s.Pending {
		t.Fatal(s, err)
	}
	if err = e.Discard(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestRepeatedUpdatesKeepThreeProgramSlotsAndPreviousRunnable(t *testing.T) {
	e, c, _ := setup(t)
	ctx := context.Background()
	for i := 0; i < 7; i++ {
		p := []byte(fmt.Sprintf("program-%d", i))
		c.Size = int64(len(p))
		c.SHA256 = sum(p)
		if err := e.Stage(ctx, c, bytes.NewReader(p)); err != nil {
			t.Fatal(i, err)
		}
		if _, err := e.Apply(ctx); err != nil {
			t.Fatal(i, err)
		}
		s, err := e.Status(ctx)
		if err != nil || s.Current == s.Previous || s.RepairRequired {
			t.Fatal(s, err)
		}
	}
	entries, _ := os.ReadDir(e.Root)
	managed := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "asd-update-") && strings.HasSuffix(entry.Name(), ".exe") {
			managed++
		}
	}
	if managed != 3 {
		t.Fatal("unbounded program slots", managed)
	}
	if _, err := e.Rollback(ctx); err != nil {
		t.Fatal("retained previous not runnable", err)
	}
}

func TestCommittedIntentRepairDoesNotReplayAProgram(t *testing.T) {
	e, c, p := setup(t)
	ctx := context.Background()
	before, _ := os.ReadFile(filepath.Join(e.Root, "alpha.json"))
	if err := e.Stage(ctx, c, bytes.NewReader(p)); err != nil {
		t.Fatal(err)
	}
	r, err := e.Apply(ctx)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(e.Root, "alpha.json"))
	b, _ := json.Marshal(intent{1, sum(before), sum(after), r.Current, c.SHA256})
	os.WriteFile(filepath.Join(e.Root, work, "intent.json"), b, 0600)
	e.Probe = func(context.Context, string) (Probe, error) { t.Fatal("repair replayed program"); return Probe{}, nil }
	repaired, err := e.Repair(ctx)
	if err != nil || repaired.Action != "finalized_switch" || repaired.Current != r.Current {
		t.Fatal(repaired, err)
	}
}

func TestEmptyInstallationDoesNotClaimANonexistentBackup(t *testing.T) {
	e, c, p := setup(t)
	os.Remove(filepath.Join(e.Root, "data", "state", "metadata.sqlite3"))
	ctx := context.Background()
	if err := e.Stage(ctx, c, bytes.NewReader(p)); err != nil {
		t.Fatal(err)
	}
	r, err := e.Apply(ctx)
	if err != nil || r.Backup != "" {
		t.Fatal("nonexistent backup reported", r, err)
	}
}

func TestCaseAliasCannotTurnCurrentProgramIntoAnInactiveSlot(t *testing.T) {
	e, c, p := setup(t)
	old, _ := os.ReadFile(filepath.Join(e.Root, "old.exe"))
	alias := "ASD-UPDATE-1.exe"
	os.Rename(filepath.Join(e.Root, "old.exe"), filepath.Join(e.Root, alias))
	os.WriteFile(filepath.Join(e.Root, "previous.exe"), old, 0600)
	state := State{Protocol: 1, Previous: Record{Entry: "previous.exe", SHA256: sum(old), Compatibility: c.Compatibility, Revision: c.Revision}, Owned: map[string]string{"asd-update-1.exe": sum(old)}}
	a, _ := json.Marshal(map[string]any{"entrypoint": alias, "namespace": "asd-update-test", "files": map[string]string{alias: sum(old)}, "asd_update": state})
	os.WriteFile(filepath.Join(e.Root, "alpha.json"), a, 0600)
	if err := e.Stage(context.Background(), c, bytes.NewReader(p)); err == nil {
		t.Fatal("case alias treated as inactive program")
	}
}

type blockedProduct struct {
	entered, release chan struct{}
	once             sync.Once
	reader           *bytes.Reader
}

func (r *blockedProduct) Read(p []byte) (int, error) {
	r.once.Do(func() { close(r.entered) })
	<-r.release
	return r.reader.Read(p)
}
func TestConcurrentUpdateRefusesInsteadOfOverwritingStaging(t *testing.T) {
	e, c, p := setup(t)
	ctx := context.Background()
	reader := &blockedProduct{entered: make(chan struct{}), release: make(chan struct{}), reader: bytes.NewReader(p)}
	done := make(chan error, 1)
	go func() { done <- e.Stage(ctx, c, reader) }()
	<-reader.entered
	if _, err := e.Status(ctx); !errors.Is(err, ErrBusy) {
		close(reader.release)
		<-done
		t.Fatal("concurrent operation did not refuse", err)
	}
	close(reader.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestSwitchedManifestRetainsPowerShellUTF8Encoding(t *testing.T) {
	e, c, p := setup(t)
	path := filepath.Join(e.Root, "alpha.json")
	b, _ := os.ReadFile(path)
	var doc map[string]any
	json.Unmarshal(b, &doc)
	doc["localized_name"] = "会话"
	b, _ = json.Marshal(doc)
	os.WriteFile(path, append([]byte{0xef, 0xbb, 0xbf}, b...), 0600)
	if err := e.Stage(context.Background(), c, bytes.NewReader(p)); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Apply(context.Background()); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if !bytes.HasPrefix(after, []byte{0xef, 0xbb, 0xbf}) || !bytes.Contains(after, []byte("会话")) {
		t.Fatal("PowerShell 5.1 manifest encoding lost")
	}
}

func TestUnownedBackupSlotIsNeverOverwritten(t *testing.T) {
	for _, empty := range []bool{false, true} {
		t.Run(fmt.Sprintf("empty=%t", empty), func(t *testing.T) {
			e, c, p := setup(t)
			if empty {
				os.Remove(filepath.Join(e.Root, "data", "state", "metadata.sqlite3"))
			}
			ctx := context.Background()
			if err := e.Stage(ctx, c, bytes.NewReader(p)); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(e.Root, work, "backup-1.sqlite3")
			os.WriteFile(path, []byte("unowned backup"), 0600)
			if _, err := e.Apply(ctx); err == nil {
				t.Fatal("unowned backup overwritten or claimed")
			}
			b, _ := os.ReadFile(path)
			if string(b) != "unowned backup" {
				t.Fatal("backup modified")
			}
		})
	}
}

func TestRepairReleasesUncommittedProgramAndBackupSlots(t *testing.T) {
	e, c, p := setup(t)
	ctx := context.Background()
	originalBackup := e.Backup
	e.Backup = func(ctx context.Context, src, dst string) error {
		if err := originalBackup(ctx, src, dst); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(e.Root, "alpha.update.tmp"), []byte("injected publication collision"), 0600)
	}
	if err := e.Stage(ctx, c, bytes.NewReader(p)); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Apply(ctx); err == nil {
		t.Fatal("injected collision did not fail")
	}
	if _, err := e.Repair(ctx); err != nil {
		t.Fatal(err)
	}
	e.Backup = originalBackup
	if _, err := e.Apply(ctx); err != nil {
		t.Fatal("aborted preparation left occupied slots", err)
	}
}
