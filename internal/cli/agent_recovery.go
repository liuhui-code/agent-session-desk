package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/recovery"
)

func readRecoveryFile(path string, limit int) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() {
		return nil, fmt.Errorf("recovery file must be an existing regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("recovery file unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !os.SameFile(before, info) {
		return nil, fmt.Errorf("recovery file identity changed")
	}
	b, err := io.ReadAll(io.LimitReader(f, int64(limit+1)))
	if err != nil || len(b) > limit {
		return nil, fmt.Errorf("recovery file exceeds byte limit or cannot be read")
	}
	return b, nil
}

func agentRecoveryCommand(ctx context.Context, command string, args []string, opts backend.Options, stateDir string, output, diag io.Writer) error {
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	f.SetOutput(diag)
	var sid, previous, tokenFile, document, contextDir string
	var revision int64
	var confirmed bool
	switch command {
	case "recovery-profile", "recovery-status":
		f.StringVar(&sid, "session", "", "full logical Session ID")
	case "recovery-register":
		f.StringVar(&tokenFile, "token-file", "", "original Run credential file")
		f.StringVar(&document, "file", "", "UTF-8 recovery declaration")
		f.Int64Var(&revision, "revision", -1, "expected profile revision (0 for first registration)")
	case "recovery-result":
		f.StringVar(&tokenFile, "token-file", "", "new Run credential file")
		f.StringVar(&document, "file", "", "UTF-8 recovery result")
	case "recover":
		f.StringVar(&sid, "session", "", "original logical Session ID")
		f.StringVar(&previous, "previous-run", "", "latest recorded Run explicitly confirmed ended")
		f.Int64Var(&revision, "revision", 0, "expected profile revision")
		f.BoolVar(&confirmed, "confirm-ended", false, "user confirms previous Run ended; never infer this from absence")
		f.StringVar(&contextDir, "context-dir", "", "new private directory for this attempt")
	}
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected recovery arguments")
	}
	if command == "recover" || command == "recovery-profile" || command == "recovery-status" {
		if !validID(sid) {
			return fmt.Errorf("full --session ID required")
		}
	}
	if command == "recover" {
		if !validID(previous) || revision < 1 || !confirmed || contextDir == "" {
			return fmt.Errorf("recover requires --session, --previous-run, --revision, --confirm-ended and a new --context-dir")
		}
		if runtime.GOOS != "windows" {
			return fmt.Errorf("recovery v1 execution requires native Windows")
		}
	}
	if command == "recovery-register" && (revision < 0 || tokenFile == "" || document == "") {
		return fmt.Errorf("recovery-register requires --token-file, --file and --revision")
	}
	if command == "recovery-result" && (tokenFile == "" || document == "") {
		return fmt.Errorf("recovery-result requires --token-file and --file")
	}
	if command == "recovery-profile" || command == "recovery-status" {
		db, err := openMetadata(ctx, opts, stateDir)
		if err != nil {
			return err
		}
		defer db.Close()
		if command == "recovery-profile" {
			p, e := db.RecoveryProfile(ctx, sid)
			if e != nil {
				return e
			}
			return writeJSON(output, p)
		}
		a, e := db.RecoveryAttempt(ctx, sid)
		if e != nil {
			return e
		}
		return writeJSON(output, a)
	}
	var declaration recovery.Declaration
	var result recovery.Result
	credential, err := loadAgentCredentialIfProvided(tokenFile)
	if err != nil {
		return err
	}
	if document != "" {
		limit := recovery.MaxDeclarationBytes
		if command == "recovery-result" {
			limit = recovery.MaxResultBytes
		}
		b, e := readRecoveryFile(document, limit)
		if e != nil {
			return e
		}
		if command == "recovery-register" {
			declaration, err = recovery.DecodeDeclaration(b)
		} else {
			result, err = recovery.DecodeResult(b)
		}
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
	switch command {
	case "recovery-register":
		p, e := m.RegisterRecovery(ctx, credential, declaration, revision)
		if e != nil {
			return e
		}
		return writeJSON(output, p)
	case "recovery-result":
		a, e := m.ReportRecovery(ctx, credential, result)
		if e != nil {
			return e
		}
		return writeJSON(output, a)
	case "recover":
		p, e := db.RecoveryProfile(ctx, sid)
		if e != nil {
			return e
		}
		if p.Revision != revision {
			return fmt.Errorf("recovery profile revision conflict")
		}
		request, e := createRecoveryContext(opts, stateDir, contextDir, p, previous)
		if e != nil {
			return e
		}
		a, e := m.Recover(ctx, request)
		if e != nil {
			if a.ID != "" {
				_ = writeJSON(output, a)
			}
			return e
		}
		return writeJSON(output, a)
	}
	return fmt.Errorf("unknown recovery command")
}
