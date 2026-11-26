package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

// This optional synthetic form stays inside the original CLI. The reminder
// producer only reports lifecycle events; it has no automatic decision path.
func runAttentionFixture(input io.Reader, sink *fixtureSink, path string) (io.Reader, error) {
	var config struct {
		CLI                         string
		Globals                     []string
		CredentialFile, AttentionID string
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	for {
		data, err := os.ReadFile(path)
		if err == nil && len(data) <= session.MaxAttentionEventBytes && json.Unmarshal(data, &config) == nil {
			break
		}
		select {
		case <-ctx.Done():
			return input, fmt.Errorf("attention fixture config deadline")
		case <-time.After(25 * time.Millisecond):
		}
	}
	e := session.AttentionEvent{ID: config.AttentionID, Version: 1, Kind: "approval", State: "pending", Title: "Allow the simulated operation?", Detail: "Choose y or n in the original terminal. This is a controlled fixture."}
	report := func() error {
		data, err := json.Marshal(e)
		if err != nil {
			return err
		}
		call, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		args := append(append([]string(nil), config.Globals...), "attention-event", "--token-file", config.CredentialFile, "--event-file", "-")
		cmd := exec.CommandContext(call, config.CLI, args...)
		cmd.Stdin = bytes.NewReader(data)
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("attention fixture event failed: %s", &stderr)
		}
		var item session.AttentionItem
		if json.Unmarshal(out.Bytes(), &item) != nil || item.ID != e.ID || item.Version != e.Version || item.State != e.State {
			return fmt.Errorf("attention fixture receipt mismatch")
		}
		return nil
	}
	if _, err := fmt.Fprintln(sink, "Allow the simulated operation? [y/N]"); err != nil {
		return input, err
	}
	if err := report(); err != nil {
		return input, err
	}
	if err := sink.emit(struct {
		Event string `json:"event"`
	}{"decision_waiting"}); err != nil {
		return input, err
	}
	reader := bufio.NewReaderSize(input, 64*1024)
	for {
		line, err := reader.ReadSlice('\n')
		if err != nil {
			return reader, err
		}
		answer := strings.TrimSuffix(strings.TrimSuffix(string(line), "\n"), "\r")
		if err = sink.emit(struct {
			Event string `json:"event"`
			Text  string `json:"text"`
		}{"input", answer}); err != nil {
			return reader, err
		}
		if answer != "y" && answer != "n" && answer != "" {
			continue
		}
		e.Version = 2
		e.State = "resolved"
		if err = report(); err != nil {
			return reader, err
		}
		if err = sink.emit(struct {
			Event  string `json:"event"`
			Answer string `json:"answer"`
		}{"decision_resolved", answer}); err != nil {
			return reader, err
		}
		return reader, nil
	}
}
