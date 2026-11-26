package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func agentMessagingCommand(ctx context.Context, command string, args []string, opts backend.Options, stateDir string, input io.Reader, output, errorOutput io.Writer) error {
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	f.SetOutput(errorOutput)
	var tokenFile, sid, rid, bodyFile, claimFile string
	var request session.ReplyRequest
	if command != "replies" {
		f.StringVar(&tokenFile, "token-file", "", "private Run-bound credential file")
	}
	if command == "agent-bind" {
		f.StringVar(&sid, "session", "", "exact Session ID")
		f.StringVar(&rid, "run", "", "exact Run ID")
	}
	if command == "reply" {
		f.StringVar(&claimFile, "claim-file", "", "private processing claim credential")
		f.StringVar(&request.ID, "reply-id", "", "stable reply idempotency key")
		f.StringVar(&request.ReplyTo, "reply-to", "", "original message ID")
		f.StringVar(&bodyFile, "body-file", "", "UTF-8 result file, or - for stdin")
	}
	if err := f.Parse(args); err != nil {
		return err
	}
	if command == "replies" {
		if f.NArg() != 1 || !session.ValidID(f.Arg(0)) {
			return fmt.Errorf("replies requires one full message ID")
		}
		db, err := openMetadata(ctx, opts, stateDir)
		if err != nil {
			return err
		}
		defer db.Close()
		rows, err := db.Replies(ctx, f.Arg(0))
		if err != nil {
			return err
		}
		return writeJSON(output, rows)
	}
	if tokenFile == "" || f.NArg() != 0 {
		return fmt.Errorf("%s requires --token-file and no positional arguments", command)
	}
	var credential session.AgentCredential
	var err error
	if command == "agent-bind" {
		if !session.ValidID(sid) || !session.ValidID(rid) {
			return fmt.Errorf("agent-bind requires full --session and --run IDs")
		}
		tokenFile, err = filepath.Abs(tokenFile)
	} else {
		credential, err = loadAgentCredential(tokenFile)
	}
	if err != nil {
		return err
	}
	if command == "reply" {
		if !session.ValidID(request.ID) || !session.ValidID(request.ReplyTo) || bodyFile == "" || claimFile == "" {
			return fmt.Errorf("reply requires --reply-id, --reply-to, --claim-file and --body-file")
		}
		request.Body, err = readMessageBody(bodyFile, input)
		if err != nil {
			return err
		}
		request.Claim, _, err = loadTaskClaimFile(claimFile)
		if err != nil {
			return err
		}
		if request.Claim.MessageID != request.ReplyTo {
			return fmt.Errorf("claim file does not match reply-to")
		}
	}
	b, err := backend.New(opts)
	if err != nil {
		return err
	}
	if _, err = b.Probe(ctx); err != nil {
		return err
	}
	db, err := openMetadata(ctx, opts, stateDir)
	if err != nil {
		return err
	}
	defer db.Close()
	m := manager.New(b, db)
	switch command {
	case "agent-bind":
		credential, err = ensureAgentCredential(tokenFile, sid, rid)
		if err != nil {
			return err
		}
		if _, err = m.BindAgent(ctx, credential); err != nil {
			return fmt.Errorf("binding unconfirmed; preserve credential file and reuse it on retry: %w", err)
		}
		return writeJSON(output, struct {
			SessionID      string `json:"session_id"`
			RunID          string `json:"run_id"`
			CredentialFile string `json:"credential_file"`
			Result         string `json:"result"`
		}{sid, rid, tokenFile, "bound"})
	case "inbox":
		rows, e := m.Inbox(ctx, credential)
		if e != nil {
			return e
		}
		return writeJSON(output, rows)
	case "reply":
		r, e := m.Reply(ctx, credential, request)
		if e != nil {
			return e
		}
		return writeJSON(output, r)
	}
	return fmt.Errorf("unknown agent messaging command")
}
