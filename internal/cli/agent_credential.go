package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type credentialDocument struct {
	Version   int    `json:"schema_version"`
	SessionID string `json:"session_id"`
	RunID     string `json:"run_id"`
	Token     string `json:"token"`
}

func loadAgentCredential(path string) (session.AgentCredential, error) {
	var zero session.AgentCredential
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() {
		return zero, fmt.Errorf("credential file must be an existing regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return zero, fmt.Errorf("credential file unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !os.SameFile(before, info) {
		return zero, fmt.Errorf("credential file identity changed")
	}
	data, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil || len(data) > 4096 {
		return zero, fmt.Errorf("invalid credential file")
	}
	var doc credentialDocument
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&doc) != nil || decoder.Decode(new(any)) != io.EOF || doc.Version != 1 {
		return zero, fmt.Errorf("invalid credential file")
	}
	c := session.AgentCredential{SessionID: doc.SessionID, RunID: doc.RunID, Secret: doc.Token}
	return c, c.Validate()
}

// Publication is non-overwriting. On an uncertain later DB commit, the file
// remains available so a retry reuses the same capability instead of rotating it.
func ensureAgentCredential(path, sid, rid string) (session.AgentCredential, error) {
	if _, err := os.Lstat(path); err == nil {
		return matchingAgentCredential(path, sid, rid)
	} else if !os.IsNotExist(err) {
		return session.AgentCredential{}, fmt.Errorf("credential destination unavailable")
	}
	c, err := session.NewAgentCredential(sid, rid)
	if err != nil {
		return c, err
	}
	data, err := json.Marshal(credentialDocument{1, c.SessionID, c.RunID, c.Secret})
	if err != nil {
		return session.AgentCredential{}, fmt.Errorf("credential encoding failed")
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".asd-agent-credential-*")
	if err != nil {
		return session.AgentCredential{}, fmt.Errorf("credential staging failed; destination directory must exist")
	}
	stage := f.Name()
	owned, _ := f.Stat()
	defer func() {
		if now, e := os.Lstat(stage); e == nil && owned != nil && os.SameFile(owned, now) {
			_ = os.Remove(stage)
		}
	}()
	n, err := f.Write(append(data, '\n'))
	if err == nil && n != len(data)+1 {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return session.AgentCredential{}, fmt.Errorf("credential write failed")
	}
	if err = os.Link(stage, path); os.IsExist(err) {
		return matchingAgentCredential(path, sid, rid)
	} else if err != nil {
		return session.AgentCredential{}, fmt.Errorf("credential publication requires hard-link support and absent destination")
	}
	if err = os.Remove(stage); err != nil {
		return session.AgentCredential{}, fmt.Errorf("credential published; staging cleanup failed; preserve destination")
	}
	return c, nil
}

func matchingAgentCredential(path, sid, rid string) (session.AgentCredential, error) {
	c, err := loadAgentCredential(path)
	if err == nil && (c.SessionID != sid || c.RunID != rid) {
		return session.AgentCredential{}, fmt.Errorf("existing credential file belongs to another Session/Run")
	}
	return c, err
}
