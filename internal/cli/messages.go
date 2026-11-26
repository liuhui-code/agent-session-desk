package cli

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func messagingCommand(ctx context.Context, command string, args []string, opts backend.Options, stateDir string, input io.Reader, output, errorOutput io.Writer) error {
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	f.SetOutput(errorOutput)
	var request session.MessageRequest
	var file string
	var deliver bool
	lines := 60
	if command == "read" {
		f.IntVar(&lines, "lines", 60, "screen snapshot lines (1-200)")
	}
	if command == "send" {
		f.BoolVar(&deliver, "deliver", false, "explicitly paste through the backend; no Enter is appended")
		f.StringVar(&request.RequestID, "request-id", "", "stable 32-hex idempotency key")
		f.StringVar(&request.TargetSessionID, "to", "", "exact target Session ID")
		f.StringVar(&request.TargetRunID, "run", "", "exact target Run ID")
		f.StringVar(&file, "body-file", "", "UTF-8 text file, or - for stdin")
		f.StringVar(&request.SourceRunID, "source-run", "", "source snapshot Run ID")
		f.StringVar(&request.SourceRecordID, "record", "", "retained source record ID")
		f.IntVar(&request.StartLine, "start-line", 0, "first snapshot-local line (1-based)")
		f.IntVar(&request.EndLine, "end-line", 0, "last snapshot-local line (inclusive)")
	}
	if err := f.Parse(args); err != nil {
		return err
	}
	id := ""
	if command == "send" {
		if f.NArg() != 0 || (file == "") == (request.SourceRecordID == "") {
			return fmt.Errorf("send requires --body-file or --source-run with --record, and no positional arguments")
		}
		if !session.ValidID(request.RequestID) || !session.ValidID(request.TargetSessionID) || !session.ValidID(request.TargetRunID) {
			return fmt.Errorf("send requires full --request-id, --to and --run IDs")
		}
		if file != "" {
			reader := input
			if file != "-" {
				r, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("cannot read message file")
				}
				defer r.Close()
				reader = r
			}
			if reader == nil {
				return fmt.Errorf("message stdin unavailable")
			}
			data, err := io.ReadAll(io.LimitReader(reader, session.MaxMessageBytes+4))
			if err != nil {
				return fmt.Errorf("cannot read message body")
			}
			request.Body = string(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}))
		}
		if _, err := request.Fingerprint(); err != nil {
			return err
		}
	} else {
		if f.NArg() != 1 || !session.ValidID(f.Arg(0)) || lines < 1 || lines > 200 {
			return fmt.Errorf("%s requires one full ID; read --lines must be 1-200", command)
		}
		id = f.Arg(0)
	}
	// status is deliberately offline: no backend probe or input transport.
	if command == "status" {
		db, err := openMetadata(ctx, opts, stateDir)
		if err != nil {
			return err
		}
		defer db.Close()
		msg, err := db.Message(ctx, id)
		if err != nil {
			return err
		}
		return writeJSON(output, msg)
	}
	terminal, err := backend.New(opts)
	if err != nil {
		return err
	}
	if _, err = terminal.Probe(ctx); err != nil {
		return err
	}
	db, err := openMetadata(ctx, opts, stateDir)
	if err != nil {
		return err
	}
	defer db.Close()
	service := manager.New(terminal, db)
	if command == "send" {
		msg, e := service.QueueMessage(ctx, request)
		if e != nil {
			return e
		}
		if !deliver {
			return writeJSON(output, msg)
		}
		if msg.State != "queued" {
			if err := writeJSON(output, msg); err != nil {
				return err
			}
			if msg.State == "unknown" {
				return fmt.Errorf("previous delivery result unknown; no replay attempted")
			}
			return nil
		}
		result, e := service.DeliverQueued(ctx, msg.ID)
		if result.ID != "" {
			if err := writeJSON(output, result); err != nil {
				return err
			}
		}
		return e
	}
	if command == "deliver" {
		result, e := service.DeliverQueued(ctx, id)
		if result.ID != "" {
			if err := writeJSON(output, result); err != nil {
				return err
			}
		}
		return e
	}
	rows, err := service.List(ctx)
	if err != nil {
		return err
	}
	target, err := findSession(rows, id)
	if err != nil {
		return err
	}
	record, err := service.Read(ctx, target, lines)
	if err != nil {
		return err
	}
	return writeJSON(output, struct {
		session.OutputRecord
		StartLine int `json:"start_line"`
		EndLine   int `json:"end_line"`
	}{record, 1, bytes.Count([]byte(record.Text), []byte{'\n'}) + 1})
}
