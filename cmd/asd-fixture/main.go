// asd-fixture is a controlled CLI for the M0 launch/attach checks. It is not an
// agent and does not emit readiness, completion, or attention protocol events.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, os.Args[1:], os.Getenv))
}

// The optional JSONL report is an independent observation channel for native
// integration tests: terminal wrapping must not become the argv test oracle.
// It is enabled only by an explicit test-owned path and never replaces a file.
func run(stdin io.Reader, stdout, stderr io.Writer, args []string, getenv func(string) string) int {
	var report *os.File
	if path := getenv("ASD_FIXTURE_REPORT_PATH"); path != "" {
		var err error
		report, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			fmt.Fprintln(stderr, "fixture report:", err)
			return 2
		}
		defer report.Close()
	}
	exitCode := 0
	if value := getenv("ASD_FIXTURE_EXIT_CODE"); value != "" {
		var err error
		exitCode, err = strconv.Atoi(value)
		if err != nil || exitCode < 0 || exitCode > 255 {
			fmt.Fprintln(stderr, "ASD_FIXTURE_EXIT_CODE must be an integer from 0 to 255")
			return 2
		}
	}
	emit := func(value any) error {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		data = append(data, '\n')
		if report != nil {
			// No buffered writer: every complete record is visible to the test
			// while the process is still blocked waiting for terminal input.
			if _, err := report.Write(data); err != nil {
				return err
			}
		}
		_, err = stdout.Write(data)
		return err
	}
	fail := func(err error) int {
		fmt.Fprintln(stderr, "fixture:", err)
		return 2
	}
	finish := func(code int) int {
		if err := emit(struct {
			Event string `json:"event"`
			Code  int    `json:"code"`
		}{"exiting", code}); err != nil {
			return fail(err)
		}
		return code
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	env := map[string]string{}
	// Only explicitly selected synthetic keys are printed, never the full env.
	for _, key := range strings.Split(getenv("ASD_FIXTURE_ENV_KEYS"), ",") {
		if key != "" {
			env[key] = getenv(key)
		}
	}
	if err := emit(struct {
		Event string            `json:"event"`
		PID   int               `json:"pid"`
		Args  []string          `json:"args"`
		Cwd   string            `json:"cwd"`
		Env   map[string]string `json:"env"`
	}{"started", os.Getpid(), args, cwd, env}); err != nil {
		return fail(err)
	}
	if getenv("ASD_FIXTURE_EXIT_AFTER_START") == "1" {
		return finish(exitCode)
	}
	fmt.Fprintln(stdout, "CLI fixture · 中文 · e\u0301 · 🧪 — type exit to end")
	input := bufio.NewScanner(stdin)
	input.Buffer(make([]byte, 4096), 64*1024)
	for input.Scan() {
		line := input.Text()
		if err := emit(struct {
			Event string `json:"event"`
			Text  string `json:"text"`
		}{"input", line}); err != nil {
			return fail(err)
		}
		if line == "exit" {
			return finish(exitCode)
		}
	}
	if err := input.Err(); err != nil {
		fmt.Fprintln(stderr, "fixture input:", err)
		return finish(2)
	}
	return finish(exitCode)
}
