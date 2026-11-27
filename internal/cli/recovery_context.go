package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/manager"
	"github.com/liuhui-code/agent-session-desk/internal/recovery"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func loadAgentCredentialIfProvided(path string) (session.AgentCredential, error) {
	if path == "" {
		return session.AgentCredential{}, nil
	}
	return loadAgentCredential(path)
}

// This is a private local handoff, not a second launch template. Credentials
// remain in their existing independent file; public receipts never contain them.
type recoveryContext struct {
	Version        int      `json:"schema_version"`
	SessionID      string   `json:"session_id"`
	RunID          string   `json:"run_id"`
	AttemptID      string   `json:"attempt_id"`
	ConversationID string   `json:"conversation_id"`
	CredentialFile string   `json:"credential_file"`
	ResultFile     string   `json:"result_file"`
	Executable     string   `json:"asd_executable"`
	GlobalArgs     []string `json:"global_args"`
}

func createRecoveryContext(opts backend.Options, stateDir, dir string, p recovery.Profile, previous string) (manager.RecoveryRequest, error) {
	var empty manager.RecoveryRequest
	dir, err := filepath.Abs(dir)
	if err != nil {
		return empty, err
	}
	if err = os.Mkdir(dir, 0700); err != nil {
		return empty, fmt.Errorf("context directory must be new and its parent must exist")
	}
	// Preserve a partially published context after failure. Never remove a path
	// the Agent might already hold or reuse it for a different recovery attempt.
	rid, err := session.NewID()
	if err != nil {
		return empty, err
	}
	aid, err := session.NewID()
	if err != nil {
		return empty, err
	}
	token := filepath.Join(dir, "agent-credential.json")
	c, err := ensureAgentCredential(token, p.Declaration.SessionID, rid)
	if err != nil {
		return empty, err
	}
	path, _, err := metadataPath(opts, stateDir)
	if err != nil {
		return empty, err
	}
	_, location, err := backend.Location(opts)
	if err != nil {
		return empty, err
	}
	binary := opts.Binary
	if opts.TerminalBundle != "" {
		binary = filepath.Join(opts.TerminalBundle, "psmux.exe")
	}
	if binary == "" {
		binary = "psmux"
	}
	binary, err = exec.LookPath(binary)
	if err != nil {
		return empty, fmt.Errorf("recovery backend executable unavailable")
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		return empty, err
	}
	global := []string{"--backend", binary, "--runtime-dir", filepath.Dir(location), "--namespace", filepath.Base(location), "--state-dir", filepath.Dir(path)}
	if opts.TerminalBundle != "" {
		bundle, e := filepath.Abs(opts.TerminalBundle)
		if e != nil {
			return empty, e
		}
		global = append(global, "--terminal-bundle", bundle)
	}
	executable, err := os.Executable()
	if err != nil {
		return empty, err
	}
	doc := recoveryContext{1, c.SessionID, rid, aid, p.Declaration.ConversationID, token, filepath.Join(dir, "result.json"), executable, global}
	b, err := json.Marshal(doc)
	if err != nil || len(b) > 16384 {
		return empty, fmt.Errorf("recovery context exceeds byte limit")
	}
	file := filepath.Join(dir, "context.json")
	published, err := publishPrivateDocument(file, b)
	if err != nil {
		return empty, err
	}
	if !published {
		return empty, fmt.Errorf("recovery context already exists")
	}
	return manager.RecoveryRequest{SessionID: c.SessionID, PreviousRunID: previous, Revision: p.Revision, ConfirmEnded: true, AttemptID: aid, Credential: c, ContextFile: file}, nil
}
