package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This child executable records OS-delivered argv and serves a deliberately
// narrow fake multiplexer. It does not emulate ConPTY or establish Windows
// compatibility. Every test uses an isolated state and command log.
func TestMain(m *testing.M) {
	if os.Getenv("ASD_TEST_BACKEND_HELPER") == "1" {
		os.Exit(fakeBackend(os.Args[1:]))
	}
	os.Exit(m.Run())
}

type fakeState struct {
	Name     string
	Metadata string
	Dead     string
	Active   string
	Missing  bool
}

func fakeBackend(args []string) int {
	log, err := os.OpenFile(os.Getenv("ASD_TEST_BACKEND_LOG"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return 90
	}
	_ = json.NewEncoder(log).Encode(args)
	_ = log.Close()
	if os.Getenv("ASD_TEST_BACKEND_DELAY") != "" {
		time.Sleep(3 * time.Second)
	}
	var route string
	for len(args) >= 2 && (args[0] == "-L" || args[0] == "-f" || args[0] == "-t") {
		if args[0] == "-t" {
			route = args[1]
		}
		args = args[2:]
	}
	if len(args) == 0 {
		return 91
	}
	if strings.Contains("|"+os.Getenv("ASD_TEST_BACKEND_FAIL")+"|", "|"+args[0]+"|") {
		fmt.Fprint(os.Stderr, "synthetic backend refusal")
		return 1
	}
	statePath := os.Getenv("ASD_TEST_BACKEND_STATE")
	var state fakeState
	data, _ := os.ReadFile(statePath)
	_ = json.Unmarshal(data, &state)
	flag := func(name string) string {
		for i := 0; i+1 < len(args); i++ {
			if args[i] == name {
				return args[i+1]
			}
		}
		return ""
	}
	save := func() { data, _ := json.Marshal(state); _ = os.WriteFile(statePath, data, 0600) }
	switch args[0] {
	case "-V":
		fmt.Print("tmux 3.4.0\npsmux 3.4.0 (20aa1d7 2026-10-13)\n")
	case "new-session":
		state = fakeState{Name: flag("-s"), Dead: "0"}
		save()
	case "list-panes":
		fmt.Println("%7")
	case "set-option":
		if route != state.Name {
			return 92
		}
		state.Metadata = args[len(args)-1]
		save()
	case "show-options":
		if route != state.Name {
			return 93
		}
		fmt.Println(state.Metadata)
	case "list-sessions":
		if state.Name != "" {
			fmt.Println(state.Name)
		}
	case "display-message":
		if args[len(args)-1] == "#{pane_id}" && flag("-t") == state.Name {
			if state.Active == "" {
				fmt.Println("%7")
			} else {
				fmt.Println(state.Active)
			}
			break
		}
		if state.Missing || route != "" || flag("-t") != state.Name+":.%7" {
			return 94
		}
		if args[len(args)-1] == "#{pane_id}|#{pane_dead}" {
			if raw := os.Getenv("ASD_TEST_PANE_TUPLE"); raw != "" {
				fmt.Println(raw)
			} else {
				fmt.Println("%7|" + state.Dead)
			}
		} else {
			fmt.Println(state.Dead)
		}
	case "capture-pane":
		if route != "" || flag("-t") != state.Name+":.%7" {
			return 95
		}
		if path := os.Getenv("ASD_TEST_CAPTURE_FILE"); path != "" {
			data, _ := os.ReadFile(path)
			_, _ = os.Stdout.Write(data)
		} else {
			fmt.Print("fixture output\n")
		}
		if os.Getenv("ASD_TEST_PANE_REMOVED_DURING_CAPTURE") == "1" {
			state.Missing = true
			save()
		}
	case "kill-session":
		if flag("-t") != state.Name {
			return 96
		}
		state = fakeState{}
		save()
	default:
		fmt.Fprint(os.Stderr, "unexpected command")
		return 97
	}
	return 0
}

func newFake(t *testing.T) *Backend {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("ASD_TEST_BACKEND_HELPER", "1")
	t.Setenv("ASD_TEST_BACKEND_LOG", filepath.Join(dir, "commands.jsonl"))
	t.Setenv("ASD_TEST_BACKEND_STATE", filepath.Join(dir, "state.json"))
	// Child test executables have no pending work on exit. Avoid the race
	// runtime's one-second exit grace period on every contract invocation.
	t.Setenv("GORACE", os.Getenv("GORACE")+" atexit_sleep_ms=0")
	bin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return &Backend{bin: bin, kind: "psmux", dataDir: dir, namespace: "asd-test", config: filepath.Join(dir, "test.conf")}
}
