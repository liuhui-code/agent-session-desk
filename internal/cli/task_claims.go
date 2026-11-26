package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func taskClaimingCommand(ctx context.Context, command string, args []string, opts backend.Options, stateDir string, output, errorOutput io.Writer) error {
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	f.SetOutput(errorOutput)
	var tokenFile, claimFile, id, msg string
	seconds := session.DefaultClaimSeconds
	if command != "claim-status" {
		f.StringVar(&tokenFile, "token-file", "", "Run-bound agent credential file")
		f.StringVar(&claimFile, "claim-file", "", "private processing claim file")
	}
	if command == "claim" {
		f.StringVar(&id, "claim-id", "", "stable processing claim ID")
		f.StringVar(&msg, "message-id", "", "original request ID")
		f.IntVar(&seconds, "lease-seconds", session.DefaultClaimSeconds, "processing lease duration (1-3600)")
	}
	if err := f.Parse(args); err != nil {
		return err
	}
	if command == "claim-status" {
		if f.NArg() != 1 || !session.ValidID(f.Arg(0)) {
			return fmt.Errorf("claim-status requires one full message ID")
		}
		db, err := openMetadata(ctx, opts, stateDir)
		if err != nil {
			return err
		}
		defer db.Close()
		c, err := db.TaskClaimStatus(ctx, f.Arg(0))
		if err != nil {
			return err
		}
		return writeJSON(output, struct {
			MessageID string             `json:"message_id"`
			Claim     *session.TaskClaim `json:"claim"`
		}{f.Arg(0), c})
	}
	if tokenFile == "" || claimFile == "" || f.NArg() != 0 {
		return fmt.Errorf("%s requires --token-file, --claim-file and no positional arguments", command)
	}
	if command == "claim" && (!session.ValidID(id) || !session.ValidID(msg) || seconds < 1 || seconds > session.MaxClaimSeconds) {
		return fmt.Errorf("claim requires full --claim-id, --message-id and lease-seconds 1-3600")
	}
	agent, err := loadAgentCredential(tokenFile)
	if err != nil {
		return err
	}
	var key session.TaskClaimCredential
	if command == "renew" {
		key, _, err = loadTaskClaimFile(claimFile)
		if err != nil {
			return err
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
	var result *session.TaskClaim
	if command == "claim" {
		key, err = ensureTaskClaimFile(claimFile, id, msg, seconds)
		if err != nil {
			return err
		}
		result, err = m.ClaimTask(ctx, agent, key, seconds)
	} else {
		result, err = m.RenewTaskClaim(ctx, agent, key)
	}
	if result != nil {
		if e := writeJSON(output, result); e != nil {
			return e
		}
	}
	return err
}
