//go:build windows

package pet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/store"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
	"golang.org/x/sys/windows"
)

type observedWorkbench struct {
	model *ui.Model
	mu    sync.Mutex
	text  string
}

func (o *observedWorkbench) Init() tea.Cmd  { return o.model.Init() }
func (o *observedWorkbench) View() tea.View { return o.model.View() }
func (o *observedWorkbench) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	_, cmd := o.model.Update(msg)
	o.mu.Lock()
	o.text = ansi.Strip(o.model.View().Content)
	o.mu.Unlock()
	return o, cmd
}
func (o *observedWorkbench) contains(s string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return strings.Contains(o.text, s)
}

// Native psmux, two original fixture processes, real SQLite/event CLI, real
// Bubble Tea loop and native layered window. HWND messages are a controlled
// click proxy; they do not establish physical mouse or WT foreground behavior.
func TestWindowsLivePetExistingBlockingForm(t *testing.T) {
	bundle, fixture, product := os.Getenv("ASD_NATIVE_MESSAGING_BUNDLE"), os.Getenv("ASD_NATIVE_MESSAGING_FIXTURE"), os.Getenv("ASD_NATIVE_REPLY_CLI")
	if bundle == "" || fixture == "" || product == "" {
		t.Skip("explicit native bundle, fixture and CLI required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := t.TempDir()
	id, _ := session.NewID()
	opts := backend.Options{TerminalBundle: bundle, RuntimeDir: filepath.Join(root, "runtime"), Namespace: "asd-pet-" + id[:12]}
	b, err := backend.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	kind, location, err := backend.Location(opts)
	if err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(ctx, filepath.Join(root, "state", "metadata.sqlite3"), strings.ToLower(kind+":"+location))
	if err != nil {
		t.Fatal(err)
	}
	service := manager.New(b, db)
	var originals []session.Session
	defer func() {
		for _, original := range originals {
			stopCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
			if err := service.Stop(stopCtx, original); err != nil {
				t.Errorf("owned fixture cleanup: %v", err)
			}
			stop()
		}
		db.Close()
	}()
	wait := func(label string, check func() bool) {
		t.Helper()
		for !check() {
			select {
			case <-ctx.Done():
				t.Fatal("deadline: " + label)
			case <-time.After(30 * time.Millisecond):
			}
		}
	}
	config := filepath.Join(root, "form.json")
	start := func(name string, form bool) (session.Session, string) {
		journal := filepath.Join(root, name+".jsonl")
		env := []string{"ASD_FIXTURE_REPORT_PATH=" + journal}
		if form {
			env = append(env, "ASD_FIXTURE_ATTENTION_CONFIG="+config)
		}
		s, err := service.Start(ctx, session.Launch{Name: name, Executable: fixture, Dir: root, Env: env})
		if err != nil {
			t.Fatal(err)
		}
		originals = append(originals, s)
		wait("fixture started", func() bool { data, _ := os.ReadFile(journal); return bytes.Contains(data, []byte(`"event":"started"`)) })
		return s, journal
	}
	a, aLog := start("Current", false)
	requests := make(chan Activation, 1)
	companion := NewCompanion(func(a Activation) {
		select {
		case requests <- a:
		default:
		}
	})
	m := ui.NewWithOptions(ctx, service, ui.Options{AttentionObserver: companion.Publish})
	observed := &observedWorkbench{model: m}
	program := tea.NewProgram(observed, tea.WithInput(nil), tea.WithOutput(io.Discard), tea.WithoutRenderer(), tea.WithoutSignalHandler(), tea.WithWindowSize(120, 32), tea.WithContext(ctx))
	programDone := make(chan error, 1)
	go func() { _, err := program.Run(); programDone <- err }()
	defer func() {
		program.Quit()
		if err := <-programDone; err != nil {
			t.Errorf("workbench exit: %v", err)
		}
		m.Close()
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case a := <-requests:
				program.Send(ui.AttentionActivate{ID: a.ID, Version: a.Version})
			}
		}
	}()
	petCtx, stopPet := context.WithCancel(ctx)
	petDone := make(chan error, 1)
	go func() { petDone <- Run(petCtx, Options{Live: companion}) }()
	defer func() {
		stopPet()
		if err := <-petDone; err != nil {
			t.Errorf("native companion exit: %v", err)
		}
	}()
	find := user.NewProc("FindWindowExW")
	getText := user.NewProc("GetWindowTextW")
	send := user.NewProc("SendMessageW")
	var hwnd uintptr
	wait("live idle window and original A", func() bool {
		for after := uintptr(0); ; {
			candidate, _, _ := find.Call(0, after, uintptr(unsafe.Pointer(utf16("ASD.Pet.Demo"))), 0)
			if candidate == 0 {
				break
			}
			var pid uint32
			windowPID.Call(candidate, uintptr(unsafe.Pointer(&pid)))
			if pid == uint32(os.Getpid()) {
				hwnd = candidate
				break
			}
			after = candidate
		}
		return hwnd != 0 && observed.contains("CLI fixture")
	})
	// Exercise stale-window rejection using only this test's own native window.
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, uint32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	owner := &Owner{hwnd: hwnd, pid: uint32(os.Getpid()), process: handle, property: utf16("ASD.Test.Owner." + id)}
	setProp.Call(hwnd, uintptr(unsafe.Pointer(owner.property)), 1)
	if !owner.current() {
		owner.Close()
		t.Fatal("original window binding rejected")
	}
	removeProp.Call(hwnd, uintptr(unsafe.Pointer(owner.property)))
	before, _, _ := getForeground.Call()
	if owner.Activate() {
		owner.Close()
		t.Fatal("stale window binding activated")
	}
	after, _, _ := getForeground.Call()
	owner.Close()
	if before != after {
		t.Fatal("rejected window binding changed foreground")
	}
	title := func() string {
		var text [256]uint16
		getText.Call(hwnd, uintptr(unsafe.Pointer(&text[0])), 256)
		return windows.UTF16ToString(text[:])
	}
	wait("ASD idle", func() bool { return title() == "ASD Blocking Reminders · idle · 0" })
	program.Send(tea.KeyPressMsg{Code: 'd', Text: "DRAFT_A"})
	wait("unsubmitted A draft", func() bool {
		preview, err := b.Preview(ctx, a, 60)
		return err == nil && strings.Contains(preview.Text, "DRAFT_A")
	})
	target, bLog := start("Decision", true)
	globals := []string{"--terminal-bundle", bundle, "--runtime-dir", opts.RuntimeDir, "--namespace", opts.Namespace, "--state-dir", filepath.Join(root, "state")}
	credential := filepath.Join(root, "form.credential.json")
	cmd := exec.CommandContext(ctx, product, append(append([]string(nil), globals...), "agent-bind", "--session", target.ID, "--run", target.RunID, "--token-file", credential)...)
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture source binding: %v %s", err, data)
	}
	attentionID, _ := session.NewID()
	data, _ := json.Marshal(struct {
		CLI                         string
		Globals                     []string
		CredentialFile, AttentionID string
	}{product, globals, credential, attentionID})
	if err := os.WriteFile(config, data, 0600); err != nil {
		t.Fatal(err)
	}
	wait("existing event reaches native pet", func() bool { return title() == "ASD Blocking Reminders · approval · 1" })
	if !observed.contains("DRAFT_A") {
		t.Fatal("passive reminder changed active terminal")
	}
	// A native click only requests navigation; no decision keys are sent.
	send.Call(hwnd, wmLeftDown, 1, 0)
	send.Call(hwnd, wmLeftUp, 0, 0)
	wait("exact original form viewed", func() bool {
		item, err := db.Attention(ctx, attentionID)
		return err == nil && item.ViewedAt != nil && observed.contains("[y/N]")
	})
	item, err := db.Attention(ctx, attentionID)
	if err != nil || item.State != "pending" || title() != "ASD Blocking Reminders · approval · 1" {
		t.Fatal("viewing cleared pending", err)
	}
	formLog, _ := os.ReadFile(bLog)
	if bytes.Contains(formLog, []byte(`"event":"input"`)) {
		t.Fatal("pet answered the form")
	}
	send.Call(hwnd, wmCommand, 5, 0) // Hide only the pet; monitoring continues.
	visible := user.NewProc("IsWindowVisible")
	if shown, _, _ := visible.Call(hwnd); shown != 0 {
		t.Fatal("pet hide failed")
	}
	// Only this explicit terminal input resolves the original fixture request.
	program.Send(tea.KeyPressMsg{Code: 'y', Text: "y"})
	program.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	wait("source resolves and native pet clears", func() bool {
		item, err := db.Attention(ctx, attentionID)
		return err == nil && item.State == "resolved" && title() == "ASD Blocking Reminders · idle · 0"
	})
	formLog, _ = os.ReadFile(bLog)
	if bytes.Count(formLog, []byte(`"event":"input"`)) != 1 || !bytes.Contains(formLog, []byte(`"text":"y"`)) {
		t.Fatal("decision input was not exactly y once")
	}
	send.Call(hwnd, wmCommand, 5, 0)
	if shown, _, _ := visible.Call(hwnd); shown == 0 {
		t.Fatal("pet restore failed")
	}
	aData, _ := os.ReadFile(aLog)
	if bytes.Contains(aData, []byte(`"event":"input"`)) {
		t.Fatal("A draft submitted")
	}
	view, err := service.Workspace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, original := range originals {
		found := false
		for _, live := range view.Sessions {
			if live.ID == original.ID && live.RunID == original.RunID && live.BackendID == original.BackendID {
				found = true
			}
		}
		if !found {
			t.Fatal("original Run replaced")
		}
	}
	preview, err := b.Preview(ctx, a, 60)
	if err != nil || !strings.Contains(preview.Text, "DRAFT_A") {
		t.Fatal("A draft lost", err)
	}
	t.Log(fmt.Sprintf("existing pending -> real native pet click -> original form/viewed still pending -> explicit y/Enter -> source resolved/idle; original Run %s/%s preserved; no desktop focus/clipboard APIs", a.RunID, target.RunID))
}
