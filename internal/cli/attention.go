package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func readAttentionEvent(path string, input io.Reader) (session.AttentionEvent, error) {
	var e session.AttentionEvent
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return e, err
		}
		defer f.Close()
		input = f
	}
	data, err := io.ReadAll(io.LimitReader(input, session.MaxAttentionEventBytes+1))
	if err != nil || len(data) > session.MaxAttentionEventBytes {
		return e, fmt.Errorf("attention event exceeds 16 KiB or cannot be read")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err = d.Decode(&e); err != nil {
		return e, err
	}
	if d.Decode(new(any)) != io.EOF {
		return e, fmt.Errorf("one attention JSON document required")
	}
	_, err = e.Fingerprint()
	return e, err
}

func attentionCommand(ctx context.Context, command string, args []string, opts backend.Options, stateDir string, input io.Reader, output, errorOutput io.Writer) error {
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	f.SetOutput(errorOutput)
	var tokenFile, eventFile string
	if command == "attention-event" {
		f.StringVar(&tokenFile, "token-file", "", "private Run-bound credential")
		f.StringVar(&eventFile, "event-file", "", "UTF-8 JSON file, or - for stdin")
	}
	if err := f.Parse(args); err != nil {
		return err
	}
	if command == "attention-status" && (f.NArg() != 1 || !session.ValidID(f.Arg(0))) {
		return fmt.Errorf("attention-status requires one full attention ID")
	}
	if command != "attention-status" && f.NArg() != 0 {
		return fmt.Errorf("unexpected attention arguments")
	}
	var event session.AttentionEvent
	var credential session.AgentCredential
	var err error
	var b *backend.Backend
	if command == "attention-event" {
		if tokenFile == "" || eventFile == "" {
			return fmt.Errorf("attention-event requires --token-file and --event-file")
		}
		credential, err = loadAgentCredential(tokenFile)
		if err != nil {
			return err
		}
		event, err = readAttentionEvent(eventFile, input)
		if err != nil {
			return err
		}
		b, err = backend.New(opts)
		if err != nil {
			return err
		}
		if _, err = b.Probe(ctx); err != nil {
			return err
		}
	}
	db, err := openMetadata(ctx, opts, stateDir)
	if err != nil {
		return err
	}
	defer db.Close()
	switch command {
	case "attention-event":
		item, err := manager.New(b, db).ReportAttention(ctx, credential, event)
		if err != nil {
			return err
		}
		return writeJSON(output, item)
	case "attention-status":
		item, err := db.Attention(ctx, f.Arg(0))
		if err != nil {
			return err
		}
		return writeJSON(output, item)
	case "attention-list":
		items, err := db.PendingAttention(ctx)
		if err != nil {
			return err
		}
		return writeJSON(output, items)
	}
	return fmt.Errorf("unknown attention command")
}
