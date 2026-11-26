package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

type credentialDocument struct {
	Version   int    `json:"schema_version"`
	SessionID string `json:"session_id"`
	RunID     string `json:"run_id"`
	Token     string `json:"token"`
}

func loadAgentCredential(path string) (session.AgentCredential, error) {
	var doc credentialDocument
	if err := readPrivateDocument(path, &doc); err != nil {
		return session.AgentCredential{}, err
	}
	if doc.Version != 1 {
		return session.AgentCredential{}, fmt.Errorf("invalid credential file")
	}
	c := session.AgentCredential{SessionID: doc.SessionID, RunID: doc.RunID, Secret: doc.Token}
	return c, c.Validate()
}

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
	created, err := publishPrivateDocument(path, data)
	if err != nil {
		return session.AgentCredential{}, err
	}
	if !created {
		return matchingAgentCredential(path, sid, rid)
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
