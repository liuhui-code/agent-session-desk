// asd-fixture is a controlled CLI for the M0 launch/attach checks. It is not an
// agent and does not emit readiness, completion, or attention protocol events.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	env := map[string]string{}
	// Only explicitly selected synthetic keys are printed, never the full env.
	for _, key := range strings.Split(os.Getenv("ASD_FIXTURE_ENV_KEYS"), ",") {
		if key != "" {
			env[key] = os.Getenv(key)
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	_ = encoder.Encode(struct {
		Event string            `json:"event"`
		Args  []string          `json:"args"`
		Cwd   string            `json:"cwd"`
		Env   map[string]string `json:"env"`
	}{"started", os.Args[1:], cwd, env})
	fmt.Println("CLI fixture · 中文 · e\u0301 · 🧪 — type exit to end")
	input := bufio.NewScanner(os.Stdin)
	input.Buffer(make([]byte, 4096), 64*1024)
	for input.Scan() {
		line := input.Text()
		_ = encoder.Encode(struct {
			Event string `json:"event"`
			Text  string `json:"text"`
		}{"input", line})
		if line == "exit" {
			return
		}
	}
	if err := input.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "fixture input:", err)
		os.Exit(2)
	}
}
